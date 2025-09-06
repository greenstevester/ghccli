# pkg/cmd/ - Command Implementation Layer

This directory contains the implementation of all GitHub CLI commands organized by command hierarchy.

## Architecture
- **33 command groups** - Each top-level command (issue, pr, repo, etc.) has its own directory
- **Hierarchical structure** - Commands follow `<command>/<subcommand>/` pattern
- **Cobra integration** - All commands use spf13/cobra framework

## Command Structure Pattern
Each command follows this consistent pattern:
1. **File location**: `pkg/cmd/<command>/<subcommand>/<subcommand>.go`
2. **Function signature**: `NewCmd[Name](*cmdutil.Factory) *cobra.Command`
3. **Options struct**: Contains command-specific options and dependencies
4. **RunE function**: Implements the command logic
5. **Help text**: Embedded using heredoc for documentation

## Key Components
- **Factory dependency**: Commands receive `*cmdutil.Factory` for dependencies
- **IO abstraction**: Use `iostreams.IOStreams` for testable input/output
- **Authentication**: Commands check auth via factory before execution
- **Repository context**: Commands can override repo with `-R` flag

## Examples
- `issue/list/list.go` - Lists issues with filtering options
- `pr/create/create.go` - Creates pull requests with interactive prompts
- `repo/clone/clone.go` - Clones repositories with GitHub-specific enhancements

## Testing
- Commands are unit tested with mocked dependencies
- Tests avoid real API calls or filesystem operations
- Table-driven tests for multiple input variations
- Mock HTTP responses and git operations

## Command Registration
- All commands are registered in `root/root.go`
- Commands are organized into logical groups
- Help text is automatically generated from embedded documentation