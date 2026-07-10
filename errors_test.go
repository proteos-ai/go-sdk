package sdk

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_FormatString(t *testing.T) {
	e := &Error{HTTPStatus: 404, Code: ErrCodeNotFound, Message: "missing entity 'foo'"}
	require.Equal(t, "not_found: missing entity 'foo' (HTTP 404)", e.Error())
}

func TestErrorIs_MatchesByStatus(t *testing.T) {
	err := error(&Error{HTTPStatus: 404, Code: ErrCodeNotFound, Message: "x"})
	require.True(t, errors.Is(err, &Error{HTTPStatus: 404}))
	require.False(t, errors.Is(err, &Error{HTTPStatus: 500}))
}

func TestErrorIs_MatchesByCode(t *testing.T) {
	err := error(&Error{HTTPStatus: 404, Code: ErrCodeNotFound, Message: "x"})
	require.True(t, errors.Is(err, &Error{Code: ErrCodeNotFound}))
	require.False(t, errors.Is(err, &Error{Code: ErrCodeBadRequest}))
}

func TestErrorIs_WildcardOnEmptyTarget(t *testing.T) {
	err := error(&Error{HTTPStatus: 401, Code: ErrCodeUnauthorized})
	require.True(t, errors.Is(err, &Error{}), "empty target acts as a wildcard for any *Error")
}

func TestStatusHelpers(t *testing.T) {
	cases := []struct {
		name   string
		status int
		check  func(error) bool
	}{
		{"not found", 404, IsNotFound},
		{"unauthorized", 401, IsUnauthorized},
		{"forbidden", 403, IsForbidden},
		{"bad request", 400, IsBadRequest},
		{"conflict", 409, IsConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := error(&Error{HTTPStatus: tc.status})
			require.True(t, tc.check(err), "expected helper to match status %d", tc.status)
			require.True(t, IsProteosError(err))
			// Different status should not match.
			other := error(&Error{HTTPStatus: 500})
			require.False(t, tc.check(other))
		})
	}
}

func TestStatusHelpers_NonProteosError(t *testing.T) {
	require.False(t, IsNotFound(errors.New("plain")))
	require.False(t, IsUnauthorized(nil))
	require.False(t, IsProteosError(errors.New("plain")))
}

func TestStatusHelpers_WrappedError(t *testing.T) {
	inner := &Error{HTTPStatus: 404, Code: ErrCodeNotFound}
	wrapped := fmt.Errorf("get user: %w", inner)
	require.True(t, IsNotFound(wrapped))
	require.True(t, IsProteosError(wrapped))
}

func TestParseErrorResponse_JSONBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader(`{"code":"not_found","message":"missing"}`)),
	}
	e := parseErrorResponse(resp)
	require.Equal(t, 404, e.HTTPStatus)
	require.Equal(t, ErrCodeNotFound, e.Code)
	require.Equal(t, "missing", e.Message)
}

func TestParseErrorResponse_PlainTextBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader("boom")),
	}
	e := parseErrorResponse(resp)
	require.Equal(t, 500, e.HTTPStatus)
	require.Equal(t, ErrCodeInternalServerError, e.Code)
	require.Equal(t, "boom", e.Message)
}

func TestParseErrorResponse_EmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 401,
		Body:       io.NopCloser(strings.NewReader("")),
	}
	e := parseErrorResponse(resp)
	require.Equal(t, 401, e.HTTPStatus)
	require.Equal(t, ErrCodeUnauthorized, e.Code)
	require.Equal(t, "HTTP 401", e.Message)
}

func TestParseErrorResponse_PartialJSONBody(t *testing.T) {
	// JSON with only a code, no message — should keep code, fall back for message.
	resp := &http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(strings.NewReader(`{"code":"CUSTOM_CODE"}`)),
	}
	e := parseErrorResponse(resp)
	require.Equal(t, 400, e.HTTPStatus)
	require.Equal(t, ErrCode("CUSTOM_CODE"), e.Code)
	require.Equal(t, "HTTP 400", e.Message)
}

func TestDefaultErrCode_Mapping(t *testing.T) {
	require.Equal(t, ErrCodeBadRequest, defaultErrCode(400))
	require.Equal(t, ErrCodeUnauthorized, defaultErrCode(401))
	require.Equal(t, ErrCodeForbidden, defaultErrCode(403))
	require.Equal(t, ErrCodeNotFound, defaultErrCode(404))
	require.Equal(t, ErrCodeConflict, defaultErrCode(409))
	require.Equal(t, ErrCodeInternalServerError, defaultErrCode(500))
	require.Equal(t, ErrCodeInternalServerError, defaultErrCode(502))
}
