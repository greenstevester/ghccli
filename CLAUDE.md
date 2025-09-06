# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building
- **Main build**: `make` or `go run script/build.go` - builds `bin/gh` executable
- **Cross-platform build**: `go run script/build.go` (works on Windows)
- **Manual pages**: `make manpages`
- **Shell completions**: `make completions`
- **Clean**: `make clean`

### Testing
- **All tests**: `go test ./...`
- **Acceptance tests**: `go test -tags acceptance ./acceptance` 
- **Single test**: `go test ./pkg/cmd/[command]/[subcommand]/`

### Prerequisites
- Go 1.24+
- The project uses Go modules with dependencies managed in `go.mod`

## Codebase Architecture

### Project Structure
- **`cmd/`** - Main packages for building binaries (`gh` executable)
- **`pkg/cmd/`** - Implementation for individual gh commands, organized as `pkg/cmd/<command>/<subcommand>/`
- **`pkg/cmdutil/`** - Command utilities and the Factory pattern for dependency injection
- **`api/`** - GitHub API utilities and request handling
- **`git/`** - Local git repository utilities 
- **`internal/`** - Go packages specific to this project
- **`script/`** - Build and release scripts

### Command Architecture
Commands follow a consistent pattern using Cobra:
1. Each command lives in `pkg/cmd/<command>/<subcommand>/<subcommand>.go`
2. Commands expose a `NewCmd[Name](*cmdutil.Factory) *cobra.Command` function
3. The Factory provides access to dependencies (IO, HTTP client, config, git client)
4. Command logic is in a `RunE` function that calls implementation functions
5. Commands are registered in `pkg/cmd/root/root.go`

### Key Patterns
- **Factory Pattern**: `cmdutil.Factory` provides dependency injection for commands
- **Cobra Commands**: All CLI commands use the spf13/cobra library
- **IO Abstraction**: Commands use `iostreams.IOStreams` for testable I/O
- **Authentication**: Commands check auth via `cmdutil.CheckAuth()` before execution
- **GitHub API**: API calls use the `api/` package with proper authentication

## Testing Guidelines
- Commands are tested using mocks and stubs to avoid real API calls or git operations
- Test files use table-driven tests for different input variations
- Tests should never make real network requests or modify the filesystem outside test scopes
- Use existing patterns from other command tests for consistency

## Development Workflow
1. Commands require approval in issues before implementation
2. Create branch: `git checkout -b my-branch-name`
3. Add tests alongside implementation
4. Submit PR: `gh pr create --web`
5. Manual pages are auto-generated on release - don't submit manual changes

## Code Conventions
- Follow existing Go conventions in the codebase
- Command help text is embedded in command source code using heredoc
- Keep command-specific logic within the command's package
- Use the Factory pattern for accessing dependencies
- Prefer small, composable functions for testability