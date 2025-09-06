package config

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/analysisconfig"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

// ConfigOptions holds options for the config command
type ConfigOptions struct {
	IO *iostreams.IOStreams
}

// NewCmdConfig creates a new config command
func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	opts := &ConfigOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage analysis configuration",
		Long: heredoc.Doc(`
			Manage configuration settings for the analysis commands.
			
			Configuration is stored in ~/.ghccli/analysis.yaml and includes
			settings like default organization for repository searches.
		`),
	}

	cmd.AddCommand(newSetDefaultOrgCmd(opts))
	cmd.AddCommand(newShowConfigCmd(opts))

	return cmd
}

// newSetDefaultOrgCmd creates the set-default-org subcommand
func newSetDefaultOrgCmd(opts *ConfigOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "set-default-org <organization>",
		Short: "Set the default organization for repository searches",
		Long: heredoc.Doc(`
			Set the default organization that will be used when searching for
			repositories by name only (without owner prefix).
			
			This setting is persisted to ~/.ghccli/analysis.yaml.
		`),
		Example: heredoc.Doc(`
			# Set default organization
			$ gh analysis config set-default-org myorg
			
			# Now you can use repository names without org prefix
			$ gh analysis patterns myrepo --patterns ports
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			org := args[0]
			
			config, err := analysisconfig.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if err := config.SetDefaultOrg(org); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Fprintf(opts.IO.Out, "Default organization set to: %s\n", org)
			fmt.Fprintf(opts.IO.Out, "Configuration saved to: %s\n", config.ConfigPath())
			return nil
		},
	}
}

// newShowConfigCmd creates the show subcommand
func newShowConfigCmd(opts *ConfigOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "show",
		Aliases: []string{"list", "view"},
		Short:   "Show current analysis configuration",
		Long: heredoc.Doc(`
			Display the current analysis configuration including default
			organization and other settings.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := analysisconfig.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Fprintf(opts.IO.Out, "Analysis Configuration:\n\n")
			fmt.Fprintf(opts.IO.Out, "Config file: %s\n", config.ConfigPath())
			
			if config.DefaultOrg != "" {
				fmt.Fprintf(opts.IO.Out, "Default organization: %s\n", config.DefaultOrg)
			} else {
				fmt.Fprintf(opts.IO.Out, "Default organization: (not set)\n")
			}
			
			fmt.Fprintf(opts.IO.Out, "Output format: %s\n", config.OutputFormat)
			
			if len(config.SearchPatterns) > 0 {
				fmt.Fprintf(opts.IO.Out, "\nCustom search patterns:\n")
				for name, pattern := range config.SearchPatterns {
					fmt.Fprintf(opts.IO.Out, "  %s: %s\n", name, pattern)
				}
			}

			return nil
		},
	}
}