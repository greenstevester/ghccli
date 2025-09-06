# Analysis Extension Design for GitHub CLI

## Overview

This document outlines how to extend the GitHub CLI project to support analysis-type queries for searching patterns like ports and IP addresses across repositories with key-value output format.

## Architecture Integration

### Command Structure
Following the existing pattern in `pkg/cmd/search/`, we'll create:

```
pkg/cmd/analysis/
├── analysis.go           # Root analysis command
├── patterns/
│   ├── patterns.go      # Pattern-based analysis (ports, IPs, etc.)
│   └── patterns_test.go
├── security/
│   ├── security.go      # Security-focused analysis
│   └── security_test.go
└── custom/
    ├── custom.go        # Custom regex patterns
    └── custom_test.go
```

### Registration in Root Command

In `pkg/cmd/root/root.go`, add:

```go
analysisCmd "github.com/cli/cli/v2/pkg/cmd/analysis"

// In NewCmdRoot function:
cmd.AddCommand(analysisCmd.NewCmdAnalysis(f))
```

## Configuration Management

### Default Organization Persistence

Create `~/.ghccli/config.yaml` for storing user preferences:

```go
// internal/analysisconfig/config.go
package analysisconfig

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

type AnalysisConfig struct {
    DefaultOrg     string            `yaml:"default_org"`
    SearchPatterns map[string]string `yaml:"search_patterns"`
    OutputFormat   string            `yaml:"output_format"`
}

const configFileName = ".ghccli/config.yaml"

func (c *AnalysisConfig) ConfigPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, configFileName)
}

func LoadConfig() (*AnalysisConfig, error) {
    // Implementation for loading/creating config
}

func (c *AnalysisConfig) Save() error {
    // Implementation for saving config
}
```

## Repository Selection Logic

### Auto-login with Environment Token

```go
// pkg/cmd/analysis/auth.go
package analysis

import (
    "os"
    "github.com/cli/cli/v2/internal/gh"
)

func ensureAuthentication(config gh.Config) error {
    // Check for GH_TOKEN environment variable first
    if token := os.Getenv("GH_TOKEN"); token != "" {
        return nil // Already authenticated via env
    }
    
    // Fall back to existing auth check
    if !cmdutil.CheckAuth(config) {
        return fmt.Errorf("authentication required: set GH_TOKEN or run 'gh auth login'")
    }
    
    return nil
}
```

### Repository Matching and Selection

```go
// pkg/cmd/analysis/repo_selector.go
package analysis

import (
    "fmt"
    "github.com/cli/cli/v2/api"
    "github.com/cli/cli/v2/internal/ghrepo"
    "github.com/cli/cli/v2/internal/prompter"
)

type RepoSelector struct {
    client   *api.Client
    prompter prompter.Prompter
    config   *analysisconfig.AnalysisConfig
}

func (rs *RepoSelector) SelectRepository(query string) (ghrepo.Interface, error) {
    // 1. Try exact match first
    if repo := rs.tryExactMatch(query); repo != nil {
        return repo, nil
    }
    
    // 2. Search in default org if set
    if rs.config.DefaultOrg != "" {
        if repo := rs.searchInOrg(rs.config.DefaultOrg, query); repo != nil {
            return repo, nil
        }
    }
    
    // 3. Global search and present options
    return rs.searchAndSelect(query)
}

func (rs *RepoSelector) tryExactMatch(query string) ghrepo.Interface {
    // Try to parse as owner/repo
    if repo, err := ghrepo.FromFullName(query); err == nil {
        // Validate repository exists
        if rs.repositoryExists(repo) {
            return repo
        }
    }
    return nil
}

func (rs *RepoSelector) searchInOrg(org, query string) ghrepo.Interface {
    // Search for repositories in the default org
    searchQuery := fmt.Sprintf("org:%s %s in:name", org, query)
    return rs.performSearch(searchQuery)
}

func (rs *RepoSelector) searchAndSelect(query string) (ghrepo.Interface, error) {
    // Perform global search and let user select
    repos := rs.globalSearch(query)
    if len(repos) == 0 {
        return nil, fmt.Errorf("no repositories found matching '%s'", query)
    }
    
    if len(repos) == 1 {
        return repos[0], nil
    }
    
    // Present selection menu
    return rs.prompter.SelectRepository(repos)
}
```

## Pattern Analysis Engine

### Core Pattern Definitions

```go
// pkg/cmd/analysis/patterns/patterns.go
package patterns

import (
    "regexp"
    "fmt"
)

type PatternDefinition struct {
    Name        string
    Description string
    Pattern     *regexp.Regexp
    Examples    []string
}

var BuiltinPatterns = map[string]PatternDefinition{
    "ports": {
        Name:        "ports",
        Description: "Network ports (1-65535)",
        Pattern:     regexp.MustCompile(`\b(?:port|PORT)\s*[:=]\s*(['"]?)(\d{1,5})\1\b`),
        Examples:    []string{`port: 8080`, `PORT="3000"`},
    },
    "ipv4": {
        Name:        "ipv4", 
        Description: "IPv4 addresses",
        Pattern:     regexp.MustCompile(`\b(?:ip|host|server|address|HOST|IP)\s*[:=]\s*(['"]?)(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\1\b`),
        Examples:    []string{`host: "192.168.1.1"`, `IP=10.0.0.1`},
    },
    "secrets": {
        Name:        "secrets",
        Description: "Potential secrets (keys, tokens, passwords)",
        Pattern:     regexp.MustCompile(`\b(?:api[_-]?key|token|password|secret|pwd|auth)\s*[:=]\s*['"]([^'"]{8,})['"]`),
        Examples:    []string{`api_key: "abc123xyz789"`, `token="ghp_xxxxxxxxxxxx"`},
    },
}

type AnalysisResult struct {
    File        string              `json:"file"`
    LineNumber  int                 `json:"line_number"`
    Line        string              `json:"line"`
    Pattern     string              `json:"pattern"`
    Key         string              `json:"key"`
    Value       string              `json:"value"`
    Repository  string              `json:"repository"`
}

func AnalyzeContent(content string, filename string, patterns []string) []AnalysisResult {
    var results []AnalysisResult
    
    lines := strings.Split(content, "\n")
    for lineNum, line := range lines {
        for _, patternName := range patterns {
            if pattern, exists := BuiltinPatterns[patternName]; exists {
                matches := pattern.Pattern.FindAllStringSubmatch(line, -1)
                for _, match := range matches {
                    if len(match) >= 3 {
                        results = append(results, AnalysisResult{
                            File:       filename,
                            LineNumber: lineNum + 1,
                            Line:       strings.TrimSpace(line),
                            Pattern:    patternName,
                            Key:        extractKey(line, match[0]),
                            Value:      match[2], // The captured value
                        })
                    }
                }
            }
        }
    }
    
    return results
}
```

## Command Implementation

### Main Analysis Command

```go
// pkg/cmd/analysis/analysis.go
package analysis

import (
    "github.com/MakeNowJust/heredoc"
    "github.com/cli/cli/v2/pkg/cmdutil"
    "github.com/spf13/cobra"
    
    patternsCmd "github.com/cli/cli/v2/pkg/cmd/analysis/patterns"
)

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
        `),
        Example: heredoc.Doc(`
            # Analyze ports and IP addresses in a repository
            $ gh analysis patterns cli/cli --patterns ports,ipv4
            
            # Search for potential secrets in your default org
            $ gh analysis patterns myrepo --patterns secrets
            
            # Set default organization for future searches
            $ gh analysis config set-default-org myorg
        `),
    }
    
    cmd.AddCommand(patternsCmd.NewCmdPatterns(f, nil))
    cmd.AddCommand(NewCmdConfig(f))
    
    return cmd
}
```

### Patterns Command Implementation

```go
// pkg/cmd/analysis/patterns/patterns.go
package patterns

type PatternsOptions struct {
    HttpClient   func() (*http.Client, error)
    Config       func() (gh.Config, error)
    IO           *iostreams.IOStreams
    Exporter     cmdutil.Exporter
    
    Repository   string
    Patterns     []string
    OutputFormat string
    Recursive    bool
    FileTypes    []string
}

func NewCmdPatterns(f *cmdutil.Factory, runF func(*PatternsOptions) error) *cobra.Command {
    opts := &PatternsOptions{
        HttpClient: f.HttpClient,
        Config:     f.Config,
        IO:         f.IOStreams,
        Patterns:   []string{"ports", "ipv4"},
        OutputFormat: "json",
    }
    
    cmd := &cobra.Command{
        Use:   "patterns <repository>",
        Short: "Analyze code patterns in repository",
        Long: heredoc.Doc(`
            Search for patterns like ports, IP addresses, and configuration values
            in repository code with line numbers and file locations.
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
    
    cmd.Flags().StringSliceVar(&opts.Patterns, "patterns", opts.Patterns, "Patterns to search for (ports,ipv4,secrets)")
    cmd.Flags().StringVar(&opts.OutputFormat, "format", opts.OutputFormat, "Output format (json,table,csv)")
    cmd.Flags().BoolVar(&opts.Recursive, "recursive", true, "Search recursively through directories")
    cmd.Flags().StringSliceVar(&opts.FileTypes, "file-types", []string{}, "Limit to specific file types")
    
    return cmd
}

func runPatterns(opts *PatternsOptions) error {
    // 1. Ensure authentication
    config, err := opts.Config()
    if err != nil {
        return err
    }
    
    if err := ensureAuthentication(config); err != nil {
        return err
    }
    
    // 2. Select repository
    httpClient, err := opts.HttpClient()
    if err != nil {
        return err
    }
    
    client := api.NewClientFromHTTP(httpClient)
    selector := &RepoSelector{
        client:   client,
        prompter: f.Prompter,
        config:   loadAnalysisConfig(),
    }
    
    repo, err := selector.SelectRepository(opts.Repository)
    if err != nil {
        return err
    }
    
    // 3. Analyze repository
    analyzer := &RepositoryAnalyzer{
        client: client,
        repo:   repo,
    }
    
    results, err := analyzer.AnalyzePatterns(opts.Patterns, opts.FileTypes)
    if err != nil {
        return err
    }
    
    // 4. Output results
    return outputResults(opts.IO, results, opts.OutputFormat)
}
```

## Output Formatting

### Multiple Output Formats

```go
// pkg/cmd/analysis/output.go
package analysis

import (
    "encoding/json"
    "encoding/csv"
    "fmt"
    "github.com/cli/cli/v2/pkg/iostreams"
    "github.com/cli/cli/v2/internal/tableprinter"
)

func outputResults(io *iostreams.IOStreams, results []AnalysisResult, format string) error {
    switch format {
    case "json":
        return outputJSON(io, results)
    case "table":
        return outputTable(io, results) 
    case "csv":
        return outputCSV(io, results)
    default:
        return fmt.Errorf("unsupported output format: %s", format)
    }
}

func outputJSON(io *iostreams.IOStreams, results []AnalysisResult) error {
    encoder := json.NewEncoder(io.Out)
    encoder.SetIndent("", "  ")
    return encoder.Encode(results)
}

func outputTable(io *iostreams.IOStreams, results []AnalysisResult) error {
    tp := tableprinter.New(io, tableprinter.WithHeader("FILE", "LINE", "PATTERN", "KEY", "VALUE"))
    
    for _, result := range results {
        tp.AddField(result.File)
        tp.AddField(fmt.Sprintf("%d", result.LineNumber))
        tp.AddField(result.Pattern)
        tp.AddField(result.Key)
        tp.AddField(result.Value)
        tp.EndRow()
    }
    
    return tp.Render()
}

func outputCSV(io *iostreams.IOStreams, results []AnalysisResult) error {
    writer := csv.NewWriter(io.Out)
    defer writer.Flush()
    
    // Write header
    if err := writer.Write([]string{"file", "line_number", "pattern", "key", "value", "repository"}); err != nil {
        return err
    }
    
    // Write data
    for _, result := range results {
        record := []string{
            result.File,
            fmt.Sprintf("%d", result.LineNumber),
            result.Pattern,
            result.Key,
            result.Value,
            result.Repository,
        }
        if err := writer.Write(record); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Usage Examples

### Command Line Usage

```bash
# Set default organization (persisted to ~/.ghccli/config.yaml)
$ gh analysis config set-default-org myorg

# Analyze patterns in a specific repository
$ gh analysis patterns cli/cli --patterns ports,ipv4 --format json

# Search for secrets with table output
$ gh analysis patterns myrepo --patterns secrets --format table

# Analyze specific file types only
$ gh analysis patterns webapp --patterns ports --file-types go,yaml,json

# Export to CSV
$ gh analysis patterns api-service --patterns ports,ipv4 --format csv > analysis.csv
```

### Sample Output

**JSON Format:**
```json
[
  {
    "file": "config/server.yaml",
    "line_number": 15,
    "line": "  port: 8080",
    "pattern": "ports",
    "key": "port",
    "value": "8080",
    "repository": "myorg/webapp"
  },
  {
    "file": "docker-compose.yml", 
    "line_number": 23,
    "line": "    - \"HOST=192.168.1.100\"",
    "pattern": "ipv4",
    "key": "HOST", 
    "value": "192.168.1.100",
    "repository": "myorg/webapp"
  }
]
```

**Table Format:**
```
FILE                 LINE  PATTERN  KEY   VALUE
config/server.yaml   15    ports    port  8080
docker-compose.yml   23    ipv4     HOST  192.168.1.100
```

## Integration Points

### Authentication Integration
- Leverages existing `pkg/cmdutil/auth_check.go`
- Supports `GH_TOKEN` environment variable
- Falls back to standard `gh auth login` flow

### API Integration  
- Uses existing `api/client.go` for GitHub API calls
- Leverages repository search and content API
- Follows existing rate limiting and error handling patterns

### Configuration Integration
- Extends existing config system in `internal/config/`
- Stores analysis-specific settings separately
- Follows existing configuration file patterns

This design provides a comprehensive extension to the GitHub CLI that follows existing architectural patterns while adding powerful code analysis capabilities with user-friendly repository selection and flexible output formatting.