package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"sync"

	"github.com/Ace-Extrenof/gameserver/types"
	"github.com/anthdm/hollywood/actor"
	"github.com/gorilla/websocket"
)

type GameServer struct {
    ctx *actor.Context
    sessions map[int]*actor.PID
    clients map[*websocket.Conn]bool
    mu sync.Mutex
}

type PlayerSession struct {
    sessionID int
    clientID int
    username string
    inLobby bool
    conn *websocket.Conn
    ctx *actor.Context
    serverPID *actor.PID
}

func newPlayerSession(serverPID *actor.PID, sid int, conn *websocket.Conn) actor.Producer {
    return func() actor.Receiver {
        return &PlayerSession {
            sessionID: sid,
            conn: conn,
            serverPID: serverPID,
        }
    }
}

func (ps *PlayerSession) Receive(c *actor.Context) {
    switch msg := c.Message().(type) {
    case actor.Started:
        ps.ctx = c
        ps.readLoop()
    case *types.PlayerState:
        ps.sendPlayerState(msg)
    }
}

func (ps *PlayerSession) sendPlayerState(state *types.PlayerState) {
    b, err := json.Marshal(state)
    if err != nil {
        panic(err)
    }
    msg := types.WSMessage{
        Type: "state",
        Data: b,
    }
    if err := ps.conn.WriteJSON(msg); err != nil {
        panic(err)
    }
}

func (ps *PlayerSession) readLoop() {
    var msg types.WSMessage
    for {
        if err := ps.conn.ReadJSON(&msg); err != nil {
            fmt.Println("read err:", err)
            return
        }
        go ps.handleMessage(msg)
    }
}

func (s *PlayerSession) handleMessage(msg types.WSMessage) {
    switch msg.Type {
    case "login":
        var loginMsg types.Login
        if err := json.Unmarshal(msg.Data, &loginMsg); err != nil {
            panic(err)
        }
        s.clientID = loginMsg.ClientID
        s.username = loginMsg.Username
        s.inLobby = true
        fmt.Printf("client logged in: %s (ID: %d)\n", s.username, s.clientID)

    case "playerState":
        var ps types.PlayerState
        if err := json.Unmarshal(msg.Data, &ps); err != nil {
            panic(err)
        }
        ps.SessionID = s.sessionID
        fmt.Printf("received playerState: %+v\n", ps)

        if s.ctx != nil {
            s.ctx.Send(s.serverPID, &ps)
        }
    }
}

func newGameServer() actor.Receiver {
    return &GameServer{
        sessions: make(map[int]*actor.PID),
        clients: make(map[*websocket.Conn]bool),
    }
}

func (s *GameServer) Receive(c *actor.Context) {
    switch msg := c.Message().(type) {
    case *types.PlayerState:
        s.bcast(c.Sender(), msg)
    case actor.Started:
        s.startHTTP()
        s.ctx = c
        _ = msg
    default:
        fmt.Println("recv", msg)
    }
}

func (s *GameServer) bcast(exclude *actor.PID, state *types.PlayerState) {
    for _, pid := range s.sessions {
        if pid != exclude {
            s.ctx.Send(pid, state)
        }
    }
}

func (s *GameServer) startHTTP() {
    fmt.Println("Starting server -> 8000")
    go func() {
        http.HandleFunc("/ws", s.handleWS)
        http.ListenAndServe(":8000", nil)
    }()
}

func (s *GameServer) handleWS(w http.ResponseWriter, r *http.Request) {
    var upgrader = websocket.Upgrader{
        ReadBufferSize: 1024,
        WriteBufferSize: 1024,
        CheckOrigin: func(r *http.Request) bool { return true }, // allows every conn
    }

    conn, err := upgrader.Upgrade(w, r, nil)

    if err != nil {
        fmt.Println(err)
    }
    defer conn.Close()

    s.mu.Lock()
    s.clients[conn] = true
    s.mu.Unlock()

    fmt.Println("new client connected")

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            fmt.Println("err reading msg:", err)
            break
        }
        fmt.Printf("received msg: %s\n", msg)
    }

    sid := rand.IntN(math.MaxInt)
    pid := s.ctx.SpawnChild(newPlayerSession(s.ctx.PID(), sid, conn), fmt.Sprintf("session_%d", sid))
    s.sessions[sid] = pid

    s.mu.Lock()
    delete(s.clients, conn)
    s.mu.Unlock()
    fmt.Println("client disconnected")
}

func main() {
    e, err := actor.NewEngine(actor.NewEngineConfig())
    e.Spawn(newGameServer, "game_server")

    if err != nil {
        fmt.Println(err)
    }

    select { }
}
