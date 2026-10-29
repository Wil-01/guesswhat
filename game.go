package main

import "fmt"

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
	h.startRound(h.admin)
	return nil
}

// The helpers below must be called with h.mu held.

func (h *Hub) startRound(drawer *Client) {
	h.drawer = drawer
	h.word = randomWord()
	h.clearDrawing()
	for _, c := range h.clients {
		h.send(c, h.roundPacket(c))
	}
	h.broadcast(ChatPacket{Type: TypeSystem, Message: fmt.Sprintf("%s is drawing.", drawer.nickname)})
	h.broadcastPlayers()
}

func (h *Hub) roundPacket(c *Client) RoundPacket {
	p := RoundPacket{Type: TypeRound, Drawer: h.drawer.id, Hint: hint(h.word)}
	if c == h.drawer {
		p.Word = h.word
	}
	return p
}

func (h *Hub) stopGame(reason string) {
	h.state = StateWaiting
	h.drawer = nil
	h.word = ""
	h.broadcast(ChatPacket{Type: TypeSystem, Message: reason})
	h.broadcast(GamePacket{Type: TypeGame, State: h.state})
}
