package main

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/net/websocket"
)

type Client struct {
	conn     *websocket.Conn
	nickname string
}

type Hub struct {
	mu      sync.Mutex
	clients map[*Client]bool
	history []ChatPacket
	drawing []Stroke
	guests  int
}

func newHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		history: []ChatPacket{},
		drawing: []Stroke{},
	}
}

func (h *Hub) join(conn *websocket.Conn) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.guests++
	c := &Client{conn: conn, nickname: fmt.Sprintf("guest%d", h.guests)}
	h.clients[c] = true
	h.send(c, NicknamePacket{Type: TypeNickname, Nickname: c.nickname})
	h.send(c, HistoryPacket{Type: TypeHistory, Messages: h.history})
	h.send(c, DrawingPacket{Type: TypeDrawing, Strokes: h.drawing})
	return c
}

func (h *Hub) leave(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
}

func (h *Hub) rename(c *Client, nickname string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	old := c.nickname
	c.nickname = nickname
	h.send(c, NicknamePacket{Type: TypeNickname, Nickname: nickname})
	h.broadcast(ChatPacket{Type: TypeSystem, Message: fmt.Sprintf("%s is now %s", old, nickname)})
}

func (h *Hub) chat(c *Client, text string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	msg := ChatPacket{Type: TypeMessage, Message: text, Nickname: c.nickname}
	h.history = append(h.history, msg)
	h.broadcast(msg)
}

func (h *Hub) draw(from *Client, s Stroke) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.drawing = append(h.drawing, s)
	h.broadcastExcept(from, DrawPacket{Type: TypeDraw, Stroke: s})
}

func (h *Hub) clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.drawing = []Stroke{}
	h.broadcast(Envelope{Type: TypeClear})
}

// Callers must hold h.mu.
func (h *Hub) broadcast(v any) {
	h.broadcastExcept(nil, v)
}

// Callers must hold h.mu.
func (h *Hub) broadcastExcept(except *Client, v any) {
	for c := range h.clients {
		if c != except {
			h.send(c, v)
		}
	}
}

func (h *Hub) send(c *Client, v any) {
	if err := websocket.JSON.Send(c.conn, v); err != nil {
		fmt.Fprintln(os.Stderr, "send error:", err)
	}
}
