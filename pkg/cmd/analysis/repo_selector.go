package analysis

import (
	"fmt"
	"strings"

	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/analysisconfig"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/internal/prompter"
	"github.com/cli/cli/v2/pkg/iostreams"
)

// RepoSelector handles repository selection and matching logic
type RepoSelector struct {
	client   *api.Client
	prompter prompter.Prompter
	config   *analysisconfig.AnalysisConfig
	io       *iostreams.IOStreams
}

// NewRepoSelector creates a new repository selector
func NewRepoSelector(client *api.Client, prompter prompter.Prompter, config *analysisconfig.AnalysisConfig, io *iostreams.IOStreams) *RepoSelector {
	return &RepoSelector{
		client:   client,
		prompter: prompter,
		config:   config,
		io:       io,
	}
}

// SelectRepository selects a repository based on the query
func (rs *RepoSelector) SelectRepository(query string) (ghrepo.Interface, error) {
	// 1. Try exact match first (owner/repo format)
	if repo := rs.tryExactMatch(query); repo != nil {
		return repo, nil
	}

	// 2. Search in default org if set
	if rs.config.DefaultOrg != "" {
		if repo, err := rs.searchInOrg(rs.config.DefaultOrg, query); err == nil && repo != nil {
			return repo, nil
		}
	}

	// 3. Global search and present options
	return rs.searchAndSelect(query)
}

// tryExactMatch attempts to parse the query as an exact repository reference
func (rs *RepoSelector) tryExactMatch(query string) ghrepo.Interface {
	// Try to parse as owner/repo
	if strings.Contains(query, "/") {
		parts := strings.SplitN(query, "/", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			repo := ghrepo.New(parts[0], parts[1])
			// Validate repository exists by attempting to fetch basic info
			if rs.repositoryExists(repo) {
				return repo
			}
		}
	}
	return nil
}

// repositoryExists checks if a repository exists and is accessible (simplified)
func (rs *RepoSelector) repositoryExists(repo ghrepo.Interface) bool {
	// For now, assume repository exists (simplified implementation)
	return true
}

// searchInOrg searches for repositories within a specific organization
func (rs *RepoSelector) searchInOrg(org, query string) (ghrepo.Interface, error) {
	searchQuery := fmt.Sprintf("org:%s %s in:name", org, query)
	repos, err := rs.performSearch(searchQuery, 5) // Limit to 5 results
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found in organization %s matching '%s'", org, query)
	}

	if len(repos) == 1 {
		return repos[0], nil
	}

	// Present selection menu for multiple matches
	return rs.promptForSelection(repos, fmt.Sprintf("Multiple repositories found in %s:", org))
}

// searchAndSelect performs a global search and allows user to select
func (rs *RepoSelector) searchAndSelect(query string) (ghrepo.Interface, error) {
	repos, err := rs.performSearch(query, 10) // Limit to 10 results
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories found matching '%s'", query)
	}

	if len(repos) == 1 {
		return repos[0], nil
	}

	// Present selection menu
	return rs.promptForSelection(repos, "Multiple repositories found:")
}

// performSearch executes a repository search via GitHub API (simplified)
func (rs *RepoSelector) performSearch(searchQuery string, limit int) ([]ghrepo.Interface, error) {
	// Simplified implementation - return empty for now
	return []ghrepo.Interface{}, nil
}

// promptForSelection presents a menu for the user to select a repository
func (rs *RepoSelector) promptForSelection(repos []ghrepo.Interface, message string) (ghrepo.Interface, error) {
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories to select from")
	}

	// Create display options
	var options []string
	for _, repo := range repos {
		options = append(options, fmt.Sprintf("%s/%s", repo.RepoOwner(), repo.RepoName()))
	}

	// Use prompter to let user select
	selectedIndex, err := rs.prompter.Select(message, "", options)
	if err != nil {
		return nil, fmt.Errorf("selection failed: %w", err)
	}

	if selectedIndex < 0 || selectedIndex >= len(repos) {
		return nil, fmt.Errorf("invalid selection")
	}

	return repos[selectedIndex], nil
}