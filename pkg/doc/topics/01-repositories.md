---
Title: Working with Clay Repositories
Slug: repositories
Short: Learn how to manage and organize commands in a hierarchical structure using Clay's repository system
Topics:
- repositories
- commands
- organization
- file system watching
Commands:
- NewRepository
- LoadCommands
- Watch
- CollectCommands
- NewMultiRepository
- Mount
Flags:
- none
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Repositories

The repositories package (`github.com/go-go-golems/clay/pkg/repositories`) provides a flexible way to manage and organize commands in a hierarchical structure, with support for file system watching and dynamic updates.

## Overview

The repository system is built around a common interface that allows different repository implementations, including:
- Basic repository with trie-based storage
- Multi-repository that can mount other repositories at specific paths
- Custom repository implementations for specialized needs

The system supports:
- Loading commands from directories and files
- Watching directories for changes
- Dynamic command updates
- Command aliasing
- Hierarchical organization
- Mounting repositories at specific paths

## Repository Interface

The repository system is built around the `RepositoryInterface` which defines the core functionality that all repository implementations must provide:

```go
type RepositoryInterface interface {
    // LoadCommands initializes the repository by loading all commands
    LoadCommands(helpSystem *help.HelpSystem, options ...cmds.CommandDescriptionOption) error

    // Add adds one or more commands to the repository
    Add(commands ...cmds.Command)

    // Remove removes commands with the given prefixes from the repository
    Remove(prefixes ...[]string)

    // CollectCommands returns all commands under a given prefix
    CollectCommands(prefix []string, recurse bool) []cmds.Command

    // GetCommand returns a single command by its full path name
    GetCommand(name string) (cmds.Command, bool)

    // FindNode returns the TrieNode at the given prefix
    FindNode(prefix []string) *trie.TrieNode

    // GetRenderNode returns a RenderNode for visualization purposes
    GetRenderNode(prefix []string) (*trie.RenderNode, bool)

    // ListTools returns all commands as tools for MCP compatibility
    ListTools(ctx context.Context, cursor string) ([]mcp.Tool, string, error)

    // Watch sets up file system watching for the repository
    Watch(ctx context.Context, options ...watcher.Option) error
}
```

### Core Operations

The interface provides several key operations:

1. Command Management:
   - `LoadCommands`: Initialize the repository with commands
   - `Add`: Add new commands to the repository
   - `Remove`: Remove commands by their prefix paths

2. Command Retrieval:
   - `CollectCommands`: Get all commands under a prefix
   - `GetCommand`: Find a specific command by its full path
   - `FindNode`: Access the underlying trie structure
   - `GetRenderNode`: Get a visualization-friendly representation

3. Tool Integration:
   - `ListTools`: Convert commands to MCP-compatible tools

4. File System Watching:
   - `Watch`: Set up file system watching for dynamic updates

### File System Watching

The repository system supports dynamic updates through file system watching:

```go
// Set up watching with options
err := repo.Watch(ctx, 
    watcher.WithMask("**/*.yaml"),  // Only watch yaml files
    watcher.WithBreakOnError(false), // Continue on errors
)
```

The watcher will:
- Monitor specified directories and files for changes
- Automatically reload commands when files are modified
- Remove commands when files are deleted
- Support filtering by file patterns
- Handle errors gracefully

### Implementing Custom Repositories

You can create custom repository implementations by implementing the `RepositoryInterface`. Common use cases include:

```go
// Example custom repository
type CustomRepository struct {
    // Your custom fields
}

// Implement interface methods
func (c *CustomRepository) LoadCommands(helpSystem *help.HelpSystem, 
    options ...cmds.CommandDescriptionOption) error {
    // Your implementation
    return nil
}

func (c *CustomRepository) Add(commands ...cmds.Command) {
    // Your implementation
}

// ... implement other methods ...

func (c *CustomRepository) Watch(ctx context.Context, options ...watcher.Option) error {
    // Implement file system watching if needed
    return nil
}
```

### Available Implementations

The package provides two main implementations:

1. `Repository`: The standard implementation using a trie data structure
   ```go
   repo := repositories.NewRepository(
       repositories.WithDirectories(...),
       repositories.WithFiles(...),
   )
   ```

2. `MultiRepository`: An implementation that can mount other repositories
   ```go
   mr := repositories.NewMultiRepository()
   mr.Mount("/path", someRepository)
   ```

## Basic Usage

### Creating a Repository

```go
import (
    "github.com/go-go-golems/clay/pkg/repositories"
    "github.com/go-go-golems/glazed/pkg/help"
)

// Create a new repository with both directories and individual files
repo := repositories.NewRepository(
    repositories.WithDirectories(repositories.Directory{
        FS:               os.DirFS("/path/to/commands"),
        RootDirectory:    ".",
        RootDocDirectory: "doc",
        WatchDirectory:   "/path/to/commands",
        Name:            "my-commands",
        SourcePrefix:    "file",
    }),
    repositories.WithFiles([]string{
        "/path/to/specific/command1.yaml",
        "/path/to/specific/command2.yaml",
    }),
)

// Initialize help system
helpSystem := help.NewHelpSystem()

// Load commands
err := repo.LoadCommands(helpSystem)
if err != nil {
    log.Fatal(err)
}

// Set up watching
ctx := context.Background()
go func() {
    err := repo.Watch(ctx)
    if err != nil {
        log.Printf("Watch error: %v", err)
    }
}()
```

### Loading Commands from Multiple Sources

You can load commands from both directories and individual files:

```go
repo := repositories.NewRepository(
    repositories.WithDirectories(
        repositories.Directory{
            FS:               os.DirFS("/path/to/commands1"),
            RootDirectory:    ".",
            Name:            "commands1",
        },
    ),
    repositories.WithFiles([]string{
        "/path/to/command1.yaml",
        "/path/to/command2.yaml",
    }),
)
```

## Multi-Repository Support

The multi-repository allows mounting multiple repositories under different paths, creating a unified command hierarchy:

```go
import (
    "github.com/go-go-golems/clay/pkg/repositories"
)

// Create individual repositories
baseRepo := repositories.NewRepository(...)
toolsRepo := repositories.NewRepository(...)
pluginsRepo := repositories.NewRepository(...)

// Create multi-repository
mr := repositories.NewMultiRepository()

// Mount repositories at different paths
mr.Mount("/", baseRepo)           // Root mount
mr.Mount("/tools", toolsRepo)     // Tools under /tools
mr.Mount("/plugins", pluginsRepo) // Plugins under /plugins

// Load commands from all repositories
err := mr.LoadCommands(helpSystem)
if err != nil {
    log.Fatal(err)
}

// Set up watching for all repositories
go func() {
    err := mr.Watch(ctx)
    if err != nil {
        log.Printf("Watch error: %v", err)
    }
}()
```

### Path Handling in Multi-Repositories

Multi-repositories handle paths differently based on mount points:

1. Root-mounted repositories (`/`):
   - Commands maintain their original paths
   - No path prefix is added
   - Example: `my-command` stays as `my-command`

2. Path-mounted repositories:
   - Mount path is prepended to all commands
   - Commands are only accessible through their full path
   - Example: Command `build` in `/tools` becomes `tools/build`

```go
// Access commands through their full paths
rootCmd, found := mr.GetCommand("my-command")        // From root repository
toolCmd, found := mr.GetCommand("tools/build")       // From tools repository
pluginCmd, found := mr.GetCommand("plugins/my-plugin") // From plugins repository

// Collect commands from specific paths
toolCommands := mr.CollectCommands([]string{"tools"}, true)  // All commands under /tools
```

### Collecting Commands

The repository allows you to collect commands by prefix:

```go
// Get all commands from all mounted repositories
allCommands := mr.CollectCommands([]string{}, true)

// Get commands under specific prefix
subCommands := mr.CollectCommands([]string{"group", "subgroup"}, true)

// Get commands without recursion
directCommands := mr.CollectCommands([]string{"group"}, false)
```

## File System Watching

The repository system can watch both directories and individual files for changes and automatically update commands.

### Setting Up a Watcher

```go
import (
    "context"
    "github.com/go-go-golems/clay/pkg/watcher"
)

ctx := context.Background()
err := repo.Watch(ctx, 
    watcher.WithMask("**/*.yaml"),  // Only watch yaml files
    watcher.WithBreakOnError(false), // Continue on errors
)
if err != nil {
    log.Fatal(err)
}
```

The watcher will automatically:
- Load new commands when files are created or modified
- Remove commands when files are deleted
- Update the repository's command tree accordingly
- Efficiently watch individual files by monitoring their parent directories
- Only trigger on events for specifically watched files when monitoring individual files

### Custom Watch Callbacks

You can customize the watch behavior by providing additional options:

```go
repo.Watch(ctx,
    watcher.WithWriteCallback(func(path string) error {
        log.Printf("File changed: %s", path)
        return nil
    }),
    watcher.WithRemoveCallback(func(path string) error {
        log.Printf("File removed: %s", path)
        return nil
    }),
)
```

## Command Organization

Commands in a repository are organized in a trie structure, allowing for efficient lookup and hierarchical organization.

### Command Paths

Commands are identified by their path components:

```go
// Command at root
repo.InsertCommand([]string{}, rootCommand)

// Command in group
repo.InsertCommand([]string{"group"}, groupCommand)

// Command in nested group
repo.InsertCommand([]string{"group", "subgroup"}, subCommand)
```

### Finding Commands

```go
// Find a specific command
cmd, found := repo.GetCommand("group/subgroup/command")

// Find a node in the command tree
node := repo.FindNode([]string{"group", "subgroup"})
```

## Advanced Features

### Command Removal

```go
// Remove a specific command
removedCmds := repo.Remove([]string{"group", "subgroup", "command"})

// Remove all commands under a prefix
removedCmds := repo.Remove([]string{"group"})
```

### Loading from Multiple Inputs

The package provides helper functions for loading commands from multiple sources:

```go
commands, err := repositories.LoadCommandsFromInputs(
    commandLoader,
    []string{
        "/path/to/command1.yaml",
        "/path/to/commands/directory",
    },
)
```

### Integration with Cobra

The package includes helpers for integrating with Cobra command-line applications:

```go
rootCmd := &cobra.Command{}
commands, err := repositories.LoadRepositories(
    helpSystem,
    rootCmd,
    []*repositories.Repository{repo},
    // Additional cobra parser options...
)
```

## Common Patterns

### Repository with Auto-reload

```go
repo := repositories.NewRepository(
    repositories.WithCommandLoader(myLoader),
    repositories.WithDirectories(myDirectory),
)

// Load initial commands
err := repo.LoadCommands(helpSystem)
if err != nil {
    log.Fatal(err)
}

// Start watching for changes
go func() {
    err := repo.Watch(ctx)
    if err != nil {
        log.Printf("Watch error: %v", err)
    }
}()
```

### Multi-Repository with Mixed Sources

```go
// Create repositories with different sources
localRepo := repositories.NewRepository(
    repositories.WithDirectories(repositories.Directory{
        FS:            os.DirFS("/local/commands"),
        RootDirectory: ".",
    }),
)

pluginRepo := repositories.NewRepository(
    repositories.WithFiles([]string{
        "/plugins/plugin1.yaml",
        "/plugins/plugin2.yaml",
    }),
)

// Create and configure multi-repository
mr := repositories.NewMultiRepository()
mr.Mount("/", localRepo)
mr.Mount("/plugins", pluginRepo)

// Watch both repositories
go func() {
    err := mr.Watch(ctx)
    if err != nil {
        log.Printf("Watch error: %v", err)
    }
}()
```

This setup creates a repository that automatically reloads commands when files change, making it ideal for development environments or dynamic command systems.
