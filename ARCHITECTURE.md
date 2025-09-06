# GitHub CLI Architecture Documentation

This document provides comprehensive architecture diagrams for the GitHub CLI project using Mermaid notation.

## Overall Layered Architecture

```mermaid
graph TB
    subgraph "Entry Points Layer"
        CMD[cmd/gh/main.go]
        DOCS[cmd/gen-docs/main.go]
    end
    
    subgraph "Command Implementation Layer" 
        COMMANDS[pkg/cmd/*]
        ROOT[pkg/cmd/root/]
        ISSUE[pkg/cmd/issue/]
        PR[pkg/cmd/pr/]
        REPO[pkg/cmd/repo/]
        AUTH[pkg/cmd/auth/]
        OTHER[pkg/cmd/... 33 total]
    end
    
    subgraph "Command Utilities Layer"
        FACTORY[pkg/cmdutil/Factory]
        AUTHUTIL[pkg/cmdutil/auth]
        FLAGS[pkg/cmdutil/flags]
        REPOOVER[pkg/cmdutil/repo_override]
    end
    
    subgraph "Shared Packages Layer"
        IOSTREAMS[pkg/iostreams/]
        EXTENSIONS[pkg/extensions/]
        SEARCH[pkg/search/]
        JSON[pkg/jsoncolor/]
        SSH[pkg/ssh/]
        MARKDOWN[pkg/markdown/]
    end
    
    subgraph "API Layer"
        APICLIENT[api/client.go]
        QUERIES[api/queries_*.go]
        TYPES[api/types]
    end
    
    subgraph "Git Operations Layer"
        GITCLIENT[git/client.go]
        GITCMD[git/command.go]
        GITURL[git/url.go]
    end
    
    subgraph "Internal Packages Layer"
        GHREPO[internal/ghrepo/]
        CONFIG[internal/config/]
        PROMPTER[internal/prompter/]
        GHINSTANCE[internal/ghinstance/]
        BUILD[internal/build/]
    end
    
    subgraph "External Dependencies"
        COBRA[spf13/cobra]
        GITHUB[GitHub API]
        GITBIN[Git Binary]
        FILESYSTEM[File System]
    end

    %% Connections
    CMD --> ROOT
    DOCS --> ROOT
    ROOT --> COMMANDS
    COMMANDS --> FACTORY
    FACTORY --> IOSTREAMS
    FACTORY --> APICLIENT
    FACTORY --> GITCLIENT
    FACTORY --> CONFIG
    COMMANDS --> AUTHUTIL
    COMMANDS --> FLAGS
    APICLIENT --> QUERIES
    APICLIENT --> GITHUB
    GITCLIENT --> GITBIN
    COMMANDS --> GHREPO
    COMMANDS --> PROMPTER
    ROOT --> COBRA
    CONFIG --> FILESYSTEM
```

## Component Interaction Flow

```mermaid
graph LR
    subgraph "User Interface"
        CLI[CLI Command]
    end
    
    subgraph "Command Processing"
        COBRA[Cobra Router]
        CMD[Command Handler]
        OPTS[Options Struct]
    end
    
    subgraph "Dependency Injection"
        FACTORY[Factory]
        IO[IOStreams]
        HTTP[HTTP Client]
        GIT[Git Client]
        CFG[Config]
    end
    
    subgraph "Business Logic"
        AUTH[Auth Check]
        REPO[Repo Detection]
        API[API Calls]
        GITOPS[Git Operations]
    end
    
    subgraph "External Systems"
        GHAPI[GitHub API]
        GITREPO[Git Repository]
        FS[File System]
    end

    CLI --> COBRA
    COBRA --> CMD
    CMD --> FACTORY
    FACTORY --> IO
    FACTORY --> HTTP
    FACTORY --> GIT
    FACTORY --> CFG
    CMD --> OPTS
    OPTS --> AUTH
    OPTS --> REPO
    OPTS --> API
    OPTS --> GITOPS
    API --> GHAPI
    GITOPS --> GITREPO
    CFG --> FS
    IO --> FS
```

## Command Execution Sequence Diagram

```mermaid
sequenceDiagram
    participant User
    participant Main as cmd/gh/main.go
    participant GHCmd as internal/ghcmd
    participant Root as pkg/cmd/root
    participant Factory as pkg/cmdutil/Factory
    participant Command as pkg/cmd/issue/list
    participant Auth as pkg/cmdutil/auth
    participant API as api/client
    participant GitHub as GitHub API
    participant IO as pkg/iostreams

    User->>Main: gh issue list --limit 5
    Main->>GHCmd: Main()
    GHCmd->>Root: NewCmdRoot(factory)
    Root->>Factory: Initialize dependencies
    Factory-->>Root: Factory instance
    Root->>Command: Execute "issue list"
    Command->>Auth: CheckAuth(config)
    Auth-->>Command: Auth validated
    Command->>Factory: HttpClient()
    Factory-->>Command: HTTP Client
    Command->>API: GraphQL query for issues
    API->>GitHub: HTTP request
    GitHub-->>API: Issue data
    API-->>Command: Parsed issues
    Command->>IO: Format and display issues
    IO-->>User: Issue list output
    Command-->>Root: Success
    Root-->>GHCmd: Exit code 0
    GHCmd-->>Main: Exit code 0
    Main-->>User: Process complete
```

## API Interaction Sequence Diagram

```mermaid
sequenceDiagram
    participant Command as Command Handler
    participant Client as api/Client
    participant Auth as Authentication
    participant GraphQL as GraphQL Client
    participant REST as REST Client
    participant GitHub as GitHub API
    participant Cache as Local Cache

    Command->>Client: GraphQL("github.com", query, variables, &data)
    Client->>Auth: Get authentication token
    Auth-->>Client: Bearer token
    Client->>GraphQL: Execute query with auth headers
    
    alt GraphQL Request
        GraphQL->>GitHub: POST /graphql
        GitHub-->>GraphQL: JSON response
        GraphQL->>GraphQL: Parse response
        alt Has errors
            GraphQL-->>Client: GraphQLError
        else Success
            GraphQL-->>Client: Parsed data
        end
    else REST Request
        Command->>Client: REST("POST", "/repos/owner/repo/issues", data)
        Client->>REST: Execute with auth
        REST->>GitHub: POST /repos/owner/repo/issues
        GitHub-->>REST: JSON response
        alt Rate limit exceeded
            REST-->>Client: HTTPError with retry info
        else Success
            REST-->>Client: Response data
        end
    end
    
    Client->>Cache: Store response (optional)
    Client-->>Command: Final response data
    
    Note over Command,GitHub: All requests include GitHub CLI user agent<br/>and handle pagination automatically
```

## Authentication Flow Diagram

```mermaid
sequenceDiagram
    participant User
    participant Command as Command
    participant Auth as pkg/cmdutil/auth
    participant Config as internal/config
    participant Prompter as internal/prompter
    participant Browser as internal/browser
    participant GitHub as GitHub API
    participant Keyring as Keyring

    Command->>Auth: CheckAuth(config)
    Auth->>Config: Get authentication state
    
    alt Has environment token
        Config-->>Auth: GH_TOKEN found
        Auth-->>Command: Auth success
    else Has stored token
        Config->>Keyring: Get stored token
        Keyring-->>Config: Token retrieved
        Config-->>Auth: Token available
        Auth-->>Command: Auth success
    else No authentication
        Auth-->>Command: Auth required
        Command->>User: Display auth help message
        User->>Command: gh auth login
        Command->>Prompter: Select auth method
        Prompter-->>Command: Web browser selected
        Command->>Browser: Open GitHub auth URL
        Browser->>GitHub: OAuth flow
        GitHub-->>Browser: Authorization code
        Browser-->>Command: Auth complete
        Command->>GitHub: Exchange code for token
        GitHub-->>Command: Access token
        Command->>Config: Store token
        Config->>Keyring: Save token securely
        Keyring-->>Config: Token saved
        Command-->>User: Auth success
    end
    
    Note over User,Keyring: Supports multiple GitHub hosts<br/>and remembers auth per host
```

## Git Operations Sequence Diagram

```mermaid
sequenceDiagram
    participant Command as Command
    participant Factory as Factory
    participant GitClient as git/Client
    participant GitCmd as git/Command
    participant GitBinary as Git Binary
    participant Repo as Local Repository
    participant Remote as Remote Repository

    Command->>Factory: GitClient()
    Factory-->>Command: Git Client instance
    Command->>GitClient: CurrentBranch()
    GitClient->>GitCmd: Run("symbolic-ref --short HEAD")
    GitCmd->>GitBinary: Execute git command
    GitBinary->>Repo: Query repository state
    Repo-->>GitBinary: Current branch name
    GitBinary-->>GitCmd: "main"
    GitCmd-->>GitClient: "main"
    GitClient-->>Command: "main"
    
    Command->>GitClient: Push("origin", "feature-branch")
    GitClient->>GitCmd: Run("push origin feature-branch")
    GitCmd->>GitBinary: Execute push command
    GitBinary->>Remote: Push commits
    
    alt Push successful
        Remote-->>GitBinary: Success
        GitBinary-->>GitCmd: Exit code 0
        GitCmd-->>GitClient: Success
        GitClient-->>Command: Push complete
    else Push failed
        Remote-->>GitBinary: Rejection (e.g., non-fast-forward)
        GitBinary-->>GitCmd: Exit code 1 + stderr
        GitCmd-->>GitClient: GitError with details
        GitClient-->>Command: Error with suggested actions
    end
    
    Note over Command,Remote: Git operations are wrapped with<br/>proper error handling and context
```

## Factory Dependency Injection Pattern

```mermaid
graph TD
    subgraph "Factory Pattern"
        F[Factory Instance]
        F --> IO[IOStreams]
        F --> HTTP[HttpClient]
        F --> GIT[GitClient]  
        F --> CFG[Config]
        F --> BROWSER[Browser]
        F --> PROMPTER[Prompter]
        F --> BASEREPOFN[BaseRepo Function]
        F --> BRANCHFN[Branch Function]
        F --> REMOTESFN[Remotes Function]
    end
    
    subgraph "Command Uses Factory"
        CMD[Command Constructor]
        OPTS[Options Struct]
        RUNFUNC[Run Function]
    end
    
    subgraph "Runtime Dependencies"
        STDIN[Standard Input]
        STDOUT[Standard Output] 
        STDERR[Standard Error]
        HTTPAPI[HTTP API Client]
        GITREPO[Git Repository]
        CONFIGFILE[Config Files]
        WEBBROWSER[Web Browser]
        TERMINAL[Terminal Interface]
    end
    
    CMD --> F
    F --> OPTS
    OPTS --> RUNFUNC
    
    IO --> STDIN
    IO --> STDOUT
    IO --> STDERR
    HTTP --> HTTPAPI
    GIT --> GITREPO
    CFG --> CONFIGFILE
    BROWSER --> WEBBROWSER
    PROMPTER --> TERMINAL
```

## Complete Pull Request Creation Flow

```mermaid
sequenceDiagram
    participant User
    participant CMD as gh pr create
    participant Factory
    participant Git as git/Client
    participant API as api/Client
    participant Prompter as Prompter
    participant Editor as Text Editor
    participant GitHub as GitHub API

    User->>CMD: gh pr create --title "Fix bug"
    CMD->>Factory: Get dependencies
    Factory-->>CMD: GitClient, APIClient, Prompter, etc.
    
    CMD->>Git: CurrentBranch()
    Git-->>CMD: "fix-bug-123"
    
    CMD->>Git: Remotes()
    Git-->>CMD: [origin, upstream]
    
    CMD->>Git: Push("origin", "fix-bug-123")
    Git-->>CMD: Push successful
    
    alt Title not provided
        CMD->>Prompter: "Title for pull request:"
        Prompter-->>CMD: User input title
    end
    
    alt Body not provided  
        CMD->>Editor: Open editor for PR body
        Editor-->>CMD: PR body text
    end
    
    CMD->>API: CreatePullRequest(repo, title, body, head, base)
    API->>GitHub: POST /repos/owner/repo/pulls
    GitHub-->>API: PR created with #123
    API-->>CMD: PR data
    
    CMD->>User: Display PR URL and number
    
    Note over User,GitHub: Handles edge cases like:<br/>- Branch already has PR<br/>- No commits to push<br/>- Permission issues
```

## Extension System Architecture

```mermaid
graph TB
    subgraph "Core CLI"
        GHBIN[gh binary]
        EXTMGR[Extension Manager]
        EXTCMD[gh extension]
    end
    
    subgraph "Extension Discovery"
        GHREPO[GitHub Repository Search]
        LOCAL[Local Extensions]
        MANIFEST[Extension Manifests]
    end
    
    subgraph "Extension Types"
        PREBUILT[Prebuilt Binaries]
        SCRIPT[Shell Scripts]
        GOLANG[Go Extensions]
        OTHER[Other Languages]
    end
    
    subgraph "Extension Runtime"
        PATHEXEC[PATH Execution]
        ENVVARS[Environment Variables]
        AUTHTOKEN[Auth Token Passing]
        STDIO[STDIO Inheritance]
    end
    
    GHBIN --> EXTMGR
    EXTMGR --> GHREPO
    EXTMGR --> LOCAL
    EXTMGR --> MANIFEST
    
    EXTMGR --> PREBUILT
    EXTMGR --> SCRIPT  
    EXTMGR --> GOLANG
    EXTMGR --> OTHER
    
    EXTMGR --> PATHEXEC
    EXTMGR --> ENVVARS
    EXTMGR --> AUTHTOKEN
    EXTMGR --> STDIO
    
    EXTCMD --> EXTMGR
```

## Data Flow Summary

This architecture follows these key principles:

1. **Separation of Concerns**: Each layer has a distinct responsibility
2. **Dependency Injection**: Factory pattern provides testable dependencies
3. **Interface-Driven Design**: Components interact through well-defined interfaces
4. **Error Handling**: Comprehensive error handling with user-friendly messages
5. **Cross-Platform Support**: Abstractions handle platform differences
6. **Extensibility**: Plugin system allows community extensions
7. **Testing**: Architecture enables comprehensive unit and integration testing

The GitHub CLI successfully combines command-line usability with GitHub's rich API functionality through this well-structured, maintainable architecture.