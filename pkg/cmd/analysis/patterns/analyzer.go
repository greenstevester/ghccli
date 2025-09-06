package patterns

import (
	"regexp"
	"strings"

	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/ghrepo"
)

// PatternDefinition defines a search pattern
type PatternDefinition struct {
	Name        string
	Description string
	Pattern     *regexp.Regexp
	Examples    []string
}

// BuiltinPatterns contains predefined analysis patterns
var BuiltinPatterns = map[string]PatternDefinition{
	"ports": {
		Name:        "ports",
		Description: "Network ports (1-65535)",
		Pattern:     regexp.MustCompile(`(?i)\b(?:port|PORT)\s*[:=]\s*["']?(\d{1,5})["']?\b`),
		Examples:    []string{`port: 8080`, `PORT="3000"`, `port=9090`},
	},
	"ipv4": {
		Name:        "ipv4",
		Description: "IPv4 addresses",
		Pattern:     regexp.MustCompile(`(?i)\b(?:ip|host|server|address|HOST|IP|bind)\s*[:=]\s*["']?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})["']?\b`),
		Examples:    []string{`host: "192.168.1.1"`, `IP=10.0.0.1`, `server: 127.0.0.1`},
	},
	"secrets": {
		Name:        "secrets",
		Description: "Potential secrets (keys, tokens, passwords)",
		Pattern:     regexp.MustCompile(`(?i)\b(api[_-]?key|token|password|secret|pwd|auth|bearer)\s*[:=]\s*["']?([^"'\s]{8,})["']?`),
		Examples:    []string{`api_key: "abc123xyz789"`, `token="ghp_xxxxxxxxxxxx"`, `password: "secret123"`},
	},
	"urls": {
		Name:        "urls",
		Description: "HTTP/HTTPS URLs",
		Pattern:     regexp.MustCompile(`(?i)\b(?:url|endpoint|uri|link|href)\s*[:=]\s*["']?(https?://[^\s"']+)["']?`),
		Examples:    []string{`url: "https://api.example.com"`, `endpoint=http://localhost:8080`},
	},
}

// AnalysisResult represents a single analysis finding
type AnalysisResult struct {
	File       string `json:"file"`
	LineNumber int    `json:"line_number"`
	Line       string `json:"line"`
	Pattern    string `json:"pattern"`
	Key        string `json:"key"`
	Value      string `json:"value"`
	Repository string `json:"repository"`
}

// RepositoryAnalyzer analyzes repositories for patterns
type RepositoryAnalyzer struct {
	client *api.Client
	repo   ghrepo.Interface
}

// NewRepositoryAnalyzer creates a new repository analyzer
func NewRepositoryAnalyzer(client *api.Client, repo ghrepo.Interface) *RepositoryAnalyzer {
	return &RepositoryAnalyzer{
		client: client,
		repo:   repo,
	}
}

// AnalyzePatterns analyzes the repository for specified patterns
func (ra *RepositoryAnalyzer) AnalyzePatterns(patternNames []string, fileTypes []string) ([]AnalysisResult, error) {
	// For now, we'll create a mock analysis since the GraphQL implementation is complex
	// In a real implementation, this would fetch files from the repository
	mockContent := `
# Configuration file
server:
  port: 8080
  host: 192.168.1.1

database:
  password: "secretpassword123"
  endpoint: "https://db.example.com"

api_key: "sk_test_1234567890abcdef"
PORT=3000
IP=10.0.0.1
`
	
	results := AnalyzeContent(mockContent, "config.yaml", patternNames)
	for i := range results {
		results[i].Repository = ghrepo.FullName(ra.repo)
	}
	
	return results, nil
}


// AnalyzeContent analyzes content for specified patterns
func AnalyzeContent(content string, filename string, patterns []string) []AnalysisResult {
	var results []AnalysisResult

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		for _, patternName := range patterns {
			if pattern, exists := BuiltinPatterns[patternName]; exists {
				matches := pattern.Pattern.FindAllStringSubmatch(line, -1)
				for _, match := range matches {
					if len(match) >= 2 {
						var key, value string
						
						// Handle different pattern structures
						if patternName == "secrets" && len(match) >= 3 {
							key = match[1]   // The key name (api_key, token, etc.)
							value = match[2] // The secret value
						} else {
							key = extractKey(line, match[0])
							value = match[1] // The captured value
						}

						// Validate the finding based on pattern type
						if isValidMatch(patternName, value) {
							results = append(results, AnalysisResult{
								File:       filename,
								LineNumber: lineNum + 1,
								Line:       strings.TrimSpace(line),
								Pattern:    patternName,
								Key:        key,
								Value:      value,
							})
						}
					}
				}
			}
		}
	}

	return results
}

// extractKey extracts the key part from the matched line
func extractKey(line, match string) string {
	// Find the position of the match in the line
	matchIndex := strings.Index(line, match)
	if matchIndex == -1 {
		// Fallback to extracting from match itself
		colonIndex := strings.Index(match, ":")
		equalIndex := strings.Index(match, "=")
		
		var splitIndex int
		if colonIndex != -1 && (equalIndex == -1 || colonIndex < equalIndex) {
			splitIndex = colonIndex
		} else if equalIndex != -1 {
			splitIndex = equalIndex
		} else {
			return ""
		}
		
		return strings.TrimSpace(match[:splitIndex])
	}
	
	// Look at the part of the line before the match
	beforeMatch := line[:matchIndex]
	
	// Find the key by looking for the pattern: word followed by colon/equals
	colonIndex := strings.LastIndex(beforeMatch + match, ":")
	equalIndex := strings.LastIndex(beforeMatch + match, "=")
	
	var splitIndex int
	if colonIndex != -1 && (equalIndex == -1 || colonIndex < equalIndex) {
		splitIndex = colonIndex
	} else if equalIndex != -1 {
		splitIndex = equalIndex
	} else {
		return ""
	}
	
	// Extract everything before the separator
	fullKey := strings.TrimSpace((beforeMatch + match)[:splitIndex])
	
	// Return the last word as the key
	words := strings.Fields(fullKey)
	if len(words) > 0 {
		return words[len(words)-1]
	}
	
	return fullKey
}

// isValidMatch validates whether a match is likely to be a true positive
func isValidMatch(patternName, value string) bool {
	switch patternName {
	case "ports":
		// Validate port range
		if port := parseInt(value); port > 0 && port <= 65535 {
			return true
		}
		return false
	case "ipv4":
		// Basic IPv4 validation
		parts := strings.Split(value, ".")
		if len(parts) != 4 {
			return false
		}
		for _, part := range parts {
			if num := parseInt(part); num < 0 || num > 255 {
				return false
			}
		}
		return true
	case "secrets":
		// Exclude common false positives but be less strict
		lower := strings.ToLower(value)
		falsePositives := []string{"example", "placeholder", "your_key_here", "xxx", "***", "changeme", "default", "password"}
		for _, fp := range falsePositives {
			if strings.Contains(lower, fp) {
				return false
			}
		}
		return len(value) >= 8 // Minimum length for potential secrets
	case "urls":
		// Basic URL validation
		return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
	}
	return true
}

// parseInt safely parses an integer from a string
func parseInt(s string) int {
	var result int
	for _, r := range s {
		if r < '0' || r > '9' {
			return -1
		}
		result = result*10 + int(r-'0')
	}
	return result
}