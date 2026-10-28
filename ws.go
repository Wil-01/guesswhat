package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/net/websocket"
)

func handleWS(ws *websocket.Conn) {
	defer ws.Close()
	for {
		var msg string
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "receive error:", err)
			}
			return
		}
		if err := websocket.Message.Send(ws, strings.ToUpper(msg)); err != nil {
			fmt.Fprintln(os.Stderr, "send error:", err)
			return
		}
	}
}
