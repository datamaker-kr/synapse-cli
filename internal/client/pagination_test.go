package client

import (
	"bytes"
	"strings"
	"testing"
)

func TestParsePaginatedResponse(t *testing.T) {
	body := `{
		"data": [{"id": 1}, {"id": 2}],
		"meta": {
			"request_id": "req_123",
			"pagination": {
				"next_cursor": "abc",
				"previous_cursor": null,
				"per_page": 20
			}
		}
	}`

	pr, err := ParsePaginatedResponse([]byte(body))
	if err != nil {
		t.Fatalf("ParsePaginatedResponse failed: %v", err)
	}
	if !pr.HasNextPage() {
		t.Fatal("expected HasNextPage true")
	}
	if pr.NextCursor() != "abc" {
		t.Fatalf("expected cursor 'abc', got %q", pr.NextCursor())
	}
	if pr.Meta.RequestID != "req_123" {
		t.Fatalf("expected request_id req_123, got %s", pr.Meta.RequestID)
	}
}

func TestParsePaginatedResponse_NoNextPage(t *testing.T) {
	body := `{
		"data": [],
		"meta": {
			"request_id": "req_456",
			"pagination": {"next_cursor": null, "per_page": 20}
		}
	}`

	pr, err := ParsePaginatedResponse([]byte(body))
	if err != nil {
		t.Fatalf("ParsePaginatedResponse failed: %v", err)
	}
	if pr.HasNextPage() {
		t.Fatal("expected HasNextPage false")
	}
}

func TestStreamNDJSON(t *testing.T) {
	data := `[{"id":1,"name":"a"},{"id":2,"name":"b"}]`
	buf := new(bytes.Buffer)

	if err := StreamNDJSON(buf, []byte(data)); err != nil {
		t.Fatalf("StreamNDJSON failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}
	if !strings.Contains(lines[0], `"id":1`) {
		t.Fatalf("expected first line to contain id:1, got %s", lines[0])
	}
}
