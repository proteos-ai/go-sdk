package sdk

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.proteos.ai/sdk/internal/httpx"
)

type echoResponse struct {
	OK   bool   `json:"ok"`
	Path string `json:"path"`
}

func newServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL), WithToken("test-token"))
	require.NoError(t, err)
	return srv, c
}

func TestNewClient_RequiresBaseURL(t *testing.T) {
	_, err := NewClient()
	require.Error(t, err)
	require.Contains(t, err.Error(), "WithBaseURL is required")
}

func TestNewClient_StripsTrailingSlash(t *testing.T) {
	c, err := NewClient(WithBaseURL("https://api.example.com/"))
	require.NoError(t, err)
	require.Equal(t, "https://api.example.com", c.BaseURL())
}

func TestRequest_AddsBearerToken(t *testing.T) {
	var got string
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, c.Do(context.Background(), http.MethodGet, "/x", nil, nil))
	require.Equal(t, "Bearer test-token", got)
}

func TestRequest_AddsAcceptAndContentType(t *testing.T) {
	var accept, contentType string
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		accept = r.Header.Get("Accept")
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, c.Do(context.Background(), http.MethodPost, "/x", map[string]string{"k": "v"}, nil))
	require.Equal(t, "application/json", accept)
	require.Equal(t, "application/json", contentType)
}

func TestRequest_NoContentTypeWhenNoBody(t *testing.T) {
	var contentType string
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, c.Do(context.Background(), http.MethodGet, "/x", nil, nil))
	require.Empty(t, contentType)
}

func TestRequest_AppendsQueryParams(t *testing.T) {
	var url string
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		url = r.URL.String()
		_, _ = w.Write([]byte(`{"ok":true,"path":"` + r.URL.Path + `"}`))
	})
	type query struct {
		Page     int    `query:"page"`
		PageSize int    `query:"page_size"`
		Sort     string `query:"sort,omitempty"`
	}
	var resp echoResponse
	require.NoError(t, c.DoWithQuery(context.Background(), http.MethodGet, "/items", query{Page: 2, PageSize: 50, Sort: "name"}, nil, &resp))
	require.Contains(t, url, "page=2")
	require.Contains(t, url, "page_size=50")
	require.Contains(t, url, "sort=name")
	require.True(t, resp.OK)
	require.Equal(t, "/items", resp.Path)
}

func TestRequest_PostsJSONBody(t *testing.T) {
	var body []byte
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	in := map[string]any{"a": 1, "b": "two"}
	var out map[string]any
	require.NoError(t, c.Do(context.Background(), http.MethodPost, "/x", in, &out))
	require.Contains(t, string(body), `"a":1`)
	require.Contains(t, string(body), `"b":"two"`)
}

func TestRequest_204NoContentReturnsNil(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	var out map[string]any
	require.NoError(t, c.Do(context.Background(), http.MethodDelete, "/x", nil, &out))
	require.Nil(t, out)
}

func TestRequest_ParsesJSONErrorResponse(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"not_found","message":"missing"}`))
	})
	err := c.Do(context.Background(), http.MethodGet, "/x", nil, nil)
	require.Error(t, err)
	var sdkErr *Error
	require.True(t, errors.As(err, &sdkErr))
	require.Equal(t, 404, sdkErr.HTTPStatus)
	require.Equal(t, ErrCodeNotFound, sdkErr.Code)
	require.Equal(t, "missing", sdkErr.Message)
	require.True(t, IsNotFound(err))
}

func TestRequest_ParsesPlainTextErrorResponse(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	})
	err := c.Do(context.Background(), http.MethodGet, "/x", nil, nil)
	require.Error(t, err)
	var sdkErr *Error
	require.True(t, errors.As(err, &sdkErr))
	require.Equal(t, 500, sdkErr.HTTPStatus)
	require.Equal(t, ErrCodeInternalServerError, sdkErr.Code)
	require.Equal(t, "boom", sdkErr.Message)
}

func TestRequest_HonorsContextCancellation(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	err := c.Do(ctx, http.MethodGet, "/x", nil, nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, context.Canceled), "expected wrapped context.Canceled, got %v", err)
}

func TestRequest_HonorsTimeout(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	})
	c.cfg.timeout = 30 * time.Millisecond
	err := c.Do(context.Background(), http.MethodGet, "/x", nil, nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, context.DeadlineExceeded), "expected DeadlineExceeded, got %v", err)
}

func TestRequest_PerRequestTimeoutOverridesClientTimeout(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(500 * time.Millisecond):
		}
	})
	c.cfg.timeout = 10 * time.Second
	err := c.Do(context.Background(), http.MethodGet, "/x", nil, nil, httpx.WithRequestTimeout(20*time.Millisecond))
	require.Error(t, err)
	require.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestRequest_TokenProvider_CalledPerRequest(t *testing.T) {
	var calls int
	// TokenProvider contract: returns the full Authorization header
	// verbatim. The SDK emits it as-is; no normalization on this path.
	tokens := []string{"Bearer t1", "Bearer t2"}
	provider := func(ctx context.Context) (string, error) {
		tok := tokens[calls]
		calls++
		return tok, nil
	}
	c, err := NewClient(WithBaseURL("http://placeholder"), WithTokenProvider(provider))
	require.NoError(t, err)

	got := []string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	c.cfg.baseURL = srv.URL

	require.NoError(t, c.Do(context.Background(), http.MethodGet, "/", nil, nil))
	require.NoError(t, c.Do(context.Background(), http.MethodGet, "/", nil, nil))
	require.Equal(t, []string{"Bearer t1", "Bearer t2"}, got)
}

func TestWithToken_NormalizesRawTokenToBearerHeader(t *testing.T) {
	// WithToken is the construction-time convenience: raw JWT in,
	// full header stored internally so the rest of the codebase sees
	// the canonical "full header" format.
	c, err := NewClient(WithBaseURL("http://placeholder"), WithToken("eyJraw"))
	require.NoError(t, err)
	require.Equal(t, "Bearer eyJraw", c.cfg.token)
}

func TestWithToken_FullHeaderInputIsIdempotent(t *testing.T) {
	// Passing the already-prefixed form must not produce
	// "Bearer Bearer eyJ..." — the normalization is case-insensitive
	// and only adds the prefix when missing.
	c, err := NewClient(WithBaseURL("http://placeholder"), WithToken("Bearer eyJ"))
	require.NoError(t, err)
	require.Equal(t, "Bearer eyJ", c.cfg.token)

	c2, err := NewClient(WithBaseURL("http://placeholder"), WithToken("bearer eyJ"))
	require.NoError(t, err)
	require.Equal(t, "bearer eyJ", c2.cfg.token, "preserves the user-supplied casing when prefix is already present")
}

func TestWithToken_EmptyStaysEmpty(t *testing.T) {
	// Empty token must NOT become "Bearer " — that would attach an
	// empty Authorization header to every request.
	c, err := NewClient(WithBaseURL("http://placeholder"), WithToken(""))
	require.NoError(t, err)
	require.Equal(t, "", c.cfg.token)
}

func TestRequest_TokenProvider_ErrorPropagates(t *testing.T) {
	wantErr := errors.New("token failed")
	c, err := NewClient(WithBaseURL("http://placeholder"), WithTokenProvider(func(ctx context.Context) (string, error) {
		return "", wantErr
	}))
	require.NoError(t, err)
	err = c.Do(context.Background(), http.MethodGet, "/", nil, nil)
	require.ErrorIs(t, err, wantErr)
}

func TestRequest_DefaultHeadersIncluded(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("X-Custom")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL), WithHeader("X-Custom", "default"))
	require.NoError(t, err)
	require.NoError(t, c.Do(context.Background(), http.MethodGet, "/", nil, nil))
	require.Equal(t, "default", seen)
}

func TestRequest_PerRequestHeaderOverrides(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("X-Custom")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL), WithHeader("X-Custom", "default"))
	require.NoError(t, err)
	require.NoError(t, c.Do(context.Background(), http.MethodGet, "/", nil, nil, httpx.WithRequestHeader("X-Custom", "override")))
	require.Equal(t, "override", seen)
}

func TestNewClientFromEnv_ReadsEnvVars(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://api.example.com/")
	// PROTEOS_API_TOKEN ergonomics: users set the raw JWT in env;
	// WithToken normalizes to the full header at construction.
	t.Setenv(EnvToken, "env-token")
	c, err := NewClientFromEnv()
	require.NoError(t, err)
	require.Equal(t, "https://api.example.com", c.BaseURL())
	require.Equal(t, "Bearer env-token", c.cfg.token)
}

func TestNewClientFromEnv_MissingURLErrors(t *testing.T) {
	t.Setenv(EnvBaseURL, "")
	t.Setenv(EnvToken, "")
	_, err := NewClientFromEnv()
	require.Error(t, err)
	require.Contains(t, err.Error(), EnvBaseURL)
}

func TestNewClientFromEnv_OptsOverrideEnv(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://from-env.example.com")
	t.Setenv(EnvToken, "env-token")
	c, err := NewClientFromEnv(WithToken("explicit"))
	require.NoError(t, err)
	require.Equal(t, "Bearer explicit", c.cfg.token, "later options should win over env")
}

func TestRequest_ParseFailureReturnsParseError(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	})
	var out struct {
		Foo string `json:"foo"`
	}
	err := c.Do(context.Background(), http.MethodGet, "/x", nil, &out)
	require.Error(t, err)
	var sdkErr *Error
	require.True(t, errors.As(err, &sdkErr))
	require.Equal(t, ErrCodeParse, sdkErr.Code)
}

func TestRequestMultipart_BuildsCorrectBody(t *testing.T) {
	var fields map[string][]string
	var fileBytes []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseMultipartForm(32<<20))
		fields = r.MultipartForm.Value
		hdr := r.MultipartForm.File["file"][0]
		f, err := hdr.Open()
		require.NoError(t, err)
		defer f.Close()
		fileBytes, _ = io.ReadAll(f)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL), WithToken("t"))
	require.NoError(t, err)

	var out map[string]any
	err = c.DoMultipart(
		context.Background(),
		http.MethodPost,
		"/upload",
		map[string]string{"slug": "s", "name": "n"},
		httpx.MultipartFile{
			FieldName:   "file",
			Filename:    "module.wasm",
			ContentType: "application/wasm",
			Reader:      strings.NewReader("WASM-BYTES"),
		},
		&out,
	)
	require.NoError(t, err)
	require.Equal(t, []string{"s"}, fields["slug"])
	require.Equal(t, []string{"n"}, fields["name"])
	require.Equal(t, "WASM-BYTES", string(fileBytes))
}

func TestRequestMultipart_HonorsContentTypeBoundary(t *testing.T) {
	var ct string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.DoMultipart(context.Background(), http.MethodPost, "/x", nil, httpx.MultipartFile{
		FieldName: "f", Filename: "x.bin", Reader: strings.NewReader("hi"),
	}, nil))
	require.True(t, strings.HasPrefix(ct, "multipart/form-data; boundary="), "got %q", ct)
}

func TestRequestRaw_StreamsBody(t *testing.T) {
	payload := strings.Repeat("X", 1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, payload)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL))
	require.NoError(t, err)

	body, resp, err := c.DoRaw(context.Background(), http.MethodGet, "/x", nil)
	require.NoError(t, err)
	defer body.Close()
	require.Equal(t, 200, resp.StatusCode)
	got, err := io.ReadAll(body)
	require.NoError(t, err)
	require.Equal(t, payload, string(got))
}

func TestRequestRaw_ReturnsErrorForNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"code":"not_found","message":"x"}`))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL))
	require.NoError(t, err)
	body, _, err := c.DoRaw(context.Background(), http.MethodGet, "/x", nil)
	require.Nil(t, body)
	require.True(t, IsNotFound(err))
}

// Confirm multipart and embedded use of *http.Client share their plumbing.
func TestWithHTTPClient_UsesProvidedClient(t *testing.T) {
	var hits int
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		hits++
		rec := httptest.NewRecorder()
		rec.WriteHeader(204)
		return rec.Result(), nil
	})
	c, err := NewClient(WithBaseURL("https://placeholder"), WithHTTPClient(&http.Client{Transport: rt}))
	require.NoError(t, err)
	require.NoError(t, c.Do(context.Background(), http.MethodGet, "/x", nil, nil))
	require.Equal(t, 1, hits)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Ensure form data fields don't bleed into JSON content type when no file.
func TestRequestMultipart_NoFile(t *testing.T) {
	var hadFile bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseMultipartForm(32<<20))
		hadFile = r.MultipartForm.File != nil && len(r.MultipartForm.File) > 0
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(WithBaseURL(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.DoMultipart(context.Background(), http.MethodPost, "/x",
		map[string]string{"only": "field"}, httpx.MultipartFile{}, nil))
	require.False(t, hadFile)
}

// Multipart writer is referenced to keep import; avoid unused dep warning.
var _ = multipart.NewWriter
