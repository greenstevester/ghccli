# pkg/ - Shared Package Layer

This directory contains reusable packages that provide common functionality across the GitHub CLI application.

## Architecture
- **Reusable utilities** - Packages that can be imported by multiple components
- **Interface abstractions** - Common interfaces used throughout the application
- **Cross-cutting concerns** - Functionality needed by multiple layers
- **Testable components** - Well-isolated packages with clear boundaries

## Key Packages

### I/O Streams (`iostreams/`)
- **Testable I/O** - Abstraction over stdin, stdout, stderr for unit testing
- **Terminal detection** - Detect TTY capabilities and terminal features
- **Pager support** - Automatic paging for long output
- **Color support** - Cross-platform color handling
- **Spinner integration** - Progress indicators for long-running operations

### Extensions (`extensions/`)
- **Extension management** - Install, update, and manage GitHub CLI extensions
- **Extension discovery** - Find and list available extensions
- **Extension execution** - Run extension commands with proper context
- **Extension metadata** - Parse and validate extension manifests

### Search Utilities (`search/`)
- **Query building** - Construct GitHub search queries
- **Result processing** - Parse and format search results
- **Search types** - Support for different GitHub search endpoints
- **Filtering and sorting** - Post-process search results

### JSON Handling (`jsoncolor/`)
- **Syntax highlighting** - Colorize JSON output for better readability
- **Pretty printing** - Format JSON with proper indentation
- **Color themes** - Support different color schemes
- **Terminal compatibility** - Graceful fallback for non-color terminals

### SSH Utilities (`ssh/`)
- **SSH key management** - Generate, validate, and format SSH keys
- **SSH config parsing** - Parse SSH configuration files
- **Key authentication** - Integrate with GitHub SSH key authentication
- **Cross-platform support** - Handle SSH across different operating systems

### Markdown Processing (`markdown/`)
- **Markdown rendering** - Convert markdown to terminal-friendly output
- **Syntax highlighting** - Highlight code blocks in markdown
- **Link handling** - Process and display markdown links
- **Terminal formatting** - Adapt markdown for terminal display

## Usage Patterns
These packages are used by:
- **Commands** (`pkg/cmd/`) for specific functionality
- **API layer** (`api/`) for data processing
- **Command utilities** (`pkg/cmdutil/`) for shared operations
- **Internal packages** (`internal/`) for common tasks

## Design Principles
- **Single responsibility** - Each package has a focused purpose
- **Minimal dependencies** - Packages avoid heavy external dependencies
- **Interface-driven** - Packages expose clean interfaces for extensibility
- **Testing support** - Packages include testing utilities and mocks
- **Cross-platform** - Packages work consistently across supported platforms

## Testing
- **Unit testable** - Packages are designed for easy unit testing
- **Mock interfaces** - Provide mock implementations for testing
- **Test utilities** - Include helpers for testing dependent code
- **Integration tests** - Test interactions between packages