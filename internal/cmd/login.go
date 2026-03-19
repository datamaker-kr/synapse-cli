package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login with email and password",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			ctxCfg, err := cfg.ActiveContext()
			if err != nil {
				return fmt.Errorf("no server configured.\n  synapse config add-context <name> --server <url>")
			}

			// Prompt for credentials
			var email, password string
			fmt.Fprint(os.Stderr, "Email: ")
			if _, err := fmt.Scanln(&email); err != nil {
				return fmt.Errorf("read email: %w", err)
			}

			fmt.Fprint(os.Stderr, "Password: ")
			pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			password = string(pwBytes)

			// Call login API
			sc, err := client.NewSynapseClient(ctxCfg, cfg.ResolveLanguage(""))
			if err != nil {
				return err
			}

			body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, http.MethodPost, "/users/login/", strings.NewReader(body))
			if err != nil {
				return fmt.Errorf("login request: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				apiErr := client.ParseAPIError(resp)
				return fmt.Errorf("login failed: %s", apiErr.Error())
			}

			var result struct {
				Token string `json:"token"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("parse login response: %w", err)
			}

			// Save token
			ctxCfg.Token = result.Token
			ctxCfg.AuthMethod = "token"
			cfg.Contexts[cfg.CurrentContext] = *ctxCfg
			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Login successful! (context: %s)\n", cfg.CurrentContext)

			// Show tenant list as onboarding hint
			showTenantHint(ctxCfg, result.Token, cfg.ResolveLanguage(""))

			return nil
		},
	}
}

func showTenantHint(ctxCfg *config.ContextConfig, token, lang string) {
	tempCfg := *ctxCfg
	tempCfg.Token = token
	sc, err := client.NewSynapseClient(&tempCfg, lang)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := sc.RawRequest(ctx, http.MethodGet, "/v2/tenants/", nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	var tenantResp struct {
		Data []struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tenantResp); err != nil || len(tenantResp.Data) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "\nWorkspaces:")
	for _, t := range tenantResp.Data {
		fmt.Fprintf(os.Stderr, "  %s\t%s\n", t.Code, t.Name)
	}
	fmt.Fprintln(os.Stderr, "\nSelect a workspace:")
	fmt.Fprintln(os.Stderr, "  synapse tenant select <code>")
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear local credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if cfg.CurrentContext == "" {
				return fmt.Errorf("no context configured")
			}

			ctxCfg, ok := cfg.Contexts[cfg.CurrentContext]
			if !ok {
				return fmt.Errorf("context %q not found", cfg.CurrentContext)
			}

			ctxCfg.Token = ""
			ctxCfg.AccessToken = ""
			ctxCfg.TenantCode = ""
			ctxCfg.AuthMethod = ""
			cfg.Contexts[cfg.CurrentContext] = ctxCfg

			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Logged out from context %q\n", cfg.CurrentContext)
			return nil
		},
	}
}
