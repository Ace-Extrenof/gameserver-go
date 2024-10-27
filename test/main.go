package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/Ace-Extrenof/gameserver/types"
	"github.com/gorilla/websocket"
)

const wsServerEndpoint = "ws://localhost:8000/ws"

type GameClient struct {
    conn *websocket.Conn
    clientID int
    username string
}

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

func runGameClient(num int, username string) {
    conn, _, err := websocket.DefaultDialer.Dial(wsServerEndpoint, nil)

    c := newGameClient(conn, username)
    if err != nil {
        fmt.Printf("login failed %s: %v\n", username, err)
        return
    }
    defer conn.Close()

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
                    return
                }
                fmt.Printf("%s received state update: %+v\n", username, state)

            default:
                fmt.Printf("%s receiving unknown message type: %s\n", username, msg.Type)
            }
        }
    }()

    i := num
    for {
        x := rand.IntN(100)
        y := rand.IntN(100)

        state := types.PlayerState{
            HP: 100,
            Position: types.Position{
                X: x,
                Y: y,
            },
            SessionID: rand.IntN(math.MaxInt),
        }

        b, err := json.Marshal(state)

        if err != nil {
            fmt.Printf("data send err for %s: %v\n", username, err)
            return
        }

        msg := types.WSMessage{
            Type: "playerState",
            Data: b,
        }

        fmt.Printf("%s sending playerState: %+v\n", username, state)

        if err := conn.WriteJSON(msg); err != nil {
            fmt.Printf("write err for %s: %v\n", username, err)
            return
        }

        time.Sleep(time.Second * 1)

        i += 1
        if i == num {
            break
        } else {
            continue
        }
    }
}

func main() {
    runGameClient(5, "boom")

    select {}
}
