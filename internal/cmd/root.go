package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/datamaker-kr/synapse-cli/internal/config"
	"github.com/datamaker-kr/synapse-cli/internal/i18n"
)

var version = "dev"

// SetVersion sets the CLI version (called from main).
func SetVersion(v string) {
	version = v
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "synapse",
		Short:         "Synapse CLI — ML platform command-line interface",
		Long:          "synapse-cli is a command-line tool for interacting with the Synapse ML platform API.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	root.PersistentFlags().StringP("output", "o", "", "Output format (table|json|yaml|ndjson)")
	root.PersistentFlags().String("context", "", "Override active context")
	root.PersistentFlags().String("server", "", "Override server URL")
	root.PersistentFlags().String("token", "", "Override auth token")
	root.PersistentFlags().String("tenant", "", "Override tenant code")
	root.PersistentFlags().BoolP("verbose", "v", false, "Verbose output (include HTTP request/response)")
	root.PersistentFlags().Bool("dry-run", false, "Dry run mode (validate without executing)")
	root.PersistentFlags().Bool("skip-health-check", false, "Skip automatic health check on startup")
	root.PersistentFlags().Bool("no-logo", false, "Hide Synapse logo on startup")
	root.PersistentFlags().String("lang", "", "Language (en|ko)")

	// Version template
	root.SetVersionTemplate(fmt.Sprintf("synapse-cli version %s\n", version))

	// Entry point validation
	root.PersistentPreRunE = preRunCheck

	// Register commands
	root.AddCommand(newConfigCmd())
	root.AddCommand(newLoginCmd())
	root.AddCommand(newLogoutCmd())
	root.AddCommand(newTenantCmd())
	root.AddCommand(newTokenCmd())
	root.AddCommand(newHealthCmd())
	root.AddCommand(newAPICmd())

	// v2 API resource commands (17 resources)
	registerResourceCommands(root)

	return root
}

// Execute runs the root command.
func Execute() error {
	root := newRootCmd()

	// Initialize i18n before execution (need to parse --lang early)
	initI18N(root)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func initI18N(root *cobra.Command) {
	// Resolve language: --lang flag > env > config > OS locale > "en"
	cfg, _ := config.LoadConfig()
	langFlag := ""
	// Parse --lang from os.Args early (before cobra parses)
	for i, arg := range os.Args {
		if arg == "--lang" && i+1 < len(os.Args) {
			langFlag = os.Args[i+1]
			break
		}
	}
	lang := "en"
	if cfg != nil {
		lang = cfg.ResolveLanguage(langFlag)
	}
	i18n.Init(lang)
}
