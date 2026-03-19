package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
	"github.com/datamaker-kr/synapse-cli/internal/validation"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration and contexts",
	}

	cmd.AddCommand(newConfigAddContextCmd())
	cmd.AddCommand(newConfigUseContextCmd())
	cmd.AddCommand(newConfigListContextsCmd())
	cmd.AddCommand(newConfigDeleteContextCmd())
	cmd.AddCommand(newConfigSetServerCmd())
	cmd.AddCommand(newConfigSetTokenCmd())
	cmd.AddCommand(newConfigSetLanguageCmd())
	cmd.AddCommand(newConfigCurrentContextCmd())
	cmd.AddCommand(newConfigViewCmd())

	return cmd
}

func newConfigAddContextCmd() *cobra.Command {
	var server string
	var force bool

	cmd := &cobra.Command{
		Use:   "add-context <name>",
		Short: "Add a new environment context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := validation.ValidateServerURL(server); err != nil {
				return err
			}

			// Health check the server
			if !force {
				fmt.Fprintf(os.Stderr, "Checking server connectivity: %s\n", server)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := client.HealthCheck(ctx, server); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Cannot connect to server: %s\n", server)
					fmt.Fprintf(os.Stderr, "Use --force to save anyway.\n")
					return fmt.Errorf("health check failed: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Server OK: %s\n", server)
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			if err := cfg.AddContext(name, config.ContextConfig{Server: server}); err != nil {
				return err
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Context %q added (server: %s)\n", name, server)
			if cfg.CurrentContext == name {
				fmt.Fprintf(os.Stderr, "Switched to context %q\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&server, "server", "", "Server URL (required)")
	_ = cmd.MarkFlagRequired("server")
	cmd.Flags().BoolVar(&force, "force", false, "Save even if health check fails")

	return cmd
}

func newConfigUseContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use-context <name>",
		Short: "Switch active context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if err := cfg.SetContext(args[0]); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Switched to context %q\n", args[0])
			return nil
		},
	}
}

func newConfigListContextsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-contexts",
		Short: "List all configured contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if len(cfg.Contexts) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No contexts configured. Run: synapse config add-context <name> --server <url>")
				return nil
			}
			for name, ctx := range cfg.Contexts {
				marker := "  "
				if name == cfg.CurrentContext {
					marker = "* "
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s\t%s\n", marker, name, ctx.Server)
			}
			return nil
		},
	}
}

func newConfigDeleteContextCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete-context <name>",
		Short: "Delete a context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if err := cfg.DeleteContext(args[0], force); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Context %q deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force delete even if active context")
	return cmd
}

func newConfigSetServerCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "set-server <url>",
		Short: "Set server URL for current context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			server := args[0]
			if err := validation.ValidateServerURL(server); err != nil {
				return err
			}

			if !force {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := client.HealthCheck(ctx, server); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Cannot connect to server: %s\n", server)
					fmt.Fprintf(os.Stderr, "Use --force to save anyway.\n")
					return fmt.Errorf("health check failed: %w", err)
				}
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			ctxCfg, err := cfg.ActiveContext()
			if err != nil {
				return fmt.Errorf("no active context: %w", err)
			}
			ctxCfg.Server = server
			cfg.Contexts[cfg.CurrentContext] = *ctxCfg
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Server set to %s\n", server)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Save even if health check fails")
	return cmd
}

func newConfigSetTokenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-token <token>",
		Short: "Set access token for current context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			ctxCfg, err := cfg.ActiveContext()
			if err != nil {
				return fmt.Errorf("no active context: %w", err)
			}
			ctxCfg.AccessToken = args[0]
			ctxCfg.AuthMethod = "access_token"
			cfg.Contexts[cfg.CurrentContext] = *ctxCfg
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Access token set for context %q\n", cfg.CurrentContext)
			return nil
		},
	}
}

func newConfigSetLanguageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-language <en|ko>",
		Short: "Set CLI display language",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			lang := args[0]
			if lang != "en" && lang != "ko" {
				return fmt.Errorf("unsupported language %q: use 'en' or 'ko'", lang)
			}
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			cfg.Language = lang
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Language set to %q\n", lang)
			return nil
		},
	}
}

func newConfigCurrentContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current-context",
		Short: "Show current active context name",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if cfg.CurrentContext == "" {
				return fmt.Errorf("no context configured")
			}
			fmt.Fprintln(cmd.OutOrStdout(), cfg.CurrentContext)
			return nil
		},
	}
}

func newConfigViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Config: %s\n", config.ConfigPath())
			fmt.Fprintf(w, "Context: %s\n", cfg.CurrentContext)
			fmt.Fprintf(w, "Language: %s\n", cfg.ResolveLanguage(""))

			if cfg.CurrentContext == "" {
				return nil
			}

			ctx, ok := cfg.Contexts[cfg.CurrentContext]
			if !ok {
				return nil
			}

			fmt.Fprintln(w, "")
			fmt.Fprintf(w, "  Server:      %s%s\n", ctx.Server, envSource("SYNAPSE_SERVER"))
			fmt.Fprintf(w, "  Auth Method: %s\n", ctx.AuthMethod)

			if ctx.Token != "" {
				fmt.Fprintf(w, "  Token:       %s%s\n", config.MaskToken(ctx.Token), envSource("SYNAPSE_TOKEN"))
			}
			if ctx.AccessToken != "" {
				fmt.Fprintf(w, "  Access Token: %s%s\n", config.MaskToken(ctx.AccessToken), envSource("SYNAPSE_ACCESS_TOKEN"))
			}
			if ctx.TenantCode != "" {
				fmt.Fprintf(w, "  Tenant:      %s%s\n", ctx.TenantCode, envSource("SYNAPSE_TENANT"))
			}

			return nil
		},
	}
}

func envSource(envVar string) string {
	if os.Getenv(envVar) != "" {
		return fmt.Sprintf("  (env: %s)", envVar)
	}
	return ""
}
