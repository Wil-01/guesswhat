package main

const (
	TypeNickname = "nickname"
	TypeMessage  = "message"
	TypeHistory  = "history"
	TypeSystem   = "system"
	TypeDraw     = "draw"
	TypeClear    = "clear"
	TypeDrawing  = "drawing"
)

type Envelope struct {
	Type string `json:"type"`
}

type NicknamePacket struct {
	Type     string `json:"type"`
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
