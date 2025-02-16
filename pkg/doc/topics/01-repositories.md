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

## Command Repository

The `CommandRepository` provides a lightweight, in-memory implementation of the repository interface that focuses solely on command organization without any file system dependencies. It's ideal for scenarios where you just need to manage and organize commands programmatically.

### Creating a Command Repository

```go
import "github.com/go-go-golems/clay/pkg/repositories"

// Create a basic command repository
repo := repositories.NewCommandRepository()

// Create with a name
repo := repositories.NewCommandRepository(
    repositories.WithCommandRepositoryName("my-commands"),
)
```

### Adding Commands

The CommandRepository provides two ways to add commands:

1. Basic Add - uses the command's existing path:
```go
// Commands will be organized based on their Description().Parents
repo.Add(command1, command2)
```

2. AddUnderPath - adds commands under a specific path:
```go
// Add commands under a custom path
repo.AddUnderPath([]string{"group", "subgroup"}, command1, command2)

// Commands will be accessible as "group/subgroup/command1" etc.
```

### Command Organization

Commands are organized in a trie structure just like the full Repository:

```go
// Add commands at different levels
repo.Add(rootCommand)  // At root
repo.AddUnderPath([]string{"tools"}, toolCommand)  // Under "tools"
repo.AddUnderPath([]string{"tools", "network"}, networkCommand)  // Nested

// Retrieve commands
allCommands := repo.CollectCommands([]string{}, true)  // All commands
toolCommands := repo.CollectCommands([]string{"tools"}, true)  // All tool commands
networkTools := repo.CollectCommands([]string{"tools", "network"}, false)  // Direct network tools
```

### Key Features

The CommandRepository provides:

1. Pure in-memory storage:
   - No file system dependencies
   - No watching functionality
   - Fast and lightweight

2. Full path-based organization:
   - Hierarchical command structure
   - Flexible path-based access
   - Support for nested command groups

3. Simple API:
   - Basic Add for default paths
   - AddUnderPath for custom organization
   - Standard collection and lookup methods

4. RepositoryInterface compatibility:
   - Works with MultiRepository
   - Compatible with all repository-based tools
   - Supports visualization and rendering

### Common Use Cases

The CommandRepository is particularly useful for:

1. Testing and Development:
```go
// Create a test repository
testRepo := repositories.NewCommandRepository()
testRepo.Add(mockCommands...)
```

2. Dynamic Command Generation:
```go
// Create commands programmatically
repo := repositories.NewCommandRepository()
for _, config := range configs {
    cmd := createCommandFromConfig(config)
    repo.AddUnderPath([]string{"generated", config.Type}, cmd)
}
```

3. Temporary Command Organization:
```go
// Create a temporary command structure
tempRepo := repositories.NewCommandRepository()
tempRepo.AddUnderPath([]string{"session", sessionID}, sessionCommands...)
```

4. Plugin Systems:
```go
// Add plugin commands under their own namespace
pluginRepo := repositories.NewCommandRepository()
for _, plugin := range plugins {
    pluginRepo.AddUnderPath([]string{"plugins", plugin.Name}, plugin.Commands...)
}
```

### Differences from Full Repository

The main differences from the full Repository implementation are:

1. No File System Integration:
   - No LoadCommands implementation (returns nil)
   - No Watch support (no-op implementation)
   - No file-based command loading

2. Simplified Implementation:
   - No directory management
   - No help system integration
   - No file watching callbacks

3. Focus on Command Management:
   - Pure in-memory storage
   - Direct command manipulation
   - Path-based organization

### Example: Building a Command Menu

Here's an example of using CommandRepository to build a hierarchical command menu:

```go
// Create the repository
menuRepo := repositories.NewCommandRepository(
    repositories.WithCommandRepositoryName("menu"),
)

// Add commands in categories
menuRepo.AddUnderPath([]string{"file"}, 
    newCommand, openCommand, saveCommand)
menuRepo.AddUnderPath([]string{"edit"},
    cutCommand, copyCommand, pasteCommand)
menuRepo.AddUnderPath([]string{"view", "zoom"},
    zoomInCommand, zoomOutCommand, resetZoomCommand)

// Get commands for a specific menu
fileCommands := menuRepo.CollectCommands([]string{"file"}, false)
zoomCommands := menuRepo.CollectCommands([]string{"view", "zoom"}, false)

// Get all commands
allCommands := menuRepo.CollectCommands([]string{}, true)
```
