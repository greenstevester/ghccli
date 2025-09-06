package analysis

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"

	configCmd "github.com/cli/cli/v2/pkg/cmd/analysis/config"
	patternsCmd "github.com/cli/cli/v2/pkg/cmd/analysis/patterns"
)

// NewCmdAnalysis creates the root analysis command
func NewCmdAnalysis(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analysis <command>",
		Short: "Analyze code patterns across repositories",
		Long: heredoc.Doc(`
			Analyze code patterns like ports, IP addresses, and security configurations
			across GitHub repositories with automatic authentication and repository selection.
			
			The analysis commands support:
			- Automatic login using GH_TOKEN environment variable
			- Smart repository selection with default organization support
			- Pattern-based analysis with built-in patterns
			- Key-value formatted output with line numbers and file locations
			
			Authentication:
			Set the GH_TOKEN environment variable for automatic authentication,
			or ensure you're logged in with 'gh auth login'.
		`),
		Example: heredoc.Doc(`
			# Analyze ports and IP addresses in a repository
			$ gh analysis patterns cli/cli --patterns ports,ipv4
			
			# Search for potential secrets in your default org
			$ gh analysis patterns myrepo --patterns secrets --format table
			
			# Set default organization for future searches
			$ gh analysis config set-default-org myorg
			
			# Show available patterns
			$ gh analysis patterns --help
		`),
	}

	cmd.AddCommand(patternsCmd.NewCmdPatterns(f, nil))
	cmd.AddCommand(configCmd.NewCmdConfig(f))
	cmd.AddCommand(newListPatternsCmd(f))

	return cmd
}

// newListPatternsCmd creates a command to list available patterns
func newListPatternsCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "list-patterns",
		Aliases: []string{"patterns-list", "show-patterns"},
		Short:   "List available analysis patterns",
		Long: heredoc.Doc(`
			Display all available built-in patterns that can be used
			for code analysis, including their descriptions and examples.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return formatPatternList(f.IOStreams)
		},
	}
}

// EnsureAuthentication is exported for use by subcommands
var EnsureAuthentication = ensureAuthentication

// formatPatternList formats the list of available patterns
func formatPatternList(io *iostreams.IOStreams) error {
	fmt.Fprintln(io.Out, "Available patterns:")
	fmt.Fprintln(io.Out)

	patterns := map[string]struct {
		Description string
		Examples    []string
	}{
		"ports": {
			"Network ports (1-65535)",
			[]string{`port: 8080`, `PORT="3000"`, `port=9090`},
		},
		"ipv4": {
			"IPv4 addresses",
			[]string{`host: "192.168.1.1"`, `IP=10.0.0.1`, `server: 127.0.0.1`},
		},
		"secrets": {
			"Potential secrets (keys, tokens, passwords)",
			[]string{`api_key: "abc123xyz789"`, `token="ghp_xxxxxxxxxxxx"`, `password: "secret123"`},
		},
		"urls": {
			"HTTP/HTTPS URLs",
			[]string{`url: "https://api.example.com"`, `endpoint=http://localhost:8080`},
		},
	}

	for name, pattern := range patterns {
		fmt.Fprintf(io.Out, "  %s\n", name)
		fmt.Fprintf(io.Out, "    %s\n", pattern.Description)
		fmt.Fprintf(io.Out, "    Examples:\n")
		for _, example := range pattern.Examples {
			fmt.Fprintf(io.Out, "      %s\n", example)
		}
		fmt.Fprintln(io.Out)
	}

	return nil
}