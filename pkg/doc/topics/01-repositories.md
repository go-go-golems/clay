# Repositories

The repositories package (`github.com/go-go-golems/clay/pkg/repositories`) provides a flexible way to manage and organize commands in a hierarchical structure, with support for file system watching and dynamic updates.

## Overview

The repository system is built around a trie data structure that organizes commands by their path components. It supports:

- Loading commands from directories and files
- Watching directories for changes
- Dynamic command updates
- Command aliasing
- Hierarchical organization

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

### Collecting Commands

The repository allows you to collect commands by prefix:

```go
// Get all commands
allCommands := repo.CollectCommands([]string{}, true)

// Get commands under specific prefix
subCommands := repo.CollectCommands([]string{"group", "subgroup"}, true)

// Get commands without recursion
directCommands := repo.CollectCommands([]string{"group"}, false)
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

This setup creates a repository that automatically reloads commands when files change, making it ideal for development environments or dynamic command systems.
