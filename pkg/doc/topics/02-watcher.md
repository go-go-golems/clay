# Watcher

The Watcher package provides a simple and flexible way to watch for file system changes in Go applications. It recursively watches directories and can filter events based on file patterns.

## Features

- Recursive directory watching
- Individual file watching with parent directory tracking
- File pattern filtering using doublestar masks
- Customizable callbacks for write and remove events
- Configurable error handling

## Installation

```bash
go get github.com/your-repo/clay/pkg/watcher
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/your-repo/clay/pkg/watcher"
)

func main() {
    w := watcher.NewWatcher(
        watcher.WithPaths("./path/to/watch"),
        watcher.WithMask("**/*.txt"),
        watcher.WithWriteCallback(func(path string) error {
            fmt.Printf("File written: %s\n", path)
            return nil
        }),
        watcher.WithRemoveCallback(func(path string) error {
            fmt.Printf("File removed: %s\n", path)
            return nil
        }),
    )

    ctx := context.Background()
    if err := w.Run(ctx); err != nil {
        fmt.Printf("Watcher error: %v\n", err)
    }
}
```

### Recursive Directory Watching

The Watcher package can recursively watch directories for changes:

```go
w := watcher.NewWatcher(
    watcher.WithPaths("./path/to/watch"),
)
```

This feature allows you to monitor an entire directory tree for changes, automatically including new subdirectories as they are created.

### Individual File Watching

The Watcher can efficiently watch individual files by monitoring their parent directories:

```go
w := watcher.NewWatcher(
    watcher.WithPaths(
        "./config/app.yaml",
        "./config/db.yaml",
    ),
)
```

When watching individual files:
- The watcher monitors the parent directory
- Only events for the specified files trigger callbacks
- Other files in the same directory are ignored
- Directory watching is handled automatically

This is more efficient than watching individual files directly, especially on systems with inotify limits.

### File Pattern Filtering

You can filter events based on file patterns using doublestar masks:

```go
w := watcher.NewWatcher(
    watcher.WithPaths("./path/to/watch"),
    watcher.WithMask("**/*.txt"),
)
```

This allows you to focus on specific file types or patterns, ignoring changes to files that don't match the specified mask.

### Customizable Callbacks

The package provides separate callbacks for write and remove events:

```go
w := watcher.NewWatcher(
    watcher.WithWriteCallback(func(path string) error {
        fmt.Printf("File written: %s\n", path)
        return nil
    }),
    watcher.WithRemoveCallback(func(path string) error {
        fmt.Printf("File removed: %s\n", path)
        return nil
    }),
)
```

These callbacks allow you to define custom behavior when files are written to or removed from the watched directories.

### Configurable Error Handling

You can configure how the watcher handles errors:

```go
w := watcher.NewWatcher(
    watcher.WithBreakOnError(true),
)
```

This option allows you to decide whether the watcher should stop on the first error encountered or continue running despite errors.

### Easy Setup with Functional Options

The package uses the functional options pattern for easy and flexible configuration:

```go
w := watcher.NewWatcher(
    watcher.WithPaths("./path1", "./path2"),
    watcher.WithMask("**/*.go", "**/*.txt"),
    watcher.WithWriteCallback(writeHandler),
    watcher.WithRemoveCallback(removeHandler),
    watcher.WithBreakOnError(false),
)
```

This approach allows for clear and concise setup of the watcher with multiple options.

### Context-Based Execution

The watcher runs with a context, allowing for graceful shutdown:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

if err := w.Run(ctx); err != nil {
    log.Printf("Watcher error: %v\n", err)
}
```

This feature enables easy integration with Go's context package for managing the watcher's lifecycle.

## Tutorial: Building a Simple File Change Logger

1. Create a new Go file named `file_logger.go`:

```go
package main

import (
    "context"
    "fmt"
    "github.com/your-repo/clay/pkg/watcher"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    w := watcher.NewWatcher(
        watcher.WithPaths("./logs"),
        watcher.WithMask("**/*.log"),
        watcher.WithWriteCallback(logFileChange),
        watcher.WithRemoveCallback(logFileRemoval),
    )

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    if err := w.Run(ctx); err != nil && err != context.Canceled {
        fmt.Printf("Watcher error: %v\n", err)
    }
}

func logFileChange(path string) error {
    fmt.Printf("File changed: %s\n", path)
    return nil
}

func logFileRemoval(path string) error {
    fmt.Printf("File removed: %s\n", path)
    return nil
}
```

2. Create a `logs` directory in the same folder as your Go file.

3. Run the program:

```bash
go run file_logger.go
```

4. In another terminal, create, modify, and delete log files in the `logs` directory to see the watcher in action.

This example demonstrates how to use the Watcher package to monitor a specific directory for changes to log files, logging any modifications or removals to the console.