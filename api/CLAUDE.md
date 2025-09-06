# api/ - GitHub API Layer

This directory provides high-level abstractions for interacting with the GitHub REST and GraphQL APIs.

## Architecture
- **Client abstraction** - Wraps HTTP client with GitHub-specific functionality
- **GraphQL queries** - Structured queries for complex data fetching
- **REST API** - RESTful operations for CRUD operations
- **Type definitions** - Go structs representing GitHub API responses

## Key Components

### Client (`client.go`)
- **HTTP client wrapper** - Adds authentication, error handling, and GitHub-specific headers
- **GraphQL support** - Unified interface for GraphQL operations
- **REST support** - Traditional REST API operations
- **Error handling** - GitHub-specific error types and scope suggestions

### Query Files
Organized by GitHub entity type:
- **`queries_issue.go`** - Issue-related queries and types
- **`queries_pr.go`** - Pull request operations
- **`queries_repo.go`** - Repository information
- **`queries_user.go`** - User and organization data
- **`queries_projects_v2.go`** - GitHub Projects v2 integration

### Data Types
Each query file defines corresponding Go structs:
- **Issue** - Issue metadata, comments, labels, assignees
- **PullRequest** - PR details, reviews, checks, merge status
- **Repository** - Repo metadata, permissions, settings
- **User/Organization** - User profiles and organization data

## API Patterns
- **GraphQL preferred** - Use GraphQL for complex, nested data fetching
- **REST for mutations** - Use REST API for create/update/delete operations
- **Pagination** - Built-in support for paginated responses
- **Rate limiting** - Automatic handling of rate limit headers

## Authentication
- **Token-based** - Supports personal access tokens and OAuth
- **Scoped requests** - Suggests required scopes for failed operations
- **Multi-host** - Supports GitHub Enterprise Server instances

## Usage
Commands use this layer to:
1. Fetch data for display (issues, PRs, repos)
2. Perform mutations (create issues, merge PRs)
3. Query user permissions and repository settings
4. Handle pagination and rate limiting automatically