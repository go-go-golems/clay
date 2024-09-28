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
		_, _ = w.Write([]byte(js))
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
