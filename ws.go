package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/net/websocket"
)

type Message struct {
	Message string `json:"message"`
}

func handleWS(ws *websocket.Conn) {
	hub.add(ws)
	defer func() {
		hub.remove(ws)
		ws.Close()
	}()

	dec := json.NewDecoder(ws)
	for {
		var msg Message
		if err := dec.Decode(&msg); err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "decode error:", err)
			}
			return
		}
		msg.Message = strings.TrimSpace(msg.Message)
		if msg.Message == "" {
			continue
		}
		hub.broadcast(msg)
	}
}
