package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
	"github.com/datamaker-kr/synapse-cli/internal/output"
)

func newTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage access tokens",
	}

	cmd.AddCommand(newTokenListCmd())
	cmd.AddCommand(newTokenCreateCmd())
	cmd.AddCommand(newTokenGetCmd())
	cmd.AddCommand(newTokenDeleteCmd())

	return cmd
}

func newTokenListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List access tokens",
		RunE:  runWithClient(tokenListRun),
	}
}

func tokenListRun(cmd *cobra.Command, _ []string, sc *client.SynapseClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := sc.RawRequest(ctx, http.MethodGet, "/v2/tokens/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", client.ParseAPIError(resp).FormatHuman())
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
			{Header: "ID", Field: "id"},
			{Header: "Description", Field: "description"},
			{Header: "Last Used", Field: "last_used"},
		})
	}

	f := output.NewFormatter(format, cmd.OutOrStdout())
	var data interface{}
	_ = json.Unmarshal(body, &data)
	return f.Format(data)
}

func newTokenCreateCmd() *cobra.Command {
	var description string
	var setConfig bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}

			body := fmt.Sprintf(`{"description":%q}`, description)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, http.MethodPost, "/v2/tokens/", strings.NewReader(body))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				return fmt.Errorf("%s", client.ParseAPIError(resp).FormatHuman())
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return err
			}

			outputFlag, _ := cmd.Flags().GetString("output")
			f := output.NewFormatter(output.DetectFormat(outputFlag), cmd.OutOrStdout())
			if err := f.Format(result); err != nil {
				return err
			}

			fmt.Fprintln(os.Stderr, "\nNote: Token value is only shown once. Save it securely.")

			if setConfig {
				if token, ok := result["token"].(string); ok && token != "" {
					cfg, err := config.LoadConfig()
					if err != nil {
						return err
					}
					ctxCfg, _ := cfg.ActiveContext()
					ctxCfg.AccessToken = token
					ctxCfg.AuthMethod = "access_token"
					cfg.Contexts[cfg.CurrentContext] = *ctxCfg
					if err := cfg.Save(); err != nil {
						return err
					}
					fmt.Fprintf(os.Stderr, "Token saved to context %q\n", cfg.CurrentContext)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Token description")
	cmd.Flags().BoolVar(&setConfig, "set-config", false, "Save token to current context")

	return cmd
}

func newTokenGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get access token details (token value not included)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, http.MethodGet, "/v2/tokens/"+args[0]+"/", nil)
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

func newTokenDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an access token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Fprintf(os.Stderr, "Delete token %s? This cannot be undone. Use --force to confirm.\n", args[0])
				return fmt.Errorf("use --force to confirm deletion")
			}

			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, http.MethodDelete, "/v2/tokens/"+args[0]+"/", nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
				return fmt.Errorf("%s", client.ParseAPIError(resp).FormatHuman())
			}

			fmt.Fprintf(os.Stderr, "Token %s deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion")
	return cmd
}
