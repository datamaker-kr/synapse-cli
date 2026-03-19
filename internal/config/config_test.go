package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir_Default(t *testing.T) {
	t.Setenv("SYNAPSE_CONFIG_DIR", "")
	dir := ConfigDir()
	if dir == "" {
		t.Fatal("ConfigDir returned empty string")
	}
}

func TestConfigDir_EnvOverride(t *testing.T) {
	t.Setenv("SYNAPSE_CONFIG_DIR", "/tmp/test-synapse")
	dir := ConfigDir()
	if dir != "/tmp/test-synapse" {
		t.Fatalf("expected /tmp/test-synapse, got %s", dir)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)

	original := &Config{
		CurrentContext: "prod",
		Language:       "ko",
		Contexts: map[string]ContextConfig{
			"prod": {
				Server:     "https://api.example.com",
				AuthMethod: "token",
				Token:      "secret-token-value",
				TenantCode: "ws-001",
			},
		},
	}

	if err := original.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected 0600, got %04o", perm)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if loaded.CurrentContext != "prod" {
		t.Fatalf("expected current_context=prod, got %s", loaded.CurrentContext)
	}
	if loaded.Language != "ko" {
		t.Fatalf("expected language=ko, got %s", loaded.Language)
	}
	ctx := loaded.Contexts["prod"]
	if ctx.Server != "https://api.example.com" {
		t.Fatalf("expected server=https://api.example.com, got %s", ctx.Server)
	}
	if ctx.Token != "secret-token-value" {
		t.Fatalf("expected token to persist, got %s", ctx.Token)
	}
}

func TestConfig_LoadMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if cfg.CurrentContext != "" {
		t.Fatalf("expected empty context, got %s", cfg.CurrentContext)
	}
}

func TestConfig_AddContext(t *testing.T) {
	cfg := &Config{Contexts: make(map[string]ContextConfig)}

	err := cfg.AddContext("staging", ContextConfig{Server: "https://staging.example.com"})
	if err != nil {
		t.Fatalf("AddContext failed: %v", err)
	}
	// First context should become current
	if cfg.CurrentContext != "staging" {
		t.Fatalf("expected current=staging, got %s", cfg.CurrentContext)
	}

	// Duplicate should fail
	err = cfg.AddContext("staging", ContextConfig{Server: "https://other.com"})
	if err == nil {
		t.Fatal("expected error for duplicate context")
	}
}

func TestConfig_SetContext(t *testing.T) {
	cfg := &Config{
		CurrentContext: "a",
		Contexts: map[string]ContextConfig{
			"a": {Server: "https://a.com"},
			"b": {Server: "https://b.com"},
		},
	}

	if err := cfg.SetContext("b"); err != nil {
		t.Fatalf("SetContext failed: %v", err)
	}
	if cfg.CurrentContext != "b" {
		t.Fatalf("expected current=b, got %s", cfg.CurrentContext)
	}

	if err := cfg.SetContext("nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent context")
	}
}

func TestConfig_DeleteContext(t *testing.T) {
	cfg := &Config{
		CurrentContext: "a",
		Contexts: map[string]ContextConfig{
			"a": {Server: "https://a.com"},
			"b": {Server: "https://b.com"},
		},
	}

	// Should fail without force for active context
	if err := cfg.DeleteContext("a", false); err == nil {
		t.Fatal("expected error when deleting active context without force")
	}

	// Should succeed with force
	if err := cfg.DeleteContext("a", true); err != nil {
		t.Fatalf("DeleteContext with force failed: %v", err)
	}
	if cfg.CurrentContext != "" {
		t.Fatalf("expected empty current, got %s", cfg.CurrentContext)
	}

	// Non-active context should delete without force
	if err := cfg.DeleteContext("b", false); err != nil {
		t.Fatalf("DeleteContext failed: %v", err)
	}
}

func TestConfig_ActiveContext_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SYNAPSE_CONFIG_DIR", dir)
	t.Setenv("SYNAPSE_SERVER", "https://env-override.com")
	t.Setenv("SYNAPSE_CONTEXT", "")

	cfg := &Config{
		CurrentContext: "prod",
		Contexts: map[string]ContextConfig{
			"prod": {Server: "https://original.com", Token: "tok123", TenantCode: "ws-001"},
		},
	}

	ctx, err := cfg.ActiveContext()
	if err != nil {
		t.Fatalf("ActiveContext failed: %v", err)
	}
	if ctx.Server != "https://env-override.com" {
		t.Fatalf("expected env override server, got %s", ctx.Server)
	}
	// Token should remain from config
	if ctx.Token != "tok123" {
		t.Fatalf("expected original token, got %s", ctx.Token)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "****"},
		{"abc", "****"},
		{"abcdefgh", "***...efgh"},
		{"syn_aBcDeFgHiJkLmNoPqRsT", "***...qRsT"},
	}
	for _, tt := range tests {
		got := MaskToken(tt.input)
		if got != tt.want {
			t.Errorf("MaskToken(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestConfig_ResolveLanguage(t *testing.T) {
	cfg := &Config{Language: "ko"}

	// Flag takes priority
	if lang := cfg.ResolveLanguage("en"); lang != "en" {
		t.Fatalf("expected en from flag, got %s", lang)
	}

	// Config language when no flag
	t.Setenv("SYNAPSE_LANG", "")
	if lang := cfg.ResolveLanguage(""); lang != "ko" {
		t.Fatalf("expected ko from config, got %s", lang)
	}

	// Env var overrides config
	t.Setenv("SYNAPSE_LANG", "en")
	if lang := cfg.ResolveLanguage(""); lang != "en" {
		t.Fatalf("expected en from env, got %s", lang)
	}
}
