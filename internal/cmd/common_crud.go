package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/client"
	"github.com/datamaker-kr/synapse-cli/internal/output"
)

// ResourceDef defines a v2 API resource for CRUD command generation.
type ResourceDef struct {
	Name      string // e.g., "project"
	Plural    string // e.g., "projects" (URL path)
	APIPath   string // e.g., "/v2/projects/"
	IDField   string // e.g., "id" or "code"
	ListCols  []output.Column
	HasCreate bool
	HasUpdate bool
	HasDelete bool
}

func newResourceCmd(def ResourceDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:   def.Name,
		Short: fmt.Sprintf("Manage %s", def.Plural),
	}

	cmd.AddCommand(newResourceListCmd(def))
	cmd.AddCommand(newResourceGetCmd(def))

	if def.HasCreate {
		cmd.AddCommand(newResourceCreateCmd(def))
	}
	if def.HasUpdate {
		cmd.AddCommand(newResourceUpdateCmd(def))
	}
	if def.HasDelete {
		cmd.AddCommand(newResourceDeleteCmd(def))
	}

	return cmd
}

func newResourceListCmd(def ResourceDef) *cobra.Command {
	var perPage int
	var cursor string
	var pageAll bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s", def.Plural),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}

			path := def.APIPath
			params := []string{}
			if perPage > 0 {
				params = append(params, fmt.Sprintf("per_page=%d", perPage))
			}
			if cursor != "" {
				params = append(params, fmt.Sprintf("cursor=%s", cursor))
			}
			if len(params) > 0 {
				path += "?" + strings.Join(params, "&")
			}

			outputFlag, _ := cmd.Flags().GetString("output")
			format := output.DetectFormat(outputFlag)

			if pageAll {
				return streamAllPages(sc, path, format, cmd.OutOrStdout())
			}

			return fetchAndFormat(sc, path, format, def.ListCols, cmd.OutOrStdout())
		},
	}

	cmd.Flags().IntVar(&perPage, "per-page", 0, "Results per page")
	cmd.Flags().StringVar(&cursor, "cursor", "", "Pagination cursor")
	cmd.Flags().BoolVar(&pageAll, "page-all", false, "Fetch all pages")

	return cmd
}

func newResourceGetCmd(def ResourceDef) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("get <%s>", def.IDField),
		Short: fmt.Sprintf("Get %s details", def.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}

			path := def.APIPath + args[0] + "/"
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, "GET", path, nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
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

func newResourceCreateCmd(def ResourceDef) *cobra.Command {
	var jsonBody string

	cmd := &cobra.Command{
		Use:   "create",
		Short: fmt.Sprintf("Create a %s", def.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonBody == "" {
				return fmt.Errorf("--json is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}

			path := def.APIPath
			if dryRun {
				path += "?dry_run=true"
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, "POST", path, strings.NewReader(jsonBody))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if dryRun {
				fmt.Fprintf(os.Stderr, "[DRY RUN] POST %s\n", def.APIPath)
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("%s", client.ParseAPIError(resp).FormatHuman())
			}

			outputFlag, _ := cmd.Flags().GetString("output")
			f := output.NewFormatter(output.DetectFormat(outputFlag), cmd.OutOrStdout())
			var data interface{}
			_ = json.NewDecoder(resp.Body).Decode(&data)
			return f.Format(data)
		},
	}

	cmd.Flags().StringVar(&jsonBody, "json", "", "Request body (JSON)")
	return cmd
}

func newResourceUpdateCmd(def ResourceDef) *cobra.Command {
	var jsonBody string

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("update <%s>", def.IDField),
		Short: fmt.Sprintf("Update a %s", def.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonBody == "" {
				return fmt.Errorf("--json is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}

			path := def.APIPath + args[0] + "/"
			if dryRun {
				path += "?dry_run=true"
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, "PATCH", path, strings.NewReader(jsonBody))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if dryRun {
				fmt.Fprintf(os.Stderr, "[DRY RUN] PATCH %s%s/\n", def.APIPath, args[0])
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("%s", client.ParseAPIError(resp).FormatHuman())
			}

			outputFlag, _ := cmd.Flags().GetString("output")
			f := output.NewFormatter(output.DetectFormat(outputFlag), cmd.OutOrStdout())
			var data interface{}
			_ = json.NewDecoder(resp.Body).Decode(&data)
			return f.Format(data)
		},
	}

	cmd.Flags().StringVar(&jsonBody, "json", "", "Request body (JSON)")
	return cmd
}

func newResourceDeleteCmd(def ResourceDef) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("delete <%s>", def.IDField),
		Short: fmt.Sprintf("Delete a %s", def.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return fmt.Errorf("use --force to confirm deletion of %s %s", def.Name, args[0])
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}

			path := def.APIPath + args[0] + "/"
			if dryRun {
				path += "?dry_run=true"
				fmt.Fprintf(os.Stderr, "[DRY RUN] DELETE %s%s/\n", def.APIPath, args[0])
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resp, err := sc.RawRequest(ctx, "DELETE", path, nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				return fmt.Errorf("%s", client.ParseAPIError(resp).FormatHuman())
			}

			if !dryRun {
				fmt.Fprintf(os.Stderr, "%s %s deleted\n", def.Name, args[0])
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion")
	return cmd
}

// --- helpers ---

func fetchAndFormat(sc *client.SynapseClient, path, format string, cols []output.Column, w io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := sc.RawRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s", client.ParseAPIError(resp).FormatHuman())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if format == "table" && len(cols) > 0 {
		pr, err := client.ParsePaginatedResponse(body)
		if err != nil {
			// Fallback: might not be paginated
			f := output.NewFormatter(format, w)
			var data interface{}
			_ = json.Unmarshal(body, &data)
			return f.Format(data)
		}
		var items []map[string]interface{}
		if err := json.Unmarshal(pr.Data, &items); err != nil {
			return err
		}
		f := output.NewFormatter("table", w)
		return f.FormatList(items, cols)
	}

	f := output.NewFormatter(format, w)
	var data interface{}
	_ = json.Unmarshal(body, &data)
	return f.Format(data)
}

func streamAllPages(sc *client.SynapseClient, basePath, format string, w io.Writer) error {
	path := basePath
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		resp, err := sc.RawRequest(ctx, "GET", path, nil)
		if err != nil {
			cancel()
			return err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}

		pr, err := client.ParsePaginatedResponse(body)
		if err != nil {
			return err
		}

		if format == "ndjson" {
			if err := client.StreamNDJSON(w, pr.Data); err != nil {
				return err
			}
		} else {
			var items []interface{}
			_ = json.Unmarshal(pr.Data, &items)
			f := output.NewFormatter(format, w)
			_ = f.Format(items)
		}

		if !pr.HasNextPage() {
			break
		}

		// Build next page URL
		sep := "?"
		if strings.Contains(basePath, "?") {
			sep = "&"
		}
		path = basePath + sep + "cursor=" + pr.NextCursor()
	}
	return nil
}
