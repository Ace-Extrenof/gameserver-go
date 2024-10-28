package main

import (
    "fmt"
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins for testing
    },
}

type Client struct {
    ID   string
    Conn *websocket.Conn
    mu   sync.Mutex
}

type Message struct {
    Type string      `json:"type"`
    Data interface{} `json:"data"`
}

type PlayerState struct {
    ClientID string  `json:"clientId"`
    X        float64 `json:"x"`
    Y        float64 `json:"y"`
}

var (
    clients    = make(map[string]*Client)
    clientsMux sync.RWMutex
)

func (c *Client) sendMessage(message Message) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if err := c.Conn.WriteJSON(message); err != nil {
        log.Printf("Error sending message to client %s: %v", c.ID, err)
    }
}

func broadcastMessage(message Message, excludeID string) {
    clientsMux.RLock()
    defer clientsMux.RUnlock()

    for id, client := range clients {
        if id != excludeID {
            client.sendMessage(message)
        }
    }
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
    // Upgrade the HTTP connection to a WebSocket connection
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Error upgrading connection: %v", err)
        return
    }

    // Generate a new client ID
    clientID := fmt.Sprintf("%d", len(clients)+1)
    
    // Create new client
    client := &Client{
        ID:   clientID,
        Conn: conn,
    }

    // Add client to clients map
    clientsMux.Lock()
    clients[clientID] = client
    clientsMux.Unlock()

    log.Printf("New client connected. ID: %s", clientID)

    // Send initial message to client with their ID
    initMessage := Message{
        Type: "init",
        Data: map[string]interface{}{
            "clientId": clientID,
        },
    }
    client.sendMessage(initMessage)

    // Handle client disconnect
    defer func() {
        clientsMux.Lock()
        delete(clients, clientID)
        clientsMux.Unlock()
        conn.Close()
        log.Printf("Client disconnected. ID: %s", clientID)
    }()

    // Message handling loop
    for {
        var message Message
        err := conn.ReadJSON(&message)
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("Error reading message: %v", err)
            }
            break
        }

        // Handle different message types
        switch message.Type {
        case "playerState":
            // Parse the player state data
            if stateData, ok := message.Data.(map[string]interface{}); ok {
                state := PlayerState{
                    ClientID: clientID,
                    X:        stateData["x"].(float64),
                    Y:        stateData["y"].(float64),
                }

                // Broadcast the player state to all other clients
                broadcastMessage(Message{
                    Type: "playerState",
                    Data: state,
                }, clientID)
            }
        }
    }
}

func main() {
    // Serve static files
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/", fs)

    // WebSocket endpoint
    http.HandleFunc("/ws", handleConnection)

    // Start server
    port := ":8000"
    log.Printf("Server starting on %s", port)
    if err := http.ListenAndServe(port, nil); err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
