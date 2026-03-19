package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/config"
	"github.com/datamaker-kr/synapse-cli/internal/output"
	"github.com/datamaker-kr/synapse-cli/internal/validation"
)

func newAPICmd() *cobra.Command {
	var data string

	cmd := &cobra.Command{
		Use:   "api <METHOD> <PATH>",
		Short: "Make an arbitrary API request (escape hatch)",
		Long:  "Invoke any Synapse API endpoint. Auth headers are injected automatically from the current context.",
		Example: `  synapse api GET /v2/projects/
  synapse api POST /v2/projects/ --data '{"title":"New","category":"image","configuration":{}}'
  echo '{"title":"Test"}' | synapse api POST /v2/projects/`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			path := args[1]

			// Input validation
			if err := validation.ValidateAPIPath(path); err != nil {
				return err
			}

			// Dry-run (client-side preview)
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun && isMutation(method) {
				cfg, _ := config.LoadConfig()
				ctxCfg, _ := cfg.ActiveContext()
				fmt.Fprintf(os.Stderr, "[DRY RUN] Would %s %s%s\n", method, ctxCfg.Server, path)
				fmt.Fprintf(os.Stderr, "Headers: Authorization: Token %s, SYNAPSE-Tenant: %s\n",
					config.MaskToken(ctxCfg.Token), ctxCfg.TenantCode)
				if data != "" {
					fmt.Fprintf(os.Stderr, "Body: %s\n", data)
				}
				return nil
			}

			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}

			// Read body from --data flag or stdin
			var bodyReader io.Reader
			if data != "" {
				bodyReader = strings.NewReader(data)
			} else if !isatty.IsTerminal(os.Stdin.Fd()) {
				bodyReader = os.Stdin
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, method, path, bodyReader)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Verbose output
			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				fmt.Fprintf(os.Stderr, "< %s %s\n", resp.Proto, resp.Status)
				for k, vs := range resp.Header {
					for _, v := range vs {
						fmt.Fprintf(os.Stderr, "< %s: %s\n", k, v)
					}
				}
				fmt.Fprintln(os.Stderr)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			if resp.StatusCode >= 400 {
				apiErr := &client.APIError{StatusCode: resp.StatusCode, RawBody: body}
				_ = json.Unmarshal(body, apiErr)
				outputFlag, _ := cmd.Flags().GetString("output")
				if output.DetectFormat(outputFlag) == "json" {
					fmt.Fprintln(cmd.OutOrStdout(), string(body))
				} else {
					fmt.Fprint(cmd.OutOrStderr(), apiErr.FormatHuman())
				}
				return fmt.Errorf("HTTP %d", resp.StatusCode)
			}

			// Output response
			outputFlag, _ := cmd.Flags().GetString("output")
			format := output.DetectFormat(outputFlag)
			if format == "json" || format == "ndjson" {
				// Pass through raw JSON
				fmt.Fprintln(cmd.OutOrStdout(), string(body))
			} else {
				// Try to pretty-print
				var parsed interface{}
				if err := json.Unmarshal(body, &parsed); err == nil {
					f := output.NewFormatter(format, cmd.OutOrStdout())
					return f.Format(parsed)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(body))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&data, "data", "d", "", "Request body (JSON)")

	return cmd
}

func isMutation(method string) bool {
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	}
	return false
}
