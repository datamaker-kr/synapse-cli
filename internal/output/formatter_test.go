package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestDetectFormat_ExplicitFlag(t *testing.T) {
	if got := DetectFormat("json"); got != "json" {
		t.Fatalf("expected json, got %s", got)
	}
	if got := DetectFormat("yaml"); got != "yaml" {
		t.Fatalf("expected yaml, got %s", got)
	}
}

func TestJSONFormatter_Format(t *testing.T) {
	buf := new(bytes.Buffer)
	f := NewFormatter("json", buf)
	data := map[string]string{"name": "test", "id": "1"}
	if err := f.Format(data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"name": "test"`) {
		t.Fatalf("expected JSON output, got %s", out)
	}
}

func TestNDJSONFormatter_Format(t *testing.T) {
	buf := new(bytes.Buffer)
	f := NewFormatter("ndjson", buf)
	data := map[string]string{"name": "test"}
	if err := f.Format(data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if !strings.Contains(out, `"name":"test"`) {
		t.Fatalf("expected NDJSON output, got %s", out)
	}
	// Should be single line (no indentation)
	if strings.Contains(out, "\n") {
		t.Fatalf("NDJSON should be single line, got %s", out)
	}
}

func TestYAMLFormatter_Format(t *testing.T) {
	buf := new(bytes.Buffer)
	f := NewFormatter("yaml", buf)
	data := map[string]string{"name": "test"}
	if err := f.Format(data); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "name: test") {
		t.Fatalf("expected YAML output, got %s", out)
	}
}

func TestTableFormatter_FormatList(t *testing.T) {
	buf := new(bytes.Buffer)
	f := NewFormatter("table", buf)
	items := []map[string]string{
		{"id": "1", "name": "Project A"},
		{"id": "2", "name": "Project B"},
	}
	columns := []Column{
		{Header: "ID", Field: "id"},
		{Header: "Name", Field: "name"},
	}
	if err := f.FormatList(items, columns); err != nil {
		t.Fatalf("FormatList failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Project A") {
		t.Fatalf("expected table with Project A, got %s", out)
	}
	if !strings.Contains(out, "ID") {
		t.Fatalf("expected header ID, got %s", out)
	}
}

func TestWriteNDJSONItem(t *testing.T) {
	buf := new(bytes.Buffer)
	item := map[string]string{"id": "1"}
	if err := WriteNDJSONItem(buf, item); err != nil {
		t.Fatalf("WriteNDJSONItem failed: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != `{"id":"1"}` {
		t.Fatalf("expected single JSON line, got %s", out)
	}
}
