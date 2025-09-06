package patterns

import (
	"testing"
)

func TestAnalyzeContent(t *testing.T) {
	testContent := `
# Configuration file
server:
  port: 8080
  host: 192.168.1.1

database:
  password: "mysecret12345"
  endpoint: "https://db.example.com"

api_key: "sk_test_1234567890abcdef"
PORT=3000
IP=10.0.0.1
`

	tests := []struct {
		name           string
		patterns       []string
		expectedCount  int
		expectedValues []string
	}{
		{
			name:           "find ports",
			patterns:       []string{"ports"},
			expectedCount:  2,
			expectedValues: []string{"8080", "3000"},
		},
		{
			name:           "find IPs",
			patterns:       []string{"ipv4"},
			expectedCount:  2,
			expectedValues: []string{"192.168.1.1", "10.0.0.1"},
		},
		{
			name:           "find secrets",
			patterns:       []string{"secrets"},
			expectedCount:  2,
			expectedValues: []string{"mysecret12345", "sk_test_1234567890abcdef"},
		},
		{
			name:           "find URLs",
			patterns:       []string{"urls"},
			expectedCount:  1,
			expectedValues: []string{"https://db.example.com"},
		},
		{
			name:          "find all patterns",
			patterns:      []string{"ports", "ipv4", "secrets", "urls"},
			expectedCount: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := AnalyzeContent(testContent, "test.yaml", tt.patterns)
			
			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
				for _, r := range results {
					t.Logf("Found: %s=%s (pattern: %s, line: %d)", r.Key, r.Value, r.Pattern, r.LineNumber)
				}
			}

			if tt.expectedValues != nil {
				foundValues := make(map[string]bool)
				for _, result := range results {
					foundValues[result.Value] = true
				}
				
				for _, expectedValue := range tt.expectedValues {
					if !foundValues[expectedValue] {
						t.Errorf("expected to find value %s", expectedValue)
					}
				}
			}
		})
	}
}

func TestIsValidMatch(t *testing.T) {
	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		{"ports", "8080", true},
		{"ports", "65536", false}, // out of range
		{"ports", "abc", false},   // not a number
		{"ipv4", "192.168.1.1", true},
		{"ipv4", "256.1.1.1", false}, // invalid IP
		{"ipv4", "192.168.1", false}, // incomplete IP
		{"secrets", "mysecret123", true},
		{"secrets", "password", false}, // common false positive
		{"secrets", "abc", false},      // too short
		{"urls", "https://example.com", true},
		{"urls", "ftp://example.com", false}, // not HTTP(S)
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.value, func(t *testing.T) {
			got := isValidMatch(tt.pattern, tt.value)
			if got != tt.want {
				t.Errorf("isValidMatch(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
			}
		})
	}
}

func TestExtractKey(t *testing.T) {
	tests := []struct {
		line  string
		match string
		want  string
	}{
		{"  port: 8080", "port: 8080", "port"},
		{"PORT=3000", "PORT=3000", "PORT"},
		{"server_port: 9090", "port: 9090", "server_port"},
		{"  api_key: \"secret\"", "api_key: \"secret\"", "api_key"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := extractKey(tt.line, tt.match)
			if got != tt.want {
				t.Errorf("extractKey(%q, %q) = %q, want %q", tt.line, tt.match, got, tt.want)
			}
		})
	}
}