package i18n

import (
	"strings"
	"testing"
)

func TestT_English(t *testing.T) {
	Init("en")

	got := T("ServerNotConfigured")
	if !strings.Contains(got, "Server is not configured") {
		t.Fatalf("expected English message, got %q", got)
	}
}

func TestT_Korean(t *testing.T) {
	Init("ko")

	got := T("ServerNotConfigured")
	if !strings.Contains(got, "서버가 설정되지 않았습니다") {
		t.Fatalf("expected Korean message, got %q", got)
	}
}

func TestT_WithData(t *testing.T) {
	Init("en")

	got := T("LoginSuccess", map[string]interface{}{"Context": "production"})
	if !strings.Contains(got, "production") {
		t.Fatalf("expected data interpolation, got %q", got)
	}
}

func TestT_Fallback(t *testing.T) {
	Init("en")

	got := T("NonExistentMessageID")
	if got != "NonExistentMessageID" {
		t.Fatalf("expected fallback to message ID, got %q", got)
	}
}

func TestT_UnsupportedLanguageFallsBackToEnglish(t *testing.T) {
	Init("fr")

	got := T("AuthRequired")
	if !strings.Contains(got, "Authentication required") {
		t.Fatalf("expected English fallback for unsupported lang, got %q", got)
	}
}

func TestCurrentLang(t *testing.T) {
	Init("ko")
	if got := CurrentLang(); got != "ko" {
		t.Fatalf("expected ko, got %s", got)
	}

	Init("en")
	if got := CurrentLang(); got != "en" {
		t.Fatalf("expected en, got %s", got)
	}
}
