package client

import (
	"encoding/json"
	"fmt"
	"io"
)

// PaginatedResponse represents the v2 API paginated response format.
type PaginatedResponse struct {
	Data json.RawMessage `json:"data"`
	Meta PaginationMeta  `json:"meta"`
}

// PaginationMeta contains pagination and request metadata.
type PaginationMeta struct {
	RequestID  string     `json:"request_id"`
	Pagination Pagination `json:"pagination"`
}

// Pagination contains cursor-based pagination info.
type Pagination struct {
	NextCursor     *string `json:"next_cursor"`
	PreviousCursor *string `json:"previous_cursor"`
	PerPage        int     `json:"per_page"`
}

// HasNextPage returns true if there is a next page.
func (p *PaginatedResponse) HasNextPage() bool {
	return p.Meta.Pagination.NextCursor != nil && *p.Meta.Pagination.NextCursor != ""
}

// NextCursor returns the cursor for the next page, or empty string.
func (p *PaginatedResponse) NextCursor() string {
	if p.Meta.Pagination.NextCursor != nil {
		return *p.Meta.Pagination.NextCursor
	}
	return ""
}

// ParsePaginatedResponse parses a raw JSON body into PaginatedResponse.
func ParsePaginatedResponse(body []byte) (*PaginatedResponse, error) {
	var pr PaginatedResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("parse paginated response: %w", err)
	}
	return &pr, nil
}

// StreamNDJSON writes each item from data array as one NDJSON line.
// Used for --page-all --output ndjson streaming.
func StreamNDJSON(w io.Writer, data json.RawMessage) error {
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("parse data array: %w", err)
	}
	enc := json.NewEncoder(w)
	for _, item := range items {
		if err := enc.Encode(item); err != nil {
			return err
		}
	}
	return nil
}
