# Autoreload Package Documentation

The `autoreload` package provides a simple WebSocket-based solution for automatically reloading web pages in Go web applications. It offers real-time communication between the server and client, allowing developers to trigger page reloads or send custom messages.

## Features

- WebSocket server for real-time communication
- Customizable JavaScript client for easy integration
- Broadcast functionality to trigger reloads or send custom messages

## Installation

To use the `autoreload` package in your project, add it to your Go module:

```
go get github.com/go-go-golems/clay/pkg/autoreload
```

## Usage

Here's a short tutorial on how to use the `autoreload` package in your Go web application:

1. Import the package:

```go
import "github.com/go-go-golems/clay/pkg/autoreload"
```

2. Create a new WebSocket server instance:

```go
wsServer := autoreload.NewWebSocketServer()
```

3. Set up the WebSocket handler:

```go
http.HandleFunc("/ws", wsServer.WebSocketHandler())
```

4. Serve the JavaScript snippet:

```go
http.HandleFunc("/autoreload.js", func(w http.ResponseWriter, r *http.Request) {
    js := wsServer.GetJavaScript("/ws")
    w.Header().Set("Content-Type", "application/javascript")
    w.Write([]byte(js))
})
```

5. Include the JavaScript in your HTML:

```html
<script src="/autoreload.js"></script>
```

6. Trigger reloads or send custom messages:

```go
wsServer.Broadcast("reload")
```

## Example

Here's a complete example demonstrating how to use the `autoreload` package:


```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-go-golems/clay/pkg/autoreload"
)

func main() {
	// Create a new WebSocket server instance
	wsServer := autoreload.NewWebSocketServer()

	// Set up the WebSocket handler
	http.HandleFunc("/ws", wsServer.WebSocketHandler())

	// Serve the JavaScript snippet at a specific endpoint
	http.HandleFunc("/autoreload.js", func(w http.ResponseWriter, r *http.Request) {
		js := wsServer.GetJavaScript("/ws")
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(js))
	})

	// Serve a simple HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Autoreload Demo</title>
    <script src="/autoreload.js"></script>
</head>
<body>
    <h1>Autoreload Demo</h1>
    <p>This page will reload automatically every 5 seconds.</p>
    <p>Current time: ` + time.Now().Format(time.RFC3339) + `</p>
</body>
</html>
`
		fmt.Fprint(w, html)
	})

	// Trigger a reload every 5 seconds
	go func() {
		for {
			time.Sleep(5 * time.Second)
			wsServer.Broadcast("reload")
		}
	}()

	// Start the HTTP server
	log.Println("Server started on http://localhost:8080")
```


This example sets up a simple web server that:

1. Creates a WebSocket server for autoreload functionality
2. Serves the autoreload JavaScript at `/autoreload.js`
3. Serves a simple HTML page at the root URL
4. Triggers a reload every 5 seconds using a goroutine

To run this example:

1. Save the code in a file named `main.go`
2. Run the following command:

```
go run main.go
```

3. Open a web browser and navigate to `http://localhost:8080`

You should see a page that automatically reloads every 5 seconds, updating the current time displayed.

## Customization

The `autoreload` package allows for customization of the WebSocket endpoint and behavior. You can modify the WebSocket URL when getting the JavaScript snippet:

```go
js := wsServer.GetJavaScript("/custom-ws-endpoint")
```

You can also send custom messages instead of just "reload":

```go
wsServer.Broadcast("custom-message")
```

To handle these custom messages, you'll need to modify the JavaScript code accordingly.
