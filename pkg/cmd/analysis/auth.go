package analysis

import (
	"fmt"
	"os"

	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/pkg/cmdutil"
)

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