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

func (r *Room) start(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case c != r.admin:
		return errNotAdmin
	case r.state != StateWaiting:
		return errAlreadyStarted
	case len(r.clients) < minPlayers:
		return errNotEnoughPlayers
	}
	for _, p := range r.clients {
		p.score = 0
	}
	r.state = StatePlaying
	r.round = 0
	r.broadcast(GamePacket{Type: TypeGame, State: r.state})
	r.broadcast(ChatPacket{Type: TypeSystem, Message: "The game has started!"})
	r.startRound(nil)
	return nil
}

// The helpers below must be called with r.mu held.

func (r *Room) roundActive() bool {
	return r.state == StatePlaying && r.drawer != nil
}

func (r *Room) pickDrawer(previous *Client) *Client {
	candidates := slices.DeleteFunc(slices.Clone(r.clients), func(c *Client) bool { return c == previous })
	if len(candidates) == 0 {
		candidates = r.clients
	}
	return candidates[rand.IntN(len(candidates))]
}

func (r *Room) startRound(previous *Client) {
	drawer := r.pickDrawer(previous)
	r.round++
	r.drawer = drawer
	r.word, r.wordTheme = r.words.Random(r.theme)
	r.roundID++
	r.roundEndsAt = time.Now().Add(r.roundDuration)
	id := r.roundID
	r.timer = time.AfterFunc(r.roundDuration, func() { r.timeUp(id) })

	r.clearDrawing()
	for _, c := range r.clients {
		r.send(c, r.roundPacket(c))
	}
	r.broadcast(ChatPacket{Type: TypeSystem, Message: fmt.Sprintf("%s is drawing.", drawer.nickname)})
	r.broadcastPlayers()
}

func (r *Room) roundPacket(c *Client) RoundPacket {
	p := RoundPacket{
		Type:      TypeRound,
		Drawer:    r.drawer.id,
		Hint:      hint(r.word),
		Duration:  r.roundDuration.Milliseconds(),
		Remaining: max(time.Until(r.roundEndsAt), 0).Milliseconds(),
		Round:     r.round,
		Rounds:    r.rounds,
		Theme:     r.wordTheme,
	}
	if c == r.drawer {
		p.Word = r.word
	}
	return p
}

func (r *Room) guessed(c *Client) {
	points := pointsFor(time.Until(r.roundEndsAt), r.roundDuration)
	c.score += points
	r.broadcast(ChatPacket{Type: TypeSystem, Message: fmt.Sprintf("%s guessed the word! (+%d)", c.nickname, points)})
	r.endRound(c, points)
}

func (r *Room) endRound(winner *Client, points int) {
	r.stopTimer()
	p := RoundEndPacket{Type: TypeRoundEnd, Word: r.word, Points: points}
	if winner != nil {
		p.Winner = winner.id
	}
	r.broadcast(p)

	previous := r.drawer
	r.drawer = nil
	r.word = ""
	r.broadcastPlayers()
	if r.round >= r.rounds {
		r.endGame()
		return
	}
	id := r.roundID
	r.timer = time.AfterFunc(r.pauseDuration, func() { r.nextRound(id, previous) })
}

func (r *Room) nextRound(id int, previous *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id != r.roundID {
		return
	}
	r.startRound(previous)
}

func (r *Room) endGame() {
	ranking := r.players()
	slices.SortStableFunc(ranking, func(a, b PlayerInfo) int { return cmp.Compare(b.Score, a.Score) })
	r.broadcast(GameEndPacket{Type: TypeGameEnd, Players: ranking})
	r.close()
}

func (r *Room) stopGame(reason string) {
	r.stopTimer()
	r.reset()
	r.broadcast(ChatPacket{Type: TypeSystem, Message: reason})
	r.broadcast(GamePacket{Type: TypeGame, State: r.state})
}

func (r *Room) reset() {
	r.state = StateWaiting
	r.drawer = nil
	r.word = ""
}

// stopTimer also invalidates a callback that already fired and waits for the lock.
func (r *Room) stopTimer() {
	r.roundID++
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
}

func (r *Room) timeUp(id int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id != r.roundID {
		return
	}
	r.broadcast(ChatPacket{Type: TypeSystem, Message: "Time's up!"})
	r.endRound(nil, 0)
}

func pointsFor(remaining, total time.Duration) int {
	if total <= 0 {
		return minPoints
	}
	return max(int(maxPoints*remaining/total), minPoints)
}
