# internal/ - Internal Packages Layer

This directory contains Go packages that are specific to the GitHub CLI project and should not be imported by external projects.

## Architecture
- **Domain-specific utilities** - Packages tailored to GitHub CLI's specific needs
- **Internal abstractions** - Types and interfaces used across the project
- **Platform-specific code** - OS and environment-specific implementations
- **Private implementations** - Implementation details not exposed as public API

## Key Packages

### Repository Handling (`ghrepo/`)
- **Repository interface** - Abstract representation of GitHub repositories
- **URL parsing** - Parse GitHub repository URLs from various formats
- **Repository identification** - Determine repository context from git remotes
- **Host normalization** - Handle GitHub.com and GitHub Enterprise Server

### Configuration (`config/`, `gh/`)
- **Configuration management** - Read/write CLI configuration files
- **Authentication state** - Manage GitHub authentication tokens
- **Host-specific settings** - Per-GitHub-instance configuration
- **Default behaviors** - Global CLI preferences and defaults

### User Interface (`prompter/`)
- **Interactive prompts** - User input collection with validation
- **Accessibility support** - Screen reader and accessibility-friendly prompts
- **Testing utilities** - Mock prompter for automated testing
- **Cross-platform input** - Handle different terminal capabilities

### GitHub Instance (`ghinstance/`)
- **Instance detection** - Identify GitHub.com vs Enterprise Server
- **Host normalization** - Standardize hostname handling
- **Feature detection** - Detect available GitHub features by instance
- **API endpoint resolution** - Map hosts to API endpoints

### Text Processing (`text/`)
- **Text formatting** - String manipulation and formatting utilities
- **Template processing** - Handle text templates and placeholders
- **Markdown processing** - Parse and format markdown content
- **Truncation and wrapping** - Handle text display in terminal

### Build Information (`build/`)
- **Version information** - Compile-time version and build date
- **Build metadata** - Git commit, build environment information
- **Feature flags** - Build-time feature toggles
- **Release information** - Version comparison and update detection

## Usage Patterns
Internal packages are used by:
- Commands in `pkg/cmd/` for domain-specific logic
- Utilities in `pkg/cmdutil/` for shared functionality  
- API layer for GitHub-specific processing
- Main entry points for application setup

## Design Principles
- **No external imports** - These packages should not be used outside this project
- **Focused responsibility** - Each package has a single, clear purpose
- **Testability** - Packages are designed to be easily unit tested
- **Abstraction** - Provide clean interfaces for complex operations