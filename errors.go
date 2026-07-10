package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrCode is the API-defined error code.
type ErrCode string

const (
	ErrCodeNotFound            ErrCode = "not_found"
	ErrCodeUnauthorized        ErrCode = "unauthorized"
	ErrCodeForbidden           ErrCode = "forbidden"
	ErrCodeBadRequest          ErrCode = "bad_request"
	ErrCodeConflict            ErrCode = "conflict"
	ErrCodeInternalServerError ErrCode = "internal_server_error"
	ErrCodeInvalidPayload      ErrCode = "invalid_payload"
	ErrCodeInvalidPersona      ErrCode = "invalid_persona"
	ErrCodeParse               ErrCode = "parse_error"
)

// Error is returned for all non-2xx HTTP responses.
type Error struct {
	HTTPStatus int
	Code       ErrCode
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s (HTTP %d)", e.Code, e.Message, e.HTTPStatus)
}

// Is supports errors.Is so callers can match on a partial template:
//
//	if errors.Is(err, &sdk.Error{HTTPStatus: 404}) { ... }
//
// A zero HTTPStatus or empty Code on the target acts as a wildcard.
func (e *Error) Is(target error) bool {
	var t *Error
	if !errors.As(target, &t) {
		return false
	}
	if t.HTTPStatus != 0 && t.HTTPStatus != e.HTTPStatus {
		return false
	}
	if t.Code != "" && t.Code != e.Code {
		return false
	}
	return true
}

// IsProteosError reports whether err is or wraps a *sdk.Error.
func IsProteosError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

// IsNotFound reports whether err is or wraps a *sdk.Error with HTTP 404.
func IsNotFound(err error) bool { return statusIs(err, http.StatusNotFound) }

// IsUnauthorized reports whether err is or wraps a *sdk.Error with HTTP 401.
func IsUnauthorized(err error) bool { return statusIs(err, http.StatusUnauthorized) }

// IsForbidden reports whether err is or wraps a *sdk.Error with HTTP 403.
func IsForbidden(err error) bool { return statusIs(err, http.StatusForbidden) }

// IsBadRequest reports whether err is or wraps a *sdk.Error with HTTP 400.
func IsBadRequest(err error) bool { return statusIs(err, http.StatusBadRequest) }

// IsConflict reports whether err is or wraps a *sdk.Error with HTTP 409.
func IsConflict(err error) bool { return statusIs(err, http.StatusConflict) }

func statusIs(err error, status int) bool {
	var e *Error
	return errors.As(err, &e) && e.HTTPStatus == status
}

func defaultErrCode(status int) ErrCode {
	switch status {
	case http.StatusBadRequest:
		return ErrCodeBadRequest
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeForbidden
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeConflict
	default:
		return ErrCodeInternalServerError
	}
}

type apiErrorResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// parseErrorResponse builds a *sdk.Error from a non-2xx response.
// It tries to decode {code, message} JSON; falls back to the raw body or
// the HTTP status string.
func parseErrorResponse(resp *http.Response) *Error {
	e := &Error{
		HTTPStatus: resp.StatusCode,
		Code:       defaultErrCode(resp.StatusCode),
		Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil || len(body) == 0 {
		return e
	}
	var apiErr apiErrorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && (apiErr.Code != "" || apiErr.Message != "") {
		if apiErr.Code != "" {
			e.Code = ErrCode(apiErr.Code)
		}
		if apiErr.Message != "" {
			e.Message = apiErr.Message
		}
		return e
	}
	e.Message = string(body)
	return e
}
