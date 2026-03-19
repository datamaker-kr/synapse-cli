package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level CLI configuration.
type Config struct {
	CurrentContext string                   `yaml:"current_context"`
	Language       string                   `yaml:"language,omitempty"`
	Contexts       map[string]ContextConfig `yaml:"contexts"`
}

// ContextConfig represents a single environment profile.
type ContextConfig struct {
	Server      string `yaml:"server"`
	Environment string `yaml:"environment,omitempty"`
	AuthMethod  string `yaml:"auth_method,omitempty"`
	Token       string `yaml:"token,omitempty"`
	TenantCode  string `yaml:"tenant_code,omitempty"`
	AccessToken string `yaml:"access_token,omitempty"`
}

// ConfigDir returns the configuration directory path.
// Priority: SYNAPSE_CONFIG_DIR > os.UserConfigDir()/synapse > ~/.synapse
func ConfigDir() string {
	if dir := os.Getenv("SYNAPSE_CONFIG_DIR"); dir != "" {
		return dir
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "synapse")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".synapse")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// LoadConfig reads the config file from disk.
// Returns an empty config (not an error) if the file does not exist.
func LoadConfig() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Contexts: make(map[string]ContextConfig)}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]ContextConfig)
	}
	return &cfg, nil
}

// Save writes the config to disk with 0600 permissions.
func (c *Config) Save() error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	path := ConfigPath()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// ActiveContext returns the resolved context config after applying
// environment variable overrides.
func (c *Config) ActiveContext() (*ContextConfig, error) {
	name := c.CurrentContext
	if v := os.Getenv("SYNAPSE_CONTEXT"); v != "" {
		name = v
	}
	if name == "" {
		return nil, fmt.Errorf("no context configured")
	}

	ctx, ok := c.Contexts[name]
	if !ok {
		return nil, fmt.Errorf("context %q not found", name)
	}

	// Environment variable overrides
	if v := os.Getenv("SYNAPSE_SERVER"); v != "" {
		ctx.Server = v
	}
	if v := os.Getenv("SYNAPSE_TOKEN"); v != "" {
		ctx.Token = v
		ctx.AuthMethod = "token"
	}
	if v := os.Getenv("SYNAPSE_TENANT"); v != "" {
		ctx.TenantCode = v
	}
	if v := os.Getenv("SYNAPSE_ACCESS_TOKEN"); v != "" {
		ctx.AccessToken = v
		ctx.AuthMethod = "access_token"
	}

	return &ctx, nil
}

// SetContext switches the active context.
func (c *Config) SetContext(name string) error {
	if _, ok := c.Contexts[name]; !ok {
		return fmt.Errorf("context %q not found", name)
	}
	c.CurrentContext = name
	return nil
}

// AddContext adds a new context. Returns error if it already exists.
func (c *Config) AddContext(name string, ctx ContextConfig) error {
	if _, ok := c.Contexts[name]; ok {
		return fmt.Errorf("context %q already exists", name)
	}
	c.Contexts[name] = ctx
	if c.CurrentContext == "" {
		c.CurrentContext = name
	}
	return nil
}

// DeleteContext removes a context. Returns error if it's the active context
// and force is false.
func (c *Config) DeleteContext(name string, force bool) error {
	if _, ok := c.Contexts[name]; !ok {
		return fmt.Errorf("context %q not found", name)
	}
	if c.CurrentContext == name && !force {
		return fmt.Errorf("cannot delete active context %q (use --force)", name)
	}
	delete(c.Contexts, name)
	if c.CurrentContext == name {
		c.CurrentContext = ""
	}
	return nil
}

// MaskToken masks a token string, showing only the last 4 characters.
func MaskToken(token string) string {
	if len(token) <= 4 {
		return "****"
	}
	return "***..." + token[len(token)-4:]
}

// ResolveLanguage returns the effective language based on priority:
// flag > env > config > OS locale > "en"
func (c *Config) ResolveLanguage(flagLang string) string {
	if flagLang != "" {
		return flagLang
	}
	if v := os.Getenv("SYNAPSE_LANG"); v != "" {
		return v
	}
	if c.Language != "" {
		return c.Language
	}
	if lang := os.Getenv("LANG"); strings.HasPrefix(lang, "ko") {
		return "ko"
	}
	return "en"
}
