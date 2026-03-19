package client

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func makeResp(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestParseAPIError_Success(t *testing.T) {
	resp := makeResp(200, `{"data": []}`)
	if err := ParseAPIError(resp); err != nil {
		t.Fatalf("expected nil for 200, got %v", err)
	}
}

func TestParseAPIError_ValidationError(t *testing.T) {
	body := `{
		"error": {
			"code": "VALIDATION_ERROR",
			"message": "Invalid input.",
			"details": [{"field": "name", "message": "This field is required."}]
		},
		"meta": {"request_id": "req_abc123"}
	}`
	resp := makeResp(400, body)
	apiErr := ParseAPIError(resp)
	if apiErr == nil {
		t.Fatal("expected error for 400")
	}
	if apiErr.ErrorBody.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %s", apiErr.ErrorBody.Code)
	}
	if len(apiErr.ErrorBody.Details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(apiErr.ErrorBody.Details))
	}
	if apiErr.Meta.RequestID != "req_abc123" {
		t.Fatalf("expected request_id req_abc123, got %s", apiErr.Meta.RequestID)
	}
}

func TestParseAPIError_401(t *testing.T) {
	resp := makeResp(401, ``)
	apiErr := ParseAPIError(resp)
	if apiErr == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(apiErr.ErrorBody.Message, "synapse login") {
		t.Fatalf("expected login hint, got %s", apiErr.ErrorBody.Message)
	}
}

func TestParseAPIError_403(t *testing.T) {
	resp := makeResp(403, ``)
	apiErr := ParseAPIError(resp)
	if apiErr == nil {
		t.Fatal("expected error for 403")
	}
	if !strings.Contains(apiErr.ErrorBody.Message, "Permission") {
		t.Fatalf("expected permission message, got %s", apiErr.ErrorBody.Message)
	}
}

func TestParseAPIError_404(t *testing.T) {
	resp := makeResp(404, ``)
	apiErr := ParseAPIError(resp)
	if apiErr == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(apiErr.ErrorBody.Message, "not found") {
		t.Fatalf("expected not found message, got %s", apiErr.ErrorBody.Message)
	}
}

func TestAPIError_FormatHuman(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 400,
		ErrorBody: ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid input.",
			Details: []FieldError{
				{Field: "name", Message: "Required."},
				{Field: "category", Message: "Invalid."},
			},
		},
		Meta: Meta{RequestID: "req_123"},
	}

	out := apiErr.FormatHuman()
	if !strings.Contains(out, "Invalid input.") {
		t.Fatalf("expected message in output: %s", out)
	}
	if !strings.Contains(out, "name: Required.") {
		t.Fatalf("expected field detail: %s", out)
	}
	if !strings.Contains(out, "req_123") {
		t.Fatalf("expected request_id: %s", out)
	}
}

func TestAPIError_Error(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 400,
		ErrorBody:  ErrorBody{Message: "bad request"},
	}
	if apiErr.Error() != "bad request" {
		t.Fatalf("expected 'bad request', got %s", apiErr.Error())
	}
}
