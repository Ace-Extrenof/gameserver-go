package types

import "encoding/json"

type Login struct {
    ClientID int `json:"clientid"`
    Username string `json:"username"`
}

type Position struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

type WSMessage struct {
    Type string `json:"type"`
    Data json.RawMessage `json:"data"`
}

type PlayerState struct {
    HP int `json:"health"`
    Position Position `json:"position"`
    SessionID int `json:"sessionid"`
    ClientID int `json:"clientid"`
}
