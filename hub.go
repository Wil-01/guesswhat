package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"

	"golang.org/x/net/websocket"
)

const (
	StateWaiting = "waiting"
	StatePlaying = "playing"

	minPlayers = 2
)

var (
	errNotAdmin         = errors.New("only the admin can start the game")
	errAlreadyStarted   = errors.New("the game has already started")
	errNotEnoughPlayers = fmt.Errorf("at least %d players are needed", minPlayers)
)

type Client struct {
	conn     *websocket.Conn
	id       int
	nickname string
}

type Hub struct {
	mu      sync.Mutex
	clients []*Client
	admin   *Client
	state   string
	history []ChatPacket
	drawing []Stroke
	nextID  int
}

func newHub() *Hub {
	return &Hub{
		state:   StateWaiting,
		history: []ChatPacket{},
		drawing: []Stroke{},
	}
}

func (h *Hub) join(conn *websocket.Conn) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.nextID++
	c := &Client{conn: conn, id: h.nextID, nickname: fmt.Sprintf("guest%d", h.nextID)}
	h.clients = append(h.clients, c)
	if h.admin == nil {
		h.admin = c
	}
	h.send(c, NicknamePacket{Type: TypeNickname, ID: c.id, Nickname: c.nickname})
	h.send(c, HistoryPacket{Type: TypeHistory, Messages: h.history})
	h.send(c, DrawingPacket{Type: TypeDrawing, Strokes: h.drawing})
	h.send(c, GamePacket{Type: TypeGame, State: h.state})
	h.broadcastPlayers()
	return c
}

func (h *Hub) leave(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients = slices.DeleteFunc(h.clients, func(o *Client) bool { return o == c })
	if h.admin == c {
		h.admin = nil
		if len(h.clients) > 0 {
			h.admin = h.clients[0]
		}
	}
	h.broadcastPlayers()
}

func (h *Hub) rename(c *Client, nickname string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	old := c.nickname
	c.nickname = nickname
	h.send(c, NicknamePacket{Type: TypeNickname, ID: c.id, Nickname: nickname})
	h.broadcast(ChatPacket{Type: TypeSystem, Message: fmt.Sprintf("%s is now %s", old, nickname)})
	h.broadcastPlayers()
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

func (h *Hub) start(c *Client) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	switch {
	case c != h.admin:
		return errNotAdmin
	case h.state != StateWaiting:
		return errAlreadyStarted
	case len(h.clients) < minPlayers:
		return errNotEnoughPlayers
	}
	h.state = StatePlaying
	h.broadcast(GamePacket{Type: TypeGame, State: h.state})
	h.broadcast(ChatPacket{Type: TypeSystem, Message: "The game has started!"})
	return nil
}

func (h *Hub) reject(c *Client, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.send(c, ChatPacket{Type: TypeError, Message: err.Error()})
}

// The helpers below must be called with h.mu held.

func (h *Hub) players() []PlayerInfo {
	players := make([]PlayerInfo, len(h.clients))
	for i, c := range h.clients {
		players[i] = PlayerInfo{ID: c.id, Nickname: c.nickname, Admin: c == h.admin}
	}
	return players
}

func (h *Hub) broadcastPlayers() {
	h.broadcast(PlayersPacket{Type: TypePlayers, Players: h.players()})
}

func (h *Hub) broadcast(v any) {
	h.broadcastExcept(nil, v)
}

func (h *Hub) broadcastExcept(except *Client, v any) {
	for _, c := range h.clients {
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
