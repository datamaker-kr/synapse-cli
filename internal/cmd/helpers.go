package cmd

import (
	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
)

// buildClient creates a SynapseClient from the current config and flags.
func buildClient(cmd *cobra.Command) (*client.SynapseClient, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	ctxCfg, err := cfg.ActiveContext()
	if err != nil {
		return nil, err
	}

	// Apply flag overrides
	if v, _ := cmd.Flags().GetString("server"); v != "" {
		ctxCfg.Server = v
	}
	if v, _ := cmd.Flags().GetString("token"); v != "" {
		ctxCfg.Token = v
		ctxCfg.AuthMethod = "token"
	}
	if v, _ := cmd.Flags().GetString("tenant"); v != "" {
		ctxCfg.TenantCode = v
	}

	langFlag, _ := cmd.Flags().GetString("lang")
	lang := cfg.ResolveLanguage(langFlag)

	return client.NewSynapseClient(ctxCfg, lang)
}

// runWithClient is a helper that creates a SynapseClient and passes it to the handler.
type clientRunE func(cmd *cobra.Command, args []string, sc *client.SynapseClient) error

func runWithClient(fn clientRunE) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		sc, err := buildClient(cmd)
		if err != nil {
			return err
		}
		return fn(cmd, args, sc)
	}
}
