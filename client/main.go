package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"github.com/Ace-Extrenof/gameserver/types"
	"github.com/gorilla/websocket"
)

type GameClient struct {
    conn *websocket.Conn
    clientID int
    username string
}

const wsServerEndpoint = "ws://localhost:8000/ws"

func (c *GameClient) login() error {
    b, err := json.Marshal(types.Login{
        ClientID: c.clientID,
        Username: c.username,
    })

    if err != nil {
        return err
    }

    msg := types.WSMessage {
        Type: "login",
        Data: b,
    }

    return c.conn.WriteJSON(msg)
}

func newGameClient(conn *websocket.Conn, username string) *GameClient {
    return &GameClient{
        conn: conn,
        clientID: rand.IntN(math.MaxInt),
        username: username,
    }
}

func main() {
    dialer := websocket.Dialer {
        ReadBufferSize: 1024,
        WriteBufferSize: 1024,
    }

    conn, _, err := dialer.Dial(wsServerEndpoint, nil)
    if err != nil {
        log.Fatalf("Failed to upgrade conn : %v", err)
    }
    defer conn.Close()

    c := newGameClient(conn, "James")
    if err := c.login(); err != nil {
        log.Fatal(err)
    }

    go func() {
        var msg types.WSMessage

        for {
            if err := c.conn.ReadJSON(&msg); err != nil {
                fmt.Println("WS read err:", err)
                continue
            }
            switch msg.Type {
            case "state":
                var state types.PlayerState
                if err := json.Unmarshal(msg.Data, &state); err != nil {
                    fmt.Println("WS read err")
                    continue
                }
                fmt.Println("need to update state of player", state)

            default:
                fmt.Println("receiving unknown message")
            }
        }

    }()

    select {}
}
