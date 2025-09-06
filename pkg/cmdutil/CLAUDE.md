# pkg/cmdutil/ - Command Utilities Layer

This directory provides shared utilities and the Factory pattern for dependency injection across all commands.

## Architecture
- **Factory Pattern** - Central dependency injection container
- **Shared utilities** - Common functionality used by multiple commands
- **Authentication** - Centralized auth checking and management
- **Configuration** - Common flag handling and argument processing

## Key Components

### Factory (`factory.go`)
The Factory provides dependency injection for commands:
- **IOStreams** - Testable input/output abstraction
- **HttpClient** - Authenticated HTTP client for API calls
- **GitClient** - Git repository operations
- **Config** - Configuration management
- **Browser** - Cross-platform browser launching

### Authentication (`auth_check.go`)
- **CheckAuth()** - Validates GitHub authentication
- **DisableAuthCheck()** - Annotation to skip auth for certain commands
- **Environment token** - Supports GH_TOKEN and other auth methods

### Repository Override (`repo_override.go`)
- **EnableRepoOverride()** - Adds `-R/--repo` flag support
- **Repository completion** - Tab completion for repository names
- **Context switching** - Work with different repositories than current directory

### Argument Processing (`args.go`, `flags.go`)
- **Flag validation** - Common flag patterns and validation
- **JSON output** - Structured output formatting
- **Export functionality** - Support for different output formats

## Usage Patterns
Commands should:
1. Accept `*cmdutil.Factory` in their NewCmd constructor
2. Use factory methods to access dependencies
3. Call `cmdutil.CheckAuth()` for commands requiring authentication
4. Use shared utilities for common operations

## Testing
- Factory can be mocked for unit tests
- Utilities are tested independently
- Integration tests use test factories with mocked dependencies