package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// APIError represents the v2 API standard error response.
type APIError struct {
	StatusCode int
	ErrorBody  ErrorBody `json:"error"`
	Meta       Meta      `json:"meta"`
	RawBody    []byte    `json:"-"`
}

// ErrorBody is the error detail from the v2 API.
type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details"`
}

// FieldError is a field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Meta contains request metadata.
type Meta struct {
	RequestID string `json:"request_id"`
}

func (e *APIError) Error() string {
	if e == nil {
		return "unknown error"
	}
	if e.ErrorBody.Message != "" {
		return e.ErrorBody.Message
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

// ParseAPIError parses an HTTP response into an APIError.
// Returns nil if the response is not an error (2xx).
func ParseAPIError(resp *http.Response) *APIError {
	if resp == nil {
		return &APIError{StatusCode: 0, ErrorBody: ErrorBody{Message: "no response"}}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		RawBody:    body,
	}

	// Try to parse v2 error format
	_ = json.Unmarshal(body, apiErr)

	// Set default messages for common status codes if not parsed
	if apiErr.ErrorBody.Message == "" {
		switch resp.StatusCode {
		case 401:
			apiErr.ErrorBody.Message = "Authentication required. Run: synapse login"
			apiErr.ErrorBody.Code = "AUTHENTICATION_REQUIRED"
		case 403:
			apiErr.ErrorBody.Message = "Permission denied."
			apiErr.ErrorBody.Code = "PERMISSION_DENIED"
		case 404:
			apiErr.ErrorBody.Message = "Resource not found."
			apiErr.ErrorBody.Code = "NOT_FOUND"
		default:
			apiErr.ErrorBody.Message = fmt.Sprintf("Server error (HTTP %d)", resp.StatusCode)
			apiErr.ErrorBody.Code = "SERVER_ERROR"
		}
	}

	return apiErr
}

// FormatHuman returns a human-friendly error string.
func (e *APIError) FormatHuman() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Error: %s", e.ErrorBody.Message))
	if e.ErrorBody.Code != "" {
		sb.WriteString(fmt.Sprintf(" (%s)", e.ErrorBody.Code))
	}
	sb.WriteString("\n")

	for _, d := range e.ErrorBody.Details {
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", d.Field, d.Message))
	}

	if e.Meta.RequestID != "" {
		sb.WriteString(fmt.Sprintf("(request_id: %s)\n", e.Meta.RequestID))
	}

	return sb.String()
}

// FormatJSON returns the raw error JSON.
func (e *APIError) FormatJSON() string {
	if len(e.RawBody) > 0 {
		return string(e.RawBody)
	}
	b, _ := json.MarshalIndent(e, "", "  ")
	return string(b)
}
