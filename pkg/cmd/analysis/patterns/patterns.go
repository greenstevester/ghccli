package patterns

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/analysisconfig"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/internal/prompter"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

// PatternsOptions holds options for the patterns command
type PatternsOptions struct {
	HttpClient   func() (*http.Client, error)
	Config       func() (gh.Config, error)
	IO           *iostreams.IOStreams

	Repository   string
	Patterns     []string
	OutputFormat string
	Recursive    bool
	FileTypes    []string
	Summary      bool
}

// NewCmdPatterns creates a new patterns command
func NewCmdPatterns(f *cmdutil.Factory, runF func(*PatternsOptions) error) *cobra.Command {
	opts := &PatternsOptions{
		HttpClient:   f.HttpClient,
		Config:       f.Config,
		IO:           f.IOStreams,
		Patterns:     []string{"ports", "ipv4"},
		OutputFormat: "json",
		Recursive:    true,
	}

	cmd := &cobra.Command{
		Use:   "patterns <repository>",
		Short: "Analyze code patterns in repository",
		Long: heredoc.Doc(`
			Search for patterns like ports, IP addresses, and configuration values
			in repository code with line numbers and file locations.

			The command will automatically authenticate using the GH_TOKEN environment
			variable if available, or fall back to standard GitHub CLI authentication.

			Repository can be specified as:
			- owner/repo (exact match)
			- repo (searches in default org if configured)
			- partial name (presents selection menu)
		`),
		Example: heredoc.Doc(`
			# Analyze ports and IP addresses in a repository
			$ gh analysis patterns cli/cli --patterns ports,ipv4

			# Search for potential secrets with table output
			$ gh analysis patterns myrepo --patterns secrets --format table

			# Analyze specific file types only
			$ gh analysis patterns webapp --patterns ports --file-types go,yaml,json

			# Show summary with detailed results
			$ gh analysis patterns api-service --patterns ports,ipv4 --summary
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Repository = args[0]

			if runF != nil {
				return runF(opts)
			}
			return runPatterns(opts)
		},
	}

	cmd.Flags().StringSliceVar(&opts.Patterns, "patterns", opts.Patterns, "Patterns to search for (ports,ipv4,secrets,urls)")
	cmd.Flags().StringVar(&opts.OutputFormat, "format", opts.OutputFormat, "Output format (json,table,csv)")
	cmd.Flags().BoolVar(&opts.Recursive, "recursive", opts.Recursive, "Search recursively through directories")
	cmd.Flags().StringSliceVar(&opts.FileTypes, "file-types", []string{}, "Limit to specific file types (e.g., go,yaml,json)")
	cmd.Flags().BoolVar(&opts.Summary, "summary", false, "Show analysis summary")

	return cmd
}

// runPatterns executes the patterns analysis
func runPatterns(opts *PatternsOptions) error {
	// 1. Ensure authentication
	config, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := ensureAuthentication(config); err != nil {
		return err
	}

	// 2. Load analysis configuration
	analysisConfig, err := analysisconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load analysis config: %w", err)
	}

	// 3. Set up API client
	httpClient, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	client := api.NewClientFromHTTP(httpClient)

	// 4. Select repository
	selector := NewRepoSelector(client, nil, analysisConfig, opts.IO)
	repo, err := selector.SelectRepository(opts.Repository)
	if err != nil {
		return fmt.Errorf("failed to select repository: %w", err)
	}

	fmt.Fprintf(opts.IO.ErrOut, "Analyzing repository: %s/%s\n", repo.RepoOwner(), repo.RepoName())

	// 5. Validate patterns
	for _, pattern := range opts.Patterns {
		if _, exists := BuiltinPatterns[pattern]; !exists {
			return fmt.Errorf("unknown pattern: %s (available: ports, ipv4, secrets, urls)", pattern)
		}
	}

	// 6. Analyze repository
	analyzer := NewRepositoryAnalyzer(client, repo)
	results, err := analyzer.AnalyzePatterns(opts.Patterns, opts.FileTypes)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	fmt.Fprintf(opts.IO.ErrOut, "Found %d pattern matches\n\n", len(results))

	// 7. Output results
	formatter := NewOutputFormatter(opts.IO)
	if err := formatter.Format(results, opts.OutputFormat); err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// 8. Show summary if requested
	if opts.Summary {
		if err := formatter.OutputSummary(results); err != nil {
			return fmt.Errorf("failed to show summary: %w", err)
		}
	}

	return nil
}

// ensureAuthentication ensures the user is authenticated for GitHub API access
func ensureAuthentication(config gh.Config) error {
	// Check for GH_TOKEN environment variable first
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return nil // Already authenticated via env
	}

	// Check for GITHUB_TOKEN as well (common alternative)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return nil // Already authenticated via env
	}

	// Fall back to existing auth check
	if !cmdutil.CheckAuth(config) {
		return fmt.Errorf("authentication required: set GH_TOKEN environment variable or run 'gh auth login'")
	}

	return nil
}

// NewRepoSelector creates a new repository selector (simplified for patterns package)
func NewRepoSelector(client *api.Client, prompter prompter.Prompter, config *analysisconfig.AnalysisConfig, io *iostreams.IOStreams) *RepoSelector {
	return &RepoSelector{
		client:   client,
		prompter: prompter,
		config:   config,
		io:       io,
	}
}

// RepoSelector handles repository selection and matching logic (simplified)
type RepoSelector struct {
	client   *api.Client
	prompter prompter.Prompter
	config   *analysisconfig.AnalysisConfig
	io       *iostreams.IOStreams
}

// SelectRepository selects a repository based on the query (simplified implementation)
func (rs *RepoSelector) SelectRepository(query string) (ghrepo.Interface, error) {
	// For now, just try to parse as owner/repo
	if parts := split(query, "/"); len(parts) == 2 {
		return ghrepo.New(parts[0], parts[1]), nil
	}
	
	// If no slash, assume it's just a repo name and needs an owner
	if rs.config.DefaultOrg != "" {
		return ghrepo.New(rs.config.DefaultOrg, query), nil
	}
	
	return nil, fmt.Errorf("repository must be specified as 'owner/repo' or set a default org with 'gh analysis config set-default-org'")
}

// split is a simple string splitting function
func split(s, sep string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// OutputFormatter handles different output formats for analysis results
type OutputFormatter struct {
	io *iostreams.IOStreams
}

// NewOutputFormatter creates a new output formatter
func NewOutputFormatter(io *iostreams.IOStreams) *OutputFormatter {
	return &OutputFormatter{io: io}
}

// Format formats and outputs the analysis results
func (of *OutputFormatter) Format(results []AnalysisResult, format string) error {
	switch format {
	case "json":
		return of.outputJSON(results)
	case "table":
		return of.outputTable(results)
	case "csv":
		return of.outputCSV(results)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: json, table, csv)", format)
	}
}

// outputJSON outputs results in JSON format
func (of *OutputFormatter) outputJSON(results []AnalysisResult) error {
	encoder := json.NewEncoder(of.io.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// outputTable outputs results in a human-readable table format
func (of *OutputFormatter) outputTable(results []AnalysisResult) error {
	if len(results) == 0 {
		fmt.Fprintln(of.io.Out, "No patterns found.")
		return nil
	}

	// Simple table output for now
	fmt.Fprintf(of.io.Out, "%-30s %-6s %-10s %-15s %s\n", "FILE", "LINE", "PATTERN", "KEY", "VALUE")
	fmt.Fprintf(of.io.Out, "%s\n", strings.Repeat("-", 80))
	
	for _, result := range results {
		fmt.Fprintf(of.io.Out, "%-30s %-6d %-10s %-15s %s\n",
			truncate(result.File, 30),
			result.LineNumber,
			result.Pattern,
			truncate(result.Key, 15),
			truncate(result.Value, 20))
	}

	return nil
}

// outputCSV outputs results in CSV format
func (of *OutputFormatter) outputCSV(results []AnalysisResult) error {
	writer := csv.NewWriter(of.io.Out)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"file", "line_number", "line", "pattern", "key", "value", "repository"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, result := range results {
		record := []string{
			result.File,
			fmt.Sprintf("%d", result.LineNumber),
			result.Line,
			result.Pattern,
			result.Key,
			result.Value,
			result.Repository,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// OutputSummary outputs a summary of the analysis results
func (of *OutputFormatter) OutputSummary(results []AnalysisResult) error {
	totalResults := len(results)
	
	if totalResults == 0 {
		fmt.Fprintln(of.io.Out, "No patterns found in the repository.")
		return nil
	}

	// Count results by pattern
	patternCounts := make(map[string]int)
	fileCounts := make(map[string]int)

	for _, result := range results {
		patternCounts[result.Pattern]++
		fileCounts[result.File]++
	}

	fmt.Fprintf(of.io.Out, "\nAnalysis Summary:\n")
	fmt.Fprintf(of.io.Out, "Total findings: %d\n", totalResults)
	fmt.Fprintf(of.io.Out, "Files analyzed: %d\n", len(fileCounts))
	
	fmt.Fprintf(of.io.Out, "\nFindings by pattern:\n")
	for pattern, count := range patternCounts {
		fmt.Fprintf(of.io.Out, "  %s: %d\n", pattern, count)
	}

	return nil
}

// truncate truncates a string to maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}