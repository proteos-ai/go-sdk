package functions

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/internal/httpx"
	functionsmodel "go.proteos.ai/model/functions"
)

const logsBasePath = "/functions/v1/logs"

// tailStreamTimeout caps any single Tail*() HTTP request. The SDK
// default 30s timeout otherwise kills long-lived follow streams with
// `context deadline exceeded`; 1h is long enough that interactive
// tailing sessions don't notice it but still finite, so a wedged
// connection eventually drops instead of leaking forever.
const tailStreamTimeout = 1 * time.Hour

// LogServiceAPI is the contract a LogService satisfies. Backs the
// org-wide log endpoint (`GET /functions/v1/logs`); `Hooks` and
// `Actions` on ListLogsOptions are repeatable filters that produce a
// UNION when both are populated.
type LogServiceAPI interface {
	Query(ctx context.Context, opts *ListLogsOptions) ([]functionsmodel.LogEntry, error)
	Tail(ctx context.Context, opts *ListLogsOptions) (*LogStream, error)
}

// LogService is the org-wide log feed client.
type LogService struct{ c *sdk.Client }

var _ LogServiceAPI = (*LogService)(nil)

// ListLogsOptions filters the org-wide `GET /functions/v1/logs` feed.
// `Hooks` and `Actions` are repeatable; when both are nil/empty the
// endpoint returns every entry. See server-side `GetLogsQuery` for the
// full UNION semantics.
type ListLogsOptions struct {
	Hooks   []string
	Actions []string
	Since   string
	Level   string
}

// Query fetches a snapshot — caller decides when to re-query for
// updates. Use Tail when you want a live stream.
func (s *LogService) Query(ctx context.Context, opts *ListLogsOptions) ([]functionsmodel.LogEntry, error) {
	path := logsBasePath
	if qs := optsToQuery(opts, false); qs != "" {
		path += "?" + qs
	}
	return doLogsList(ctx, s.c, path)
}

// Tail opens an NDJSON-streaming response. Callers MUST Close the
// returned stream to release the underlying connection. Each `Next` call
// returns one entry or signals end-of-stream.
func (s *LogService) Tail(ctx context.Context, opts *ListLogsOptions) (*LogStream, error) {
	path := logsBasePath + "?" + optsToQuery(opts, true)
	return openLogStream(ctx, s.c, path)
}

// TailLogs on the per-hook surface opens an NDJSON stream of new entries
// for the specified hook. Backs `pro functions hooks logs <slug> --follow`.
func (s *HookService) TailLogs(ctx context.Context, slug string, opts *ListHookLogsOptions) (*LogStream, error) {
	path := hooksBasePath + "/" + slug + "/logs?" + perResourceOptsToQuery(opts, true)
	return openLogStream(ctx, s.c, path)
}

// Logs on the per-action surface returns the buffered entries for the
// action.
func (s *ActionService) Logs(ctx context.Context, slug string, opts *ListActionLogsOptions) ([]functionsmodel.LogEntry, error) {
	path := actionsBasePath + "/" + slug + "/logs"
	if qs := perResourceActionOptsToQuery(opts, false); qs != "" {
		path += "?" + qs
	}
	return doLogsList(ctx, s.c, path)
}

// TailLogs on the per-action surface opens an NDJSON stream.
func (s *ActionService) TailLogs(ctx context.Context, slug string, opts *ListActionLogsOptions) (*LogStream, error) {
	path := actionsBasePath + "/" + slug + "/logs?" + perResourceActionOptsToQuery(opts, true)
	return openLogStream(ctx, s.c, path)
}

// ListActionLogsOptions mirrors ListHookLogsOptions for the per-action
// /logs endpoint. Same wire shape (`since`, `level`, `follow`).
type ListActionLogsOptions struct {
	Follow bool   `query:"follow,omitempty"`
	Since  string `query:"since,omitempty"`
	Level  string `query:"level,omitempty"`
}

// LogStream wraps an open NDJSON-streaming response. Iterate with
// `Next` until it returns ok=false; remember to `Close` to release the
// underlying HTTP connection regardless of whether the iteration
// completed naturally.
type LogStream struct {
	rc      io.ReadCloser
	scanner *bufio.Scanner
}

// Next blocks until the next entry arrives, an error occurs, or the
// stream ends. `(entry, true, nil)` on a successful read, `(zero, false,
// nil)` on natural end-of-stream, `(zero, false, err)` on transport /
// decode error.
func (stream *LogStream) Next() (functionsmodel.LogEntry, bool, error) {
	var zero functionsmodel.LogEntry
	if !stream.scanner.Scan() {
		if err := stream.scanner.Err(); err != nil {
			return zero, false, err
		}
		return zero, false, nil
	}
	line := stream.scanner.Bytes()
	if len(line) == 0 {
		// Defensive: NDJSON should never emit empty lines, but skip
		// gracefully if a proxy adds one rather than erroring the iter.
		return stream.Next()
	}
	var entry functionsmodel.LogEntry
	if err := json.Unmarshal(line, &entry); err != nil {
		return zero, false, fmt.Errorf("functions: decode log entry: %w", err)
	}
	return entry, true, nil
}

// Close releases the underlying response body. Always defer this.
func (stream *LogStream) Close() error {
	if stream.rc == nil {
		return nil
	}
	return stream.rc.Close()
}

// openLogStream issues the request and wraps the response body in a
// bufio-backed scanner sized for typical log-line lengths.
func openLogStream(ctx context.Context, c *sdk.Client, path string) (*LogStream, error) {
	rc, _, err := c.DoRaw(ctx, http.MethodGet, path, nil, httpx.WithRequestTimeout(tailStreamTimeout))
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(rc)
	// Default 64KB buffer is too small for log lines with rich
	// structured fields. 1MB is generous but bounded; entries beyond
	// this size are a configuration smell upstream.
	scanner.Buffer(make([]byte, 0, 4096), 1<<20)
	return &LogStream{rc: rc, scanner: scanner}, nil
}

// doLogsList performs the GET and unwraps the `{data: [...]}` envelope.
// Used by all three non-follow log endpoints.
func doLogsList(ctx context.Context, c *sdk.Client, path string) ([]functionsmodel.LogEntry, error) {
	var envelope struct {
		Data []functionsmodel.LogEntry `json:"data"`
	}
	if err := c.Do(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// optsToQuery encodes ListLogsOptions as a URL query string. Reuses
// `url.Values` so cobra's StringSlice round-trips correctly (each entry
// gets its own `hook=` / `action=` key, not a comma-joined value).
func optsToQuery(opts *ListLogsOptions, follow bool) string {
	values := url.Values{}
	if opts != nil {
		for _, hook := range opts.Hooks {
			if hook != "" {
				values.Add("hook", hook)
			}
		}
		for _, action := range opts.Actions {
			if action != "" {
				values.Add("action", action)
			}
		}
		if opts.Since != "" {
			values.Set("since", opts.Since)
		}
		if opts.Level != "" {
			values.Set("level", opts.Level)
		}
	}
	if follow {
		values.Set("follow", "true")
	}
	return values.Encode()
}

// perResourceOptsToQuery encodes the per-hook /logs query string.
func perResourceOptsToQuery(opts *ListHookLogsOptions, forceFollow bool) string {
	values := url.Values{}
	if opts != nil {
		if opts.Since != "" {
			values.Set("since", opts.Since)
		}
		if opts.Level != "" {
			values.Set("level", opts.Level)
		}
	}
	if forceFollow || (opts != nil && opts.Follow) {
		values.Set("follow", "true")
	}
	return values.Encode()
}

// perResourceActionOptsToQuery is the action-side mirror.
func perResourceActionOptsToQuery(opts *ListActionLogsOptions, forceFollow bool) string {
	values := url.Values{}
	if opts != nil {
		if opts.Since != "" {
			values.Set("since", opts.Since)
		}
		if opts.Level != "" {
			values.Set("level", opts.Level)
		}
	}
	if forceFollow || (opts != nil && opts.Follow) {
		values.Set("follow", "true")
	}
	return values.Encode()
}

