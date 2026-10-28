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
	defer ws.Close()
	dec := json.NewDecoder(ws)
	enc := json.NewEncoder(ws)
	for {
		var msg Message
		if err := dec.Decode(&msg); err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "decode error:", err)
			}
			return
		}
		msg.Message = strings.ToUpper(msg.Message)
		if err := enc.Encode(msg); err != nil {
			fmt.Fprintln(os.Stderr, "encode error:", err)
			return
		}
	}
}
