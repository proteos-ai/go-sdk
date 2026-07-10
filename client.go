package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"
	"time"

	"go.proteos.ai/sdk/internal/httpx"
)

// Client is the base HTTP client. Sub-clients (account.Client, meta.Client,
// data.Client) are built from a *Client via their respective New
// constructors.
type Client struct {
	cfg *config
}

// NewClient builds a Client from the given options. WithBaseURL is required.
func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.baseURL == "" {
		return nil, errors.New("sdk: WithBaseURL is required")
	}
	return &Client{cfg: cfg}, nil
}

// EnvBaseURL is the environment variable read by NewClientFromEnv for the
// API base URL.
const EnvBaseURL = "PROTEOS_API_URL"

// EnvToken is the environment variable read by NewClientFromEnv for the
// bearer token.
const EnvToken = "PROTEOS_API_TOKEN"

// NewClientFromEnv builds a Client using PROTEOS_API_URL and
// PROTEOS_API_TOKEN, layering opts on top so callers can override or extend.
// Returns an error if PROTEOS_API_URL is unset.
func NewClientFromEnv(opts ...Option) (*Client, error) {
	base := os.Getenv(EnvBaseURL)
	if base == "" {
		return nil, fmt.Errorf("sdk: %s is not set", EnvBaseURL)
	}
	merged := []Option{WithBaseURL(base)}
	if tok := os.Getenv(EnvToken); tok != "" {
		merged = append(merged, WithToken(tok))
	}
	merged = append(merged, opts...)
	return NewClient(merged...)
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string { return c.cfg.baseURL }

// Do performs an HTTP request and JSON-decodes the response into out (if
// non-nil). Returns *Error for non-2xx responses.
func (c *Client) Do(ctx context.Context, method, path string, body, out any, opts ...httpx.RequestOption) error {
	return c.DoWithQuery(ctx, method, path, nil, body, out, opts...)
}

// DoWithQuery is like Do but also encodes query as a query string. The query
// argument may be a struct, *struct, or map[string]any; see httpx.ToQueryParams.
func (c *Client) DoWithQuery(ctx context.Context, method, path string, query, body, out any, opts ...httpx.RequestOption) error {
	reqCfg := httpx.NewRequestConfig(opts...)

	url := c.cfg.baseURL + path
	if query != nil {
		if qs := httpx.ToQueryParams(query); qs != "" {
			url += "?" + qs
		}
	}

	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("sdk: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	ctx, cancel := c.withTimeout(ctx, reqCfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("sdk: new request: %w", err)
	}

	if err := c.applyHeaders(ctx, req, body != nil, reqCfg.Headers); err != nil {
		return err
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sdk: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp)
	}
	if resp.StatusCode == http.StatusNoContent || out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("sdk: read response body: %w", err)
	}
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return &Error{
			HTTPStatus: resp.StatusCode,
			Code:       ErrCodeParse,
			Message:    fmt.Sprintf("failed to parse response as JSON: %s", truncate(string(raw), 200)),
		}
	}
	return nil
}

// DoMultipart performs a multipart/form-data request. The fields map is
// written as text fields; file (if its FieldName is non-empty) is written
// as a file part.
func (c *Client) DoMultipart(ctx context.Context, method, path string, fields map[string]string, file httpx.MultipartFile, out any, opts ...httpx.RequestOption) error {
	reqCfg := httpx.NewRequestConfig(opts...)
	url := c.cfg.baseURL + path

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return fmt.Errorf("sdk: write multipart field %q: %w", k, err)
		}
	}
	if file.FieldName != "" && file.Reader != nil {
		hdr := make(textproto.MIMEHeader)
		hdr.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name=%q; filename=%q`, file.FieldName, file.Filename))
		if file.ContentType != "" {
			hdr.Set("Content-Type", file.ContentType)
		}
		fw, err := mw.CreatePart(hdr)
		if err != nil {
			return fmt.Errorf("sdk: create multipart file part: %w", err)
		}
		if _, err := io.Copy(fw, file.Reader); err != nil {
			return fmt.Errorf("sdk: copy multipart file: %w", err)
		}
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("sdk: close multipart writer: %w", err)
	}

	ctx, cancel := c.withTimeout(ctx, reqCfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return fmt.Errorf("sdk: new multipart request: %w", err)
	}
	// Apply default+per-request headers, but Content-Type comes from the
	// multipart writer (it carries the boundary).
	if err := c.applyHeaders(ctx, req, false, reqCfg.Headers); err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sdk: do multipart: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp)
	}
	if resp.StatusCode == http.StatusNoContent || out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// DoMultipartJSON performs a multipart/form-data request whose metadata part
// is JSON. Unlike DoMultipart (which writes metadata as a bare text field via
// WriteField), the part named metadataField is written FIRST with an explicit
// Content-Type: application/json header, then each file part in order.
// metadata-service + storage-service both enforce that the metadata part
// declares JSON and precedes the files, so deploys through them use this path
// rather than DoMultipart. Files are written in slice order, which the
// receiving controllers rely on (e.g. bundle before source).
func (c *Client) DoMultipartJSON(ctx context.Context, method, path, metadataField, metadataJSON string, files []httpx.MultipartFile, out any, opts ...httpx.RequestOption) error {
	reqCfg := httpx.NewRequestConfig(opts...)
	url := c.cfg.baseURL + path

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	metaHdr := make(textproto.MIMEHeader)
	metaHdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q`, metadataField))
	metaHdr.Set("Content-Type", "application/json")
	metaPart, err := mw.CreatePart(metaHdr)
	if err != nil {
		return fmt.Errorf("sdk: create multipart metadata part: %w", err)
	}
	if _, err := io.WriteString(metaPart, metadataJSON); err != nil {
		return fmt.Errorf("sdk: write multipart metadata: %w", err)
	}

	for _, file := range files {
		if file.FieldName == "" || file.Reader == nil {
			continue
		}
		hdr := make(textproto.MIMEHeader)
		hdr.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name=%q; filename=%q`, file.FieldName, file.Filename))
		if file.ContentType != "" {
			hdr.Set("Content-Type", file.ContentType)
		}
		fw, err := mw.CreatePart(hdr)
		if err != nil {
			return fmt.Errorf("sdk: create multipart file part %q: %w", file.FieldName, err)
		}
		if _, err := io.Copy(fw, file.Reader); err != nil {
			return fmt.Errorf("sdk: copy multipart file %q: %w", file.FieldName, err)
		}
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("sdk: close multipart writer: %w", err)
	}

	ctx, cancel := c.withTimeout(ctx, reqCfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return fmt.Errorf("sdk: new multipart request: %w", err)
	}
	if err := c.applyHeaders(ctx, req, false, reqCfg.Headers); err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sdk: do multipart: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp)
	}
	if resp.StatusCode == http.StatusNoContent || out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// DoRaw performs a request and returns the response body as an io.ReadCloser.
// The caller must Close the body. Used for streaming downloads.
func (c *Client) DoRaw(ctx context.Context, method, path string, body any, opts ...httpx.RequestOption) (io.ReadCloser, *http.Response, error) {
	reqCfg := httpx.NewRequestConfig(opts...)
	url := c.cfg.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("sdk: marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	ctx, cancel := c.withTimeout(ctx, reqCfg.Timeout)

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("sdk: new raw request: %w", err)
	}
	if err := c.applyHeaders(ctx, req, body != nil, reqCfg.Headers); err != nil {
		cancel()
		return nil, nil, err
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("sdk: do raw: %w", err)
	}
	if resp.StatusCode >= 400 {
		err := parseErrorResponse(resp)
		_ = resp.Body.Close()
		cancel()
		return nil, resp, err
	}
	// Wrap the body so closing it releases the timeout context too.
	return &cancelReadCloser{ReadCloser: resp.Body, cancel: cancel}, resp, nil
}

type cancelReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelReadCloser) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

func (c *Client) withTimeout(ctx context.Context, override time.Duration) (context.Context, context.CancelFunc) {
	d := override
	if d <= 0 {
		d = c.cfg.timeout
	}
	if d <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, d)
}

func (c *Client) applyHeaders(ctx context.Context, req *http.Request, hasBody bool, perReq http.Header) error {
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	for k, vs := range c.cfg.headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	for k, vs := range perReq {
		// Per-request headers replace defaults for the same key.
		req.Header.Del(k)
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	token := c.cfg.token
	if c.cfg.tokenProvider != nil {
		t, err := c.cfg.tokenProvider(ctx)
		if err != nil {
			return fmt.Errorf("sdk: token provider: %w", err)
		}
		token = t
	}
	if token != "" {
		// Token is the full Authorization header value (e.g.
		// `Bearer eyJ...`). WithToken normalizes static input at
		// construction; WithTokenProvider's contract is "return the
		// full header verbatim". Set as-is.
		req.Header.Set("Authorization", token)
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return strings.ToValidUTF8(s[:max], "") + "..."
}
