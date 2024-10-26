package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/Ace-Extrenof/gameserver/types"
	"github.com/gorilla/websocket"
)

type GameClient struct {
    conn *websocket.Conn
    clientID int
    username string
    position types.Position
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

    c := newGameClient(conn, "James")
    if err := c.login(); err != nil {
        log.Fatal(err)
    }

    for {
        x := rand.IntN(math.MaxInt)
        y := rand.IntN(math.MaxInt)
        state := types.PlayerState{
            HP: 100,
            Position: types.Position{X: x, Y: y},
        }

        b, err := json.Marshal(state)

        if err != nil {
            log.Fatalf("data send err: %v", err)
        }

        msg := types.WSMessage {
            Type: "playerState",
            Data: b,
        }

        if err := conn.WriteJSON(msg); err != nil {
            log.Fatal("write err")
        }
        time.Sleep(time.Millisecond * 120)
    }
}
