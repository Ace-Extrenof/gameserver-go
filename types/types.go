package types

type Login struct {
    ClientID int `json:"clientid"`
    Username string `json:"username"`
}

type Position struct {
    ClientID int `json:"clientid"`
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

type WSMessage struct {
    Type string `json:"type"`
    Data PlayerState `json:"data"`
}

type PlayerState struct {
    HP int `json:"health"`
    Position Position `json:"position"`
    SessionID int `json:"sessionid"`
    ClientID int `json:"clientid"`
}
