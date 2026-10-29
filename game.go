package main

import (
	"cmp"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"
)

const (
	maxPoints = 100
	minPoints = 10
)

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
	for _, p := range h.clients {
		p.score = 0
	}
	h.state = StatePlaying
	h.round = 0
	h.broadcast(GamePacket{Type: TypeGame, State: h.state})
	h.broadcast(ChatPacket{Type: TypeSystem, Message: "The game has started!"})
	h.startRound(nil)
	return nil
}

// The helpers below must be called with h.mu held.

func (h *Hub) roundActive() bool {
	return h.state == StatePlaying && h.drawer != nil
}

func (h *Hub) pickDrawer(previous *Client) *Client {
	candidates := slices.DeleteFunc(slices.Clone(h.clients), func(c *Client) bool { return c == previous })
	if len(candidates) == 0 {
		candidates = h.clients
	}
	return candidates[rand.IntN(len(candidates))]
}

func (h *Hub) startRound(previous *Client) {
	drawer := h.pickDrawer(previous)
	h.round++
	h.drawer = drawer
	h.word = randomWord()
	h.roundID++
	h.roundEndsAt = time.Now().Add(h.roundDuration)
	id := h.roundID
	h.timer = time.AfterFunc(h.roundDuration, func() { h.timeUp(id) })

	h.clearDrawing()
	for _, c := range h.clients {
		h.send(c, h.roundPacket(c))
	}
	h.broadcast(ChatPacket{Type: TypeSystem, Message: fmt.Sprintf("%s is drawing.", drawer.nickname)})
	h.broadcastPlayers()
}

func (h *Hub) roundPacket(c *Client) RoundPacket {
	p := RoundPacket{
		Type:      TypeRound,
		Drawer:    h.drawer.id,
		Hint:      hint(h.word),
		Duration:  h.roundDuration.Milliseconds(),
		Remaining: max(time.Until(h.roundEndsAt), 0).Milliseconds(),
		Round:     h.round,
		Rounds:    h.rounds,
	}
	if c == h.drawer {
		p.Word = h.word
	}
	return p
}

func (h *Hub) guessed(c *Client) {
	points := pointsFor(time.Until(h.roundEndsAt), h.roundDuration)
	c.score += points
	h.broadcast(ChatPacket{Type: TypeSystem, Message: fmt.Sprintf("%s guessed the word! (+%d)", c.nickname, points)})
	h.endRound(c, points)
}

func (h *Hub) endRound(winner *Client, points int) {
	h.stopTimer()
	p := RoundEndPacket{Type: TypeRoundEnd, Word: h.word, Points: points}
	if winner != nil {
		p.Winner = winner.id
	}
	h.broadcast(p)

	previous := h.drawer
	h.drawer = nil
	h.word = ""
	h.broadcastPlayers()
	if h.round >= h.rounds {
		h.endGame()
		return
	}
	id := h.roundID
	h.timer = time.AfterFunc(h.pauseDuration, func() { h.nextRound(id, previous) })
}

func (h *Hub) nextRound(id int, previous *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if id != h.roundID {
		return
	}
	h.startRound(previous)
}

func (h *Hub) endGame() {
	ranking := h.players()
	slices.SortStableFunc(ranking, func(a, b PlayerInfo) int { return cmp.Compare(b.Score, a.Score) })
	h.reset()
	h.broadcast(GameEndPacket{Type: TypeGameEnd, Players: ranking})
	h.broadcast(GamePacket{Type: TypeGame, State: h.state})
	h.broadcastPlayers()
}

func (h *Hub) stopGame(reason string) {
	h.stopTimer()
	h.reset()
	h.broadcast(ChatPacket{Type: TypeSystem, Message: reason})
	h.broadcast(GamePacket{Type: TypeGame, State: h.state})
}

func (h *Hub) reset() {
	h.state = StateWaiting
	h.drawer = nil
	h.word = ""
}

// stopTimer also invalidates a callback that already fired and waits for the lock.
func (h *Hub) stopTimer() {
	h.roundID++
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
}

func (h *Hub) timeUp(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if id != h.roundID {
		return
	}
	h.broadcast(ChatPacket{Type: TypeSystem, Message: "Time's up!"})
	h.endRound(nil, 0)
}

func pointsFor(remaining, total time.Duration) int {
	if total <= 0 {
		return minPoints
	}
	return max(int(maxPoints*remaining/total), minPoints)
}
