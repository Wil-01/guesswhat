package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

const (
	StateWaiting = "waiting"
	StatePlaying = "playing"

	minPlayers           = 2
	maxPlayers           = 10
	defaultEmptyTTL      = time.Minute
	defaultRoundDuration = 80 * time.Second
	defaultPauseDuration = 5 * time.Second
	defaultRounds        = 3
)

var (
	errNotAdmin         = errors.New("only the admin can do that")
	errAlreadyStarted   = errors.New("the game has already started")
	errNotEnoughPlayers = fmt.Errorf("at least %d players are needed", minPlayers)
	errNotDrawer        = errors.New("only the drawer can draw")
	errRevealWord       = errors.New("you can't reveal the word")
	errRoomClosed       = errors.New("this room is closed")
	errRoomFull         = fmt.Errorf("this room is full (%d players max)", maxPlayers)
)

type Client struct {
	conn     *websocket.Conn
	id       int
	nickname string
	score    int
}

type Room struct {
	mu        sync.Mutex
	code      string
	private   bool
	lobby     *Lobby
	closed    bool
	clients   []*Client
	admin     *Client
	drawer    *Client
	word      string
	theme     string
	wordTheme string
	state     string
	history   []ChatPacket
	drawing   []Stroke
	nextID    int

	rounds        int
	round         int
	roundDuration time.Duration
	pauseDuration time.Duration
	roundID       int
	roundEndsAt   time.Time
	timer         *time.Timer

	emptyTTL time.Duration
	cleanup  *time.Timer
}

func newRoom() *Room {
	return &Room{
		state:         StateWaiting,
		history:       []ChatPacket{},
		drawing:       []Stroke{},
		rounds:        defaultRounds,
		roundDuration: defaultRoundDuration,
		pauseDuration: defaultPauseDuration,
		emptyTTL:      defaultEmptyTTL,
	}
}

func (r *Room) info() RoomInfo {
	r.mu.Lock()
	defer r.mu.Unlock()
	return RoomInfo{Code: r.code, Players: len(r.clients), State: r.state, Theme: r.theme}
}

func (r *Room) join(conn *websocket.Conn) (*Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case r.closed:
		return nil, errRoomClosed
	case len(r.clients) >= maxPlayers:
		return nil, errRoomFull
	}
	if r.cleanup != nil {
		r.cleanup.Stop()
		r.cleanup = nil
	}
	r.nextID++
	c := &Client{conn: conn, id: r.nextID, nickname: fmt.Sprintf("guest%d", r.nextID)}
	r.clients = append(r.clients, c)
	if r.admin == nil {
		r.admin = c
	}
	r.send(c, NicknamePacket{Type: TypeNickname, ID: c.id, Nickname: c.nickname})
	r.send(c, HistoryPacket{Type: TypeHistory, Messages: r.history})
	r.send(c, DrawingPacket{Type: TypeDrawing, Strokes: r.drawing})
	r.send(c, r.themePacket())
	r.send(c, GamePacket{Type: TypeGame, State: r.state})
	if r.roundActive() {
		r.send(c, r.roundPacket(c))
	}
	r.broadcastPlayers()
	return c, nil
}

func (r *Room) leave(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.clients = slices.DeleteFunc(r.clients, func(o *Client) bool { return o == c })
	if r.admin == c {
		r.admin = nil
		if len(r.clients) > 0 {
			r.admin = r.clients[0]
		}
	}
	if r.state == StatePlaying {
		switch {
		case len(r.clients) < minPlayers:
			r.stopGame("Not enough players left.")
		case c == r.drawer:
			r.broadcast(ChatPacket{Type: TypeSystem, Message: "The drawer has left."})
			r.endRound(nil, 0)
		}
	}
	if len(r.clients) == 0 {
		r.scheduleCleanup()
		return
	}
	r.broadcastPlayers()
}

func (r *Room) rename(c *Client, nickname string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	old := c.nickname
	c.nickname = nickname
	r.send(c, NicknamePacket{Type: TypeNickname, ID: c.id, Nickname: nickname})
	r.broadcast(ChatPacket{Type: TypeSystem, Message: fmt.Sprintf("%s is now %s", old, nickname)})
	r.broadcastPlayers()
}

func (r *Room) chat(c *Client, text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.roundActive() {
		if c == r.drawer && strings.Contains(normalize(text), normalize(r.word)) {
			return errRevealWord
		}
		if c != r.drawer && normalize(text) == normalize(r.word) {
			r.guessed(c)
			return nil
		}
	}
	msg := ChatPacket{Type: TypeMessage, Message: text, Nickname: c.nickname}
	r.history = append(r.history, msg)
	r.broadcast(msg)
	return nil
}

func (r *Room) draw(from *Client, s Stroke) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.canDraw(from) {
		return errNotDrawer
	}
	r.drawing = append(r.drawing, s)
	r.broadcastExcept(from, DrawPacket{Type: TypeDraw, Stroke: s})
	return nil
}

func (r *Room) undo(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.canDraw(c) {
		return errNotDrawer
	}
	if len(r.drawing) == 0 {
		return nil
	}
	start := 0
	for i := len(r.drawing) - 1; i >= 0; i-- {
		if r.drawing[i].Begin {
			start = i
			break
		}
	}
	r.drawing = r.drawing[:start]
	r.broadcast(DrawingPacket{Type: TypeDrawing, Strokes: r.drawing})
	return nil
}

func (r *Room) clear(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.canDraw(c) {
		return errNotDrawer
	}
	r.clearDrawing()
	return nil
}

func (r *Room) setTheme(c *Client, theme string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case c != r.admin:
		return errNotAdmin
	case r.state != StateWaiting:
		return errAlreadyStarted
	case !validTheme(theme):
		return fmt.Errorf("%w %q", errUnknownTheme, theme)
	}
	r.theme = theme
	r.broadcast(r.themePacket())
	return nil
}

func (r *Room) reject(c *Client, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.send(c, ChatPacket{Type: TypeError, Message: err.Error()})
}

// The helpers below must be called with r.mu held.

func (r *Room) themePacket() ThemePacket {
	return ThemePacket{Type: TypeTheme, Theme: r.theme, Themes: themeNames()}
}

func (r *Room) players() []PlayerInfo {
	players := make([]PlayerInfo, len(r.clients))
	for i, c := range r.clients {
		players[i] = PlayerInfo{
			ID:       c.id,
			Nickname: c.nickname,
			Admin:    c == r.admin,
			Drawer:   r.state == StatePlaying && c == r.drawer,
			Score:    c.score,
		}
	}
	return players
}

// scheduleCleanup closes the room if it stays empty for emptyTTL.
func (r *Room) scheduleCleanup() {
	if r.lobby == nil {
		return
	}
	if r.cleanup != nil {
		r.cleanup.Stop()
	}
	r.cleanup = time.AfterFunc(r.emptyTTL, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if len(r.clients) == 0 {
			r.close()
		}
	})
}

func (r *Room) close() {
	if r.closed {
		return
	}
	r.closed = true
	if r.lobby != nil {
		r.lobby.remove(r.code)
	}
	r.stopTimer()
	if r.cleanup != nil {
		r.cleanup.Stop()
	}
	for _, c := range r.clients {
		c.conn.Close()
	}
	r.clients = nil
}

func (r *Room) canDraw(c *Client) bool {
	return r.state != StatePlaying || c == r.drawer
}

func (r *Room) clearDrawing() {
	r.drawing = []Stroke{}
	r.broadcast(Envelope{Type: TypeClear})
}

func (r *Room) broadcastPlayers() {
	r.broadcast(PlayersPacket{Type: TypePlayers, Players: r.players()})
}

func (r *Room) broadcast(v any) {
	r.broadcastExcept(nil, v)
}

func (r *Room) broadcastExcept(except *Client, v any) {
	for _, c := range r.clients {
		if c != except {
			r.send(c, v)
		}
	}
}

func (r *Room) send(c *Client, v any) {
	if err := websocket.JSON.Send(c.conn, v); err != nil {
		fmt.Fprintln(os.Stderr, "send error:", err)
	}
}
