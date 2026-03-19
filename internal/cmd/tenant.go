package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
	"github.com/datamaker-kr/synapse-cli/internal/output"
)

func newTenantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Manage workspaces (tenants)",
	}

	cmd.AddCommand(newTenantListCmd())
	cmd.AddCommand(newTenantGetCmd())
	cmd.AddCommand(newTenantSelectCmd())

	return cmd
}

func newTenantListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List workspaces you belong to",
		RunE:  runWithClient(tenantListRun),
	}
}

func tenantListRun(cmd *cobra.Command, _ []string, sc *client.SynapseClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := sc.RawRequest(ctx, http.MethodGet, "/v2/tenants/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		apiErr := client.ParseAPIError(resp)
		return fmt.Errorf("%s", apiErr.FormatHuman())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	outputFlag, _ := cmd.Flags().GetString("output")
	format := output.DetectFormat(outputFlag)

	if format == "table" {
		pr, err := client.ParsePaginatedResponse(body)
		if err != nil {
			return err
		}
		var items []map[string]interface{}
		if err := json.Unmarshal(pr.Data, &items); err != nil {
			return err
		}
		f := output.NewFormatter("table", cmd.OutOrStdout())
		return f.FormatList(items, []output.Column{
			{Header: "Code", Field: "code"},
			{Header: "Name", Field: "name"},
		})
	}

	f := output.NewFormatter(format, cmd.OutOrStdout())
	var data interface{}
	_ = json.Unmarshal(body, &data)
	return f.Format(data)
}

func newTenantGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <code>",
		Short: "Get workspace details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, http.MethodGet, "/v2/tenants/"+args[0]+"/", nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("%s", client.ParseAPIError(resp).FormatHuman())
			}

			outputFlag, _ := cmd.Flags().GetString("output")
			f := output.NewFormatter(output.DetectFormat(outputFlag), cmd.OutOrStdout())
			var data interface{}
			_ = json.NewDecoder(resp.Body).Decode(&data)
			return f.Format(data)
		},
	}
}

func newTenantSelectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "select <code>",
		Short: "Select active workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code := args[0]

			// Verify tenant exists via API
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, http.MethodGet, "/v2/tenants/"+code+"/", nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("workspace %q not found", code)
			}

			// Save to config
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			ctxCfg, ok := cfg.Contexts[cfg.CurrentContext]
			if !ok {
				return fmt.Errorf("no active context")
			}
			ctxCfg.TenantCode = code
			cfg.Contexts[cfg.CurrentContext] = ctxCfg
			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Workspace set to %q (context: %s)\n", code, cfg.CurrentContext)
			return nil
		},
	}
}
