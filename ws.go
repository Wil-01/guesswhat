package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/websocket"
)

const maxNicknameLen = 20

func (r *Room) handleWS(conn *websocket.Conn) {
	defer conn.Close()
	client, err := r.join(conn)
	if err != nil {
		sendError(conn, err)
		return
	}
	defer r.leave(client)

	dec := json.NewDecoder(conn)
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if err != io.EOF && !errors.Is(err, net.ErrClosed) {
				fmt.Fprintln(os.Stderr, "decode error:", err)
			}
			return
		}
		if err := r.handlePacket(client, raw); err != nil {
			fmt.Fprintf(os.Stderr, "client %d: %v\n", client.id, err)
			r.reject(client, err)
		}
	}
}

func (r *Room) handlePacket(c *Client, raw json.RawMessage) error {
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
		r.rename(c, name)
	case TypeMessage:
		var p ChatPacket
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		text := strings.TrimSpace(p.Message)
		if text == "" {
			return nil
		}
		return r.chat(c, text)
	case TypeDraw:
		var p DrawPacket
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		if !p.Stroke.valid() {
			return fmt.Errorf("invalid stroke %+v", p.Stroke)
		}
		return r.draw(c, p.Stroke)
	case TypeClear:
		return r.clear(c)
	case TypeStart:
		return r.start(c)
	case TypeTheme:
		var p ThemePacket
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		return r.setTheme(c, p.Theme)
	default:
		return fmt.Errorf("unknown packet type %q", env.Type)
	}
	return nil
}

func sendError(conn *websocket.Conn, err error) {
	if err := websocket.JSON.Send(conn, ChatPacket{Type: TypeError, Message: err.Error()}); err != nil {
		fmt.Fprintln(os.Stderr, "send error:", err)
	}
}
