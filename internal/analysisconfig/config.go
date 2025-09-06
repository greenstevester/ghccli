package analysisconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AnalysisConfig holds analysis-specific configuration
type AnalysisConfig struct {
	DefaultOrg     string            `yaml:"default_org"`
	SearchPatterns map[string]string `yaml:"search_patterns"`
	OutputFormat   string            `yaml:"output_format"`
}

const configFileName = ".ghccli/analysis.yaml"

// NewAnalysisConfig creates a new config with defaults
func NewAnalysisConfig() *AnalysisConfig {
	return &AnalysisConfig{
		DefaultOrg:     "",
		SearchPatterns: make(map[string]string),
		OutputFormat:   "json",
	}
}

// ConfigPath returns the full path to the analysis config file
func (c *AnalysisConfig) ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, configFileName)
}

// LoadConfig loads the analysis configuration from file
func LoadConfig() (*AnalysisConfig, error) {
	config := NewAnalysisConfig()
	configPath := config.ConfigPath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return config, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save saves the configuration to file
func (c *AnalysisConfig) Save() error {
	configPath := c.ConfigPath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetDefaultOrg sets the default organization
func (c *AnalysisConfig) SetDefaultOrg(org string) error {
	c.DefaultOrg = org
	return c.Save()
}

// GetDefaultOrg returns the default organization
func (c *AnalysisConfig) GetDefaultOrg() string {
	return c.DefaultOrg
}