package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
)

func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check server health",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			ctxCfg, err := cfg.ActiveContext()
			if err != nil {
				return fmt.Errorf("no server configured.\n  synapse config add-context <name> --server <url>")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Server:  %s\n", ctxCfg.Server)
			fmt.Fprintf(cmd.OutOrStdout(), "Context: %s\n", cfg.CurrentContext)

			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := client.HealthCheck(ctx, ctxCfg.Server); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Status:  UNREACHABLE (%v)\n", err)
				return fmt.Errorf("health check failed")
			}

			elapsed := time.Since(start)
			fmt.Fprintf(cmd.OutOrStdout(), "Status:  OK (%s)\n", elapsed.Round(time.Millisecond))
			return nil
		},
	}
}
