package main

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/net/websocket"
)

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

var hub = &Hub{clients: make(map[*websocket.Conn]bool)}

func (h *Hub) add(ws *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[ws] = true
}

func (h *Hub) remove(ws *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, ws)
}

func (h *Hub) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

func (h *Hub) broadcast(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ws := range h.clients {
		if err := websocket.JSON.Send(ws, msg); err != nil {
			fmt.Fprintln(os.Stderr, "broadcast error:", err)
		}
	}
}
