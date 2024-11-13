# clay


```
 _______  ___      _______  __   __ 
|       ||   |    |   _   ||  | |  |
|       ||   |    |  |_|  ||  |_|  |
|       ||   |    |       ||       |
|      _||   |___ |       ||_     _|
|     |_ |       ||   _   |  |   |  
|_______||_______||__| |__|  |___|  
 __   __  _______  ___   _  _______  _______ 
|  |_|  ||   _   ||   | | ||       ||       |
|       ||  |_|  ||   |_| ||    ___||  _____|
|       ||       ||      _||   |___ | |_____ 
|       ||       ||     |_ |    ___||_____  |
| ||_|| ||   _   ||    _  ||   |___  _____| |
|_|   |_||__| |__||___| |_||_______||_______|
 _______  _______  ___      _______  __   __  _______ 
|       ||       ||   |    |       ||  |_|  ||       |
|    ___||   _   ||   |    |    ___||       ||  _____|
|   | __ |  | |  ||   |    |   |___ |       || |_____ 
|   ||  ||  |_|  ||   |___ |    ___||       ||_____  |
|   |_| ||       ||       ||   |___ | ||_|| | _____| |
|_______||_______||_______||_______||_|   |_||_______|
```

![potter](https://user-images.githubusercontent.com/128441/217436490-f24c8ab4-0202-4091-9f0e-db3e97729934.jpg)


# Clay

<p align="center">
  <img src="https://user-images.githubusercontent.com/128441/217436490-f24c8ab4-0202-4091-9f0e-db3e97729934.jpg" alt="Clay Potter" width="400"/>
</p>

![](https://img.shields.io/github/license/go-go-golems/glazed)
![](https://img.shields.io/github/actions/workflow/status/go-go-golems/glazed/push.yml?branch=main)


Clay is a collection of foundational Go packages that form the building blocks for go-go-golems projects. Like the material used to craft golems, these packages are moldable, versatile, and ready to be shaped into powerful applications.

## Overview

Clay provides essential utilities and helper packages that solve common development challenges. While originally created for go-go-golems projects, these packages are designed to be useful for any Go application requiring robust file handling, SQL operations, worker pools, or repository management.

## Packages

### 🔄 Autoreload (`pkg/autoreload`)
A WebSocket-based solution for automatically reloading web pages. Perfect for development environments where you want instant feedback on changes.

Key concepts:
- Single WebSocket server handles multiple client connections
- Configurable client-side JavaScript with customizable endpoints
- Broadcast system for server-initiated actions
- Support for both page reloads and custom message handling

```go
// Basic setup
wsServer := autoreload.NewWebSocketServer()
http.HandleFunc("/ws", wsServer.WebSocketHandler())

// Advanced usage
server := autoreload.NewWebSocketServer()
server.Broadcast("reload")  // Triggers page reload
js := server.GetJavaScript("/ws")  // Get embeddable client code
```

### 📂 FileFilter (`pkg/filefilter`)
Powerful and flexible file filtering system that combines multiple filtering strategies that can be used together or separately. Designed to integrate well with Git repositories and support common development patterns.

Features:
- File size limits
- Extension filtering
- Pattern matching
- GitIgnore integration
- Binary file detection
- Filters can be composed using functional options
- Profile-based configuration for different filtering scenarios

```go
ff := filefilter.NewFileFilter(
    filefilter.WithMaxFileSize(5 * 1024 * 1024),
    filefilter.WithIncludeExts([]string{".go", ".md"}),
)
```

### 🚶 FileWalker (`pkg/filewalker`)
Advanced file system traversal with AST-like structure representation. Features include:
- Recursive directory walking
- Symlink handling
- Customizable filters
- Path-based node retrieval

```go
walker, _ := filewalker.NewWalker(
    filewalker.WithFilter(filterFunc),
    filewalker.WithFollowSymlinks(true),
)
```

### 📝 Memoization (`pkg/memoization`)
Generic memoization implementation with LRU cache support. Useful for caching expensive operations with configurable capacity and eviction policies.

```go
cache := memoization.NewMemoCache[HString, interface{}](100)
cache.Set("key", value)
```

### 📚 Repositories (`pkg/repositories`)
Implements a trie-based command management system, designed to organize and handle hierarchical command structures in CLI applications. 

Features:
- Dynamic loading/reloading
- Hierarchical organization
- Callback-based updates
- File system watching
- Trie structure for efficient prefix-based command lookup
- Support for command aliases and parent-child relationships
- Integration with helpSystem for documentation

```go
repo := repositories.NewRepository(
    repositories.WithCommandLoader(loader),
    repositories.WithUpdateCallback(callback),
)
```

### 🗃️ SQL (`pkg/sql`)
Comprehensive SQL database utilities providing a higher-level interface for database operations in Go, with special attention to features needed in modern development workflows.

Features:
- Connection management
- Query templating
- DBT profile support
- Connection pooling
- Unified configuration handling for multiple database types
- Template functions for safe query construction
- Integration with dbt project structures
- Query helper functions like `sqlStringIn`, `sqlDate`, etc.

```go
// Basic configuration
config := &sql.DatabaseConfig{
    Host: "localhost",
    Port: 5432,
    Database: "mydb",
}
db, _ := config.Connect()

// Template-based query creation
query, err := sql.RenderQuery(
    ctx,
    db,
    "SELECT * FROM users WHERE created_at > {{ sqlDate .start_date }}",
    map[string]interface{}{
        "start_date": "2023-01-01",
    },
)

// DBT profile support
config := &sql.DatabaseConfig{
    UseDbtProfiles: true,
    DbtProfile: "analytics_profile",
}
```

### 👀 Watcher (`pkg/watcher`)
File system watching utility with rich features:
- Recursive directory monitoring
- Pattern-based filtering
- Event callbacks
- Graceful error handling

```go
w := watcher.NewWatcher(
    watcher.WithPaths("./path/to/watch"),
    watcher.WithMask("**/*.go"),
)
```

### 👷 WorkerPool (`pkg/workerpool`)
Simple yet powerful worker pool implementations for parallel task processing. Implements concurrency patterns for handling parallel task execution with a fixed number of workers.

Features:
- Basic worker pool for error-returning jobs
- Generic map pool for jobs with results
- Fixed-size worker pools with controlled concurrency
- Channel-based job distribution
- Graceful shutdown handling

```go
pool := workerpool.New(4)
pool.Start()
pool.AddJob(func() error { /* ... */ })
```

## Installation

To use Clay in your project:

```bash
go get github.com/go-go-golems/clay
```

You can import individual packages as needed:

```go
import (
    "github.com/go-go-golems/clay/pkg/filewalker"
    "github.com/go-go-golems/clay/pkg/watcher"
    // ... other packages
)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Clay is released under the MIT License. See the [LICENSE](LICENSE) file for details.

## Related Projects

Check out other go-go-golems projects that build upon Clay:
- [Glazed](https://github.com/go-go-golems/glazed): A powerful CLI application framework
- [Sqleton](https://github.com/go-go-golems/sqleton): SQL query and database management tools