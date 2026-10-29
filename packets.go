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
	TypeRound    = "round"
	TypeRoundEnd = "roundEnd"
	TypeGameEnd  = "gameEnd"
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
	Drawer   bool   `json:"drawer"`
	Score    int    `json:"score"`
}

type PlayersPacket struct {
	Type    string       `json:"type"`
	Players []PlayerInfo `json:"players"`
}

type RoundPacket struct {
	Type      string `json:"type"`
	Drawer    int    `json:"drawer"`
	Word      string `json:"word,omitempty"`
	Hint      string `json:"hint"`
	Duration  int64  `json:"duration"`
	Remaining int64  `json:"remaining"`
	Round     int    `json:"round"`
	Rounds    int    `json:"rounds"`
}

type RoundEndPacket struct {
	Type   string `json:"type"`
	Word   string `json:"word"`
	Winner int    `json:"winner"`
	Points int    `json:"points"`
}

type GameEndPacket struct {
	Type    string       `json:"type"`
	Players []PlayerInfo `json:"players"`
}

type RoomInfo struct {
	Code    string `json:"code"`
	Players int    `json:"players"`
	State   string `json:"state"`
}

type GamePacket struct {
	Type  string `json:"type"`
	State string `json:"state"`
}
