package cmd

import (
	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/output"
)

// addPermissionsCmd adds a "permissions <id>" subcommand.
func addPermissionsCmd(parent *cobra.Command, apiPath string) {
	parent.AddCommand(&cobra.Command{
		Use:   "permissions <id>",
		Short: "Show user permissions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}
			path := apiPath + args[0] + "/permissions/"
			outputFlag, _ := cmd.Flags().GetString("output")
			return fetchAndFormat(sc, path, output.DetectFormat(outputFlag), nil, cmd.OutOrStdout())
		},
	})
}

// addRolesCmd adds a "roles <id>" subcommand.
func addRolesCmd(parent *cobra.Command, apiPath string) {
	parent.AddCommand(&cobra.Command{
		Use:   "roles <id>",
		Short: "List roles",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}
			path := apiPath + args[0] + "/roles/"
			outputFlag, _ := cmd.Flags().GetString("output")
			return fetchAndFormat(sc, path, output.DetectFormat(outputFlag), nil, cmd.OutOrStdout())
		},
	})
}

// addInviteCmd adds an "invite <id>" subcommand.
func addInviteCmd(parent *cobra.Command, apiPath string) {
	parent.AddCommand(&cobra.Command{
		Use:   "invite <id>",
		Short: "Invite a member",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := buildClient(cmd)
			if err != nil {
				return err
			}
			var jsonBody string
			jsonBody, _ = cmd.Flags().GetString("json")
			if jsonBody == "" {
				jsonBody = "{}"
			}
			path := apiPath + args[0] + "/invite/"
			outputFlag, _ := cmd.Flags().GetString("output")
			_ = sc // TODO: POST invite with body
			return fetchAndFormat(sc, path, output.DetectFormat(outputFlag), nil, cmd.OutOrStdout())
		},
	})
}
