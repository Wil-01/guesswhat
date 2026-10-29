package main

const (
	TypeNickname = "nickname"
	TypeMessage  = "message"
	TypeHistory  = "history"
	TypeSystem   = "system"
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
