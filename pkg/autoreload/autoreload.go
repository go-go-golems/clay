package autoreload

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	clients    map[*websocket.Conn]bool
	clientsMux sync.Mutex
	upgrader   websocket.Upgrader
}

// NewWebSocketServer creates a new WebSocket server instance.
func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// WebSocketHandler returns an HTTP handler that upgrades HTTP connections to WebSocket connections.
func (ws *WebSocketServer) WebSocketHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := ws.upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
			return
		}
		ws.clientsMux.Lock()
		ws.clients[conn] = true
		ws.clientsMux.Unlock()

		defer func() {
			ws.clientsMux.Lock()
			delete(ws.clients, conn)
			ws.clientsMux.Unlock()
			conn.Close()
		}()

		// Keep the connection alive until an error occurs
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}

// Broadcast sends a message to all connected WebSocket clients.
func (ws *WebSocketServer) Broadcast(message string) {
	ws.clientsMux.Lock()
	defer ws.clientsMux.Unlock()
	for client := range ws.clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			client.Close()
			delete(ws.clients, client)
		}
	}
}

// GetJavaScript returns a JavaScript snippet that sets up a WebSocket connection to the server.
// The 'mountPoint' parameter should be the WebSocket endpoint (e.g., "/ws").
func (ws *WebSocketServer) GetJavaScript(mountPoint string) string {
	return `
(function() {
    const socket = new WebSocket("ws://" + window.location.host + "` + mountPoint + `");
    
    socket.onopen = function() {
        console.log("WebSocket connection established");
    };

    socket.onmessage = function(event) {
        // User-defined behavior here:
        console.log("Message from server:", event.data);
        if (event.data === "reload") {
            location.reload();
        }
    };

    socket.onclose = function() {
        console.log("WebSocket connection closed");
    };

    socket.onerror = function(error) {
        console.error("WebSocket error: ", error);
    };
})();
`
}
