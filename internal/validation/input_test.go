package validation

import "testing"

func TestValidateSafeOutputDir(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"normal path", "/v2/projects/", false},
		{"path with id", "/v2/projects/123/", false},
		{"traversal dotdot", "../../etc/passwd", true},
		{"mid traversal", "/v2/../admin/", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSafeOutputDir(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSafeOutputDir(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestRejectControlChars(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"normal text", "hello world", false},
		{"unicode ok", "한글 테스트", false},
		{"newline ok", "hello\nworld", false},
		{"tab char", "hello\tworld", true},
		{"null byte", "hello\x00world", true},
		{"bell char", "hello\x07world", true},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RejectControlChars(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RejectControlChars(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateResourceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"numeric", "123", false},
		{"uuid", "abc-def-123", false},
		{"with query", "123?page=2", true},
		{"with fragment", "123#section", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResourceID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestRejectDoubleEncoding(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"normal", "hello world", false},
		{"percent encoded", "hello%20world", true},
		{"double encoded", "hello%2520world", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RejectDoubleEncoding(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RejectDoubleEncoding(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateServerURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https", "https://api.example.com", false},
		{"http", "http://localhost:8000", false},
		{"with path", "https://api.example.com/v2", false},
		{"no scheme", "api.example.com", true},
		{"ftp", "ftp://files.example.com", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServerURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid", "/v2/projects/", false},
		{"traversal", "../admin/", true},
		{"control char", "/v2/\x00projects/", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
