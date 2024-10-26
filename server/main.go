package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"

	"github.com/Ace-Extrenof/gameserver/types"
	"github.com/anthdm/hollywood/actor"
	"github.com/gorilla/websocket"
)

type GameServer struct {
    ctx *actor.Context
    sessions map[*actor.PID]struct{}
}

type PlayerSession struct {
    sessionID int
    clientID int
    username string
    inLobby bool
    conn *websocket.Conn
}

func newPlayerSession(sid int, conn *websocket.Conn) actor.Producer {
    return func() actor.Receiver {
        return &PlayerSession {
            sessionID: sid,
            conn: conn,
        }
    }
}

func (ps *PlayerSession) Receive(c *actor.Context) {
    switch c.Message().(type) {
    case actor.Started:
        ps.readLoop()
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

func (ps *PlayerSession) handleMessage(msg types.WSMessage) {
    switch msg.Type {
    case "login":
        var loginMsg types.Login
        if err := json.Unmarshal(msg.Data, &loginMsg); err != nil {
            panic(err)
        }
        ps.clientID = loginMsg.ClientID
        ps.username = loginMsg.Username

    case "playerState":
        var ps types.PlayerState
        if err := json.Unmarshal(msg.Data, &ps); err != nil {
            panic(err)
        }
        fmt.Println(ps)
    }
}

func newGameServer() actor.Receiver {
    return &GameServer{
        sessions: make(map[*actor.PID]struct{}),
    }
}

func (s *GameServer) Receive(c *actor.Context) {
    switch msg := c.Message().(type) {
    case actor.Started:
        s.startHTTP()
        s.ctx = c
        _ = msg
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

    fmt.Print("Client trying to connect...")
    sid := rand.IntN(math.MaxInt)
    pid := s.ctx.SpawnChild(newPlayerSession(sid, conn), fmt.Sprintf("session_%d", sid))
    s.sessions[pid] = struct{}{}
}

func main() {
    e, err := actor.NewEngine(actor.NewEngineConfig())
    e.Spawn(newGameServer, "game_server")

    if err != nil {
        fmt.Println(err)
    }

    select { }
}
