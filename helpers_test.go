package main

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

type anyPacket struct {
	Type     string       `json:"type"`
	ID       int          `json:"id"`
	Nickname string       `json:"nickname"`
	Message  string       `json:"message"`
	State    string       `json:"state"`
	Messages []ChatPacket `json:"messages"`
	Stroke   Stroke       `json:"stroke"`
	Strokes  []Stroke     `json:"strokes"`
	Players  []PlayerInfo `json:"players"`
	Drawer   int          `json:"drawer"`
	Word     string       `json:"word"`
	Hint     string       `json:"hint"`
	Winner   int          `json:"winner"`
	Points   int          `json:"points"`
	Round    int          `json:"round"`
	Rounds   int          `json:"rounds"`
}

type testClient struct {
	*websocket.Conn
	id       int
	nickname string
	history  []ChatPacket
	drawing  []Stroke
	round    anyPacket
}

func newTestServer(t *testing.T) (*Room, *httptest.Server) {
	t.Helper()
	room := newRoom()
	room.rounds = 1
	room.pauseDuration = 10 * time.Millisecond
	srv := httptest.NewServer(websocket.Handler(room.handleWS))
	t.Cleanup(srv.Close)
	return room, srv
}

func next(t *testing.T, ws *websocket.Conn) anyPacket {
	t.Helper()
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	var p anyPacket
	if err := websocket.JSON.Receive(ws, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

// expect skips players updates unless they are what is expected.
func expect(t *testing.T, ws *websocket.Conn, typ string) anyPacket {
	t.Helper()
	for {
		p := next(t, ws)
		if p.Type == typ {
			return p
		}
		if p.Type != TypePlayers {
			t.Fatalf("got packet %+v, want type %q", p, typ)
		}
	}
}

func expectClosed(t *testing.T, ws *websocket.Conn) {
	t.Helper()
	for {
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		var p anyPacket
		err := websocket.JSON.Receive(ws, &p)
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatalf("expected connection close, got %v", err)
		}
		if p.Type != TypePlayers {
			t.Fatalf("expected connection close, got packet %+v", p)
		}
	}
}

func dial(t *testing.T, srv *httptest.Server, path string) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + path
	ws, err := websocket.Dial(url, "", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ws.Close() })
	return ws
}

// connect dials and consumes the welcome packets, so the client is registered on return.
func connect(t *testing.T, srv *httptest.Server) *testClient {
	t.Helper()
	return join(t, dial(t, srv, ""))
}

func join(t *testing.T, ws *websocket.Conn) *testClient {
	t.Helper()
	welcome := expect(t, ws, TypeNickname)
	c := &testClient{Conn: ws, id: welcome.ID, nickname: welcome.Nickname}
	c.history = expect(t, ws, TypeHistory).Messages
	c.drawing = expect(t, ws, TypeDrawing).Strokes
	if expect(t, ws, TypeGame).State == StatePlaying {
		c.round = expect(t, ws, TypeRound)
	}
	expect(t, ws, TypePlayers)
	return c
}

func send(t *testing.T, ws *websocket.Conn, v any) {
	t.Helper()
	if err := websocket.JSON.Send(ws, v); err != nil {
		t.Fatal(err)
	}
}

func say(t *testing.T, ws *websocket.Conn, text string) {
	t.Helper()
	send(t, ws, ChatPacket{Type: TypeMessage, Message: text})
}

func setRoom(room *Room, f func(r *Room)) {
	room.mu.Lock()
	defer room.mu.Unlock()
	f(room)
}

type round struct {
	drawer  *testClient
	word    string
	packets map[*testClient]anyPacket
}

// guesser returns any client that is not drawing.
func (r round) guesser() *testClient {
	for c := range r.packets {
		if c != r.drawer {
			return c
		}
	}
	return nil
}

// startGame starts the game as admin and returns the first round.
func startGame(t *testing.T, admin *testClient, all ...*testClient) round {
	t.Helper()
	send(t, admin.Conn, Envelope{Type: TypeStart})
	for _, c := range all {
		expect(t, c.Conn, TypeGame)
		expect(t, c.Conn, TypeSystem)
	}
	return expectRound(t, all...)
}

func expectRound(t *testing.T, all ...*testClient) round {
	t.Helper()
	r := round{packets: make(map[*testClient]anyPacket)}
	for _, c := range all {
		expect(t, c.Conn, TypeClear)
		p := expect(t, c.Conn, TypeRound)
		expect(t, c.Conn, TypeSystem)
		r.packets[c] = p
		if p.Drawer == c.id {
			r.drawer, r.word = c, p.Word
		}
	}
	if r.drawer == nil {
		t.Fatalf("no drawer among clients: %+v", r.packets)
	}
	return r
}
