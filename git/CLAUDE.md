# git/ - Git Operations Layer

This directory provides abstractions for interacting with local Git repositories and performing Git operations.

## Architecture
- **Git client** - High-level interface for Git operations
- **Command execution** - Safe execution of git commands with proper error handling
- **Repository introspection** - Query repository state, branches, remotes
- **URL parsing** - Parse and manipulate Git URLs

## Key Components

### Client (`client.go`)
High-level Git operations:
- **Repository queries** - Get current branch, remotes, repository root
- **Commit operations** - Create commits, get commit history
- **Branch management** - Switch branches, create/delete branches
- **Remote operations** - Fetch, push, clone operations

### Command Execution (`command.go`)
Safe Git command execution:
- **Process management** - Spawn git processes with proper environment
- **Error handling** - Parse git error messages and provide context
- **Input/Output** - Handle stdin/stdout for interactive git operations
- **Working directory** - Manage git operations in correct repository context

### Repository Context
- **Remote parsing** - Parse GitHub URLs from git remotes
- **Branch detection** - Identify current branch and default branches  
- **Repository root** - Find git repository boundaries
- **Configuration** - Read/write git config values

## Git Operations
Common operations provided:
- **Clone** - Clone repositories with GitHub-specific enhancements
- **Fetch/Pull** - Update local repository from remotes
- **Push** - Push changes to remote repositories
- **Branch** - Create, switch, and manage branches
- **Commit** - Create commits with proper authorship
- **Remote** - Manage remote repositories
- **Config** - Read and write git configuration

## Integration with GitHub
- **Remote URL parsing** - Extract GitHub repository information from git remotes
- **PR branch tracking** - Special handling for pull request branches
- **GitHub-specific config** - Store GitHub CLI-specific git config
- **Authentication** - Integrate with GitHub authentication for git operations

## Safety and Error Handling
- **Command validation** - Prevent execution of unsafe git commands
- **Error context** - Provide meaningful error messages for git failures
- **Repository detection** - Gracefully handle operations outside git repositories
- **Conflict resolution** - Handle merge conflicts and other git issues