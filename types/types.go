package types

type Login struct {
    ClientID int `json:"clientid"`
    Username string `json:"username"`
}

type Position struct {
    X int `json:"x"`
    Y int `json:"y"`
}

type WSMessage struct {
    Type string `json:"type"`
    Data []byte `json:"data"`
}

type PlayerState struct {
    HP int `json:"health"`
    Position Position `json:"position"`
}
