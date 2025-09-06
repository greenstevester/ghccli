package analysis

import (
	"encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/cli/cli/v2/internal/tableprinter"
	"github.com/cli/cli/v2/pkg/cmd/analysis/patterns"
	"github.com/cli/cli/v2/pkg/iostreams"
)

// OutputFormatter handles different output formats for analysis results
type OutputFormatter struct {
	io *iostreams.IOStreams
}

// NewOutputFormatter creates a new output formatter
func NewOutputFormatter(io *iostreams.IOStreams) *OutputFormatter {
	return &OutputFormatter{io: io}
}

// Format formats and outputs the analysis results
func (of *OutputFormatter) Format(results []patterns.AnalysisResult, format string) error {
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
func (of *OutputFormatter) outputJSON(results []patterns.AnalysisResult) error {
	encoder := json.NewEncoder(of.io.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// outputTable outputs results in a human-readable table format
func (of *OutputFormatter) outputTable(results []patterns.AnalysisResult) error {
	if len(results) == 0 {
		fmt.Fprintln(of.io.Out, "No patterns found.")
		return nil
	}

	tp := tableprinter.New(of.io, tableprinter.WithHeader("FILE", "LINE", "PATTERN", "KEY", "VALUE"))

	for _, result := range results {
		tp.AddField(truncateString(result.File, 40))
		tp.AddField(fmt.Sprintf("%d", result.LineNumber))
		tp.AddField(result.Pattern)
		tp.AddField(truncateString(result.Key, 20))
		tp.AddField(truncateString(result.Value, 30))
		tp.EndRow()
	}

	return tp.Render()
}

// outputCSV outputs results in CSV format
func (of *OutputFormatter) outputCSV(results []patterns.AnalysisResult) error {
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

// outputSummary outputs a summary of the analysis results
func (of *OutputFormatter) OutputSummary(results []patterns.AnalysisResult) error {
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

	if len(fileCounts) <= 10 {
		fmt.Fprintf(of.io.Out, "\nFindings by file:\n")
		for file, count := range fileCounts {
			fmt.Fprintf(of.io.Out, "  %s: %d\n", file, count)
		}
	}

	return nil
}

// truncateString truncates a string to a maximum length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// FormatPatternList formats the list of available patterns
func FormatPatternList(io *iostreams.IOStreams) error {
	fmt.Fprintln(io.Out, "Available patterns:")
	fmt.Fprintln(io.Out)

	for name, pattern := range patterns.BuiltinPatterns {
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