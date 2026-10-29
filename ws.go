package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/websocket"
)

const maxNicknameLen = 20

func (h *Hub) handleWS(conn *websocket.Conn) {
	defer conn.Close()
	client := h.join(conn)
	defer h.leave(client)

	dec := json.NewDecoder(conn)
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "decode error:", err)
			}
			return
		}
		if err := h.handlePacket(client, raw); err != nil {
			fmt.Fprintf(os.Stderr, "client %d: %v\n", client.id, err)
			h.reject(client, err)
		}
	}
}

func (h *Hub) handlePacket(c *Client, raw json.RawMessage) error {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return err
	}
	switch env.Type {
	case TypeNickname:
		var p NicknamePacket
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		name := strings.TrimSpace(p.Nickname)
		if name == "" || utf8.RuneCountInString(name) > maxNicknameLen {
			return fmt.Errorf("invalid nickname %q", p.Nickname)
		}
		h.rename(c, name)
	case TypeMessage:
		var p ChatPacket
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		text := strings.TrimSpace(p.Message)
		if text == "" {
			return nil
		}
		h.chat(c, text)
	case TypeDraw:
		var p DrawPacket
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		if !p.Stroke.valid() {
			return fmt.Errorf("invalid stroke %+v", p.Stroke)
		}
		h.draw(c, p.Stroke)
	case TypeClear:
		h.clear()
	case TypeStart:
		return h.start(c)
	default:
		return fmt.Errorf("unknown packet type %q", env.Type)
	}
	return nil
}
