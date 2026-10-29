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
	TypeUndo     = "undo"
	TypePlayers  = "players"
	TypeGame     = "game"
	TypeStart    = "start"
	TypeRound    = "round"
	TypeRoundEnd = "roundEnd"
	TypeGameEnd  = "gameEnd"
	TypeSettings = "settings"
	TypeAvatar   = "avatar"
	TypeTeam     = "team"
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

// Stroke is one drawing operation; Begin marks the first operation of an undoable action.
type Stroke struct {
	Kind   string  `json:"kind"`
	From   Point   `json:"from"`
	To     Point   `json:"to"`
	Points []Point `json:"points,omitempty"`
	Color  string  `json:"color"`
	Size   float64 `json:"size"`
	Begin  bool    `json:"begin,omitempty"`
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
	Avatar   Avatar `json:"avatar"`
	Team     int    `json:"team,omitempty"`
}

type AvatarPacket struct {
	Type   string `json:"type"`
	Avatar Avatar `json:"avatar"`
}

type TeamPacket struct {
	Type string `json:"type"`
	Team int    `json:"team"`
}

// Teams holds the team scores, in team order, when teams are enabled.
type PlayersPacket struct {
	Type    string       `json:"type"`
	Players []PlayerInfo `json:"players"`
	Teams   []int        `json:"teams,omitempty"`
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
	Theme     string `json:"theme"`
	Team      int    `json:"team,omitempty"`
}

type SettingsPacket struct {
	Type string `json:"type"`
	Settings
	Themes []string `json:"themes,omitempty"`
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
	Teams   []int        `json:"teams,omitempty"`
}

type RoomInfo struct {
	Code    string `json:"code"`
	Players int    `json:"players"`
	State   string `json:"state"`
	Theme   string `json:"theme"`
	Teams   bool   `json:"teams"`
}

type GamePacket struct {
	Type  string `json:"type"`
	State string `json:"state"`
}
