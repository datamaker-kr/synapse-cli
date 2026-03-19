package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/branding"
	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
)

// ValidationLevel defines how much pre-validation a command requires.
type ValidationLevel int

const (
	ValidationNone   ValidationLevel = 0 // config, version, completion
	ValidationServer ValidationLevel = 1 // health, login
	ValidationAuth   ValidationLevel = 2 // tenant list/select/get
	ValidationFull   ValidationLevel = 3 // all other API commands
)

func preRunCheck(cmd *cobra.Command, _ []string) error {
	// Logo
	noLogo, _ := cmd.Flags().GetBool("no-logo")
	if !noLogo && os.Getenv("SYNAPSE_NO_LOGO") == "" && isatty.IsTerminal(os.Stderr.Fd()) {
		branding.PrintLogo(os.Stderr, version, true)
	}

	level := getValidationLevel(cmd)
	if level == ValidationNone {
		return nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Level 1+: server must be configured
	if cfg.CurrentContext == "" {
		return fmt.Errorf("server is not configured.\n  synapse config add-context <name> --server <url>")
	}

	ctxCfg, err := cfg.ActiveContext()
	if err != nil || ctxCfg.Server == "" {
		return fmt.Errorf("server URL is not set.\n  synapse config set-server <url>")
	}

	// Level 2+: authentication required
	if level >= ValidationAuth {
		if ctxCfg.Token == "" && ctxCfg.AccessToken == "" {
			return fmt.Errorf("authentication required.\n  synapse login")
		}
	}

	// Level 3: tenant + health check
	if level >= ValidationFull {
		if ctxCfg.TenantCode == "" && ctxCfg.AuthMethod != "access_token" {
			return fmt.Errorf("workspace not selected.\n  synapse tenant list\n  synapse tenant select <code>")
		}

		skipHealth, _ := cmd.Flags().GetBool("skip-health-check")
		if !skipHealth {
			hctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := client.HealthCheck(hctx, ctxCfg.Server); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: cannot connect to server: %s (context: %s)\n", ctxCfg.Server, cfg.CurrentContext)
			}
		}
	}

	return nil
}

func getValidationLevel(cmd *cobra.Command) ValidationLevel {
	path := commandPath(cmd)

	switch {
	case matchPath(path, "config", "version", "completion"):
		return ValidationNone
	case matchPath(path, "health", "login"):
		return ValidationServer
	case matchPath(path, "tenant list", "tenant select", "tenant get"):
		return ValidationAuth
	default:
		return ValidationFull
	}
}

func commandPath(cmd *cobra.Command) string {
	parts := []string{}
	for c := cmd; c != nil && c.Name() != "synapse" && c.Name() != ""; c = c.Parent() {
		parts = append([]string{c.Name()}, parts...)
	}
	return strings.Join(parts, " ")
}

func matchPath(path string, targets ...string) bool {
	for _, t := range targets {
		if path == t || strings.HasPrefix(path, t+" ") {
			return true
		}
	}
	return false
}
