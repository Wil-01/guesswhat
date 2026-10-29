package main

const (
	TypeNickname = "nickname"
	TypeMessage  = "message"
	TypeHistory  = "history"
	TypeSystem   = "system"
	TypeError    = "error"
	TypeDraw     = "draw"
	TypeClear    = "clear"
	TypeDrawing  = "drawing"
	TypePlayers  = "players"
	TypeGame     = "game"
	TypeStart    = "start"
)

type Envelope struct {
	Type string `json:"type"`
}

type NicknamePacket struct {
	Type     string `json:"type"`
	ID       int    `json:"id,omitempty"`
	Nickname string `json:"nickname"`
}

type ChatPacket struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Nickname string `json:"nickname,omitempty"`
}

type HistoryPacket struct {
	Type     string       `json:"type"`
	Messages []ChatPacket `json:"messages"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Stroke struct {
	From  Point   `json:"from"`
	To    Point   `json:"to"`
	Color string  `json:"color"`
	Size  float64 `json:"size"`
}

type DrawPacket struct {
	Type   string `json:"type"`
	Stroke Stroke `json:"stroke"`
}

type DrawingPacket struct {
	Type    string   `json:"type"`
	Strokes []Stroke `json:"strokes"`
}

type PlayerInfo struct {
	ID       int    `json:"id"`
	Nickname string `json:"nickname"`
	Admin    bool   `json:"admin"`
}

type PlayersPacket struct {
	Type    string       `json:"type"`
	Players []PlayerInfo `json:"players"`
}

type GamePacket struct {
	Type  string `json:"type"`
	State string `json:"state"`
}
