package validation

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateSafeOutputDir rejects path traversal patterns.
func ValidateSafeOutputDir(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid path: path traversal detected %q. Example: /v2/projects/", path)
	}
	return nil
}

// RejectControlChars rejects input containing ASCII control characters (< 0x20),
// except for standard whitespace in some contexts.
func RejectControlChars(input string) error {
	for i, r := range input {
		if r < 0x20 && r != '\n' && r != '\r' {
			return fmt.Errorf("invalid input: control character at position %d (0x%02x)", i, r)
		}
	}
	return nil
}

// ValidateResourceID rejects resource IDs containing query/fragment characters.
func ValidateResourceID(id string) error {
	if strings.ContainsAny(id, "?#") {
		return fmt.Errorf("invalid resource ID %q: must not contain '?' or '#'. Example: 123", id)
	}
	return nil
}

// RejectDoubleEncoding rejects input containing percent signs (likely double-encoded).
func RejectDoubleEncoding(input string) error {
	if strings.Contains(input, "%") {
		return fmt.Errorf("invalid input %q: percent-encoding detected (possible double encoding). Use plain text", input)
	}
	return nil
}

// ValidateServerURL validates that the input is a well-formed HTTP(S) URL.
func ValidateServerURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid URL %q: scheme must be http or https", rawURL)
	}
	if u.Host == "" {
		return fmt.Errorf("invalid URL %q: missing host", rawURL)
	}
	return nil
}

// ValidateAPIPath validates a path for the escape hatch API command.
func ValidateAPIPath(path string) error {
	if err := ValidateSafeOutputDir(path); err != nil {
		return err
	}
	if err := RejectControlChars(path); err != nil {
		return err
	}
	return nil
}
