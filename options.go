package sdk

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// TokenProvider returns a fresh bearer token. Called before each request.
// Useful when tokens are short-lived and need refresh.
type TokenProvider func(context.Context) (string, error)

// DefaultTimeout is applied to a request when neither WithTimeout nor a
// per-request timeout is set.
const DefaultTimeout = 30 * time.Second

// Option configures a Client. Apply via NewClient or NewClientFromEnv.
type Option func(*config)

type config struct {
	baseURL       string
	token         string
	tokenProvider TokenProvider
	httpClient    *http.Client
	timeout       time.Duration
	headers       http.Header
	logger        zerolog.Logger
}

func defaultConfig() *config {
	return &config{
		// Default transport is OTel-instrumented: every SDK request becomes a
		// client span and carries W3C traceparent so the callee continues the
		// trace. Callers who pass WithHTTPClient opt out (wrap it themselves to
		// keep spans). Requests are built with NewRequestWithContext, so spans
		// parent under the caller's active span.
		httpClient: &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
		timeout:    DefaultTimeout,
		headers:    http.Header{},
		logger:     zerolog.Nop(),
	}
}

// WithBaseURL sets the API base URL (e.g. "https://api.proteos.ai").
// A trailing slash is stripped. This is required.
func WithBaseURL(u string) Option {
	return func(c *config) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithToken sets a static bearer token. Use WithTokenProvider for dynamic
// (refresh) scenarios. If both are set, WithTokenProvider takes precedence.
//
// Accepts either a raw JWT (`eyJ...`) or a full Authorization header
// value (`Bearer eyJ...`). For ergonomics — `PROTEOS_API_TOKEN=eyJ...`
// in the environment should work without the user manually prefixing
// — the value is normalized once here to the full header form. After
// this point the token travels through the codebase verbatim; the
// outbound request emits it as-is.
func WithToken(t string) Option {
	return func(c *config) { c.token = ensureBearerPrefix(t) }
}

// WithTokenProvider sets a function that returns a fresh token before each
// request. The provider MUST return the full Authorization header value
// (e.g. `Bearer eyJ...`) or an empty string for no header. The SDK does
// not normalize provider output — the value is emitted verbatim.
func WithTokenProvider(p TokenProvider) Option {
	return func(c *config) { c.tokenProvider = p }
}

// ensureBearerPrefix is the single internal normalization point in
// the SDK. Returns "" for "" so the empty-token case still suppresses
// the Authorization header.
func ensureBearerPrefix(t string) string {
	if t == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(t) >= len(prefix) && strings.EqualFold(t[:len(prefix)], prefix) {
		return t
	}
	return prefix + t
}

// WithHTTPClient overrides the *http.Client used for all requests. Useful
// for embedded use with custom transports, mTLS, hooks, etc.
func WithHTTPClient(h *http.Client) Option {
	return func(c *config) {
		if h != nil {
			c.httpClient = h
		}
	}
}

// WithTimeout sets the default per-request timeout. Defaults to
// DefaultTimeout (30s). The deadline is applied via context.WithTimeout.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// WithHeader adds a default header included in every request.
// Authorization is set automatically from token / tokenProvider and should
// not be set here.
func WithHeader(key, value string) Option {
	return func(c *config) { c.headers.Set(key, value) }
}

// WithLogger injects a zerolog logger. Defaults to zerolog.Nop().
func WithLogger(l zerolog.Logger) Option {
	return func(c *config) { c.logger = l }
}
