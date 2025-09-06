# cmd/ - Entry Points Layer

This directory contains the main entry points for building executable binaries.

## Architecture
- **`gh/`** - Main entry point for the `gh` executable
- **`gen-docs/`** - Utility for generating documentation from command help text

## Key Files
- **`gh/main.go`** - Minimal main function that delegates to `internal/ghcmd.Main()`
- **`gen-docs/main.go`** - Documentation generation tool that processes command help text

## Patterns
- Entry points are kept minimal and delegate to internal packages
- Main functions handle process exit codes and error reporting
- No business logic should be implemented here - only process lifecycle management

## Usage
- Built via `go run script/build.go` or `make`
- The `gh` binary becomes the primary user interface
- `gen-docs` is used during release process to generate manual pages