// Package httpx is the internal HTTP plumbing for the SDK. It is not part of
// the public API.
package httpx

import (
	"io"
	"net/http"
	"time"
)

// MultipartFile describes a file part for a multipart/form-data request.
type MultipartFile struct {
	FieldName   string
	Filename    string
	ContentType string
	Reader      io.Reader
}

// RequestOption configures a single request. Constructed via the helpers
// below; applied to the outgoing http.Request.
type RequestOption func(*RequestConfig)

// RequestConfig holds resolved per-request settings.
type RequestConfig struct {
	Headers http.Header
	Timeout time.Duration
}

// NewRequestConfig applies opts to a fresh RequestConfig.
func NewRequestConfig(opts ...RequestOption) RequestConfig {
	cfg := RequestConfig{Headers: http.Header{}}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// WithRequestTimeout overrides the client default timeout for one request.
func WithRequestTimeout(d time.Duration) RequestOption {
	return func(c *RequestConfig) { c.Timeout = d }
}

// WithRequestHeader sets a header on a single request.
func WithRequestHeader(key, value string) RequestOption {
	return func(c *RequestConfig) { c.Headers.Set(key, value) }
}
