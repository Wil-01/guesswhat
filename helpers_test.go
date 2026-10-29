package main

import (
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
}

type testClient struct {
	*websocket.Conn
	id       int
	nickname string
	history  []ChatPacket
	drawing  []Stroke
	round    anyPacket
}

func newTestServer(t *testing.T) (*Hub, *httptest.Server) {
	t.Helper()
	hub := newHub()
	srv := httptest.NewServer(websocket.Handler(hub.handleWS))
	t.Cleanup(srv.Close)
	return hub, srv
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

// connect dials and consumes the welcome packets, so the client is registered on return.
func connect(t *testing.T, srv *httptest.Server) *testClient {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	ws, err := websocket.Dial(url, "", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ws.Close() })
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

// startGame starts the game as admin and returns each client's round packet.
func startGame(t *testing.T, admin *testClient, all ...*testClient) map[*testClient]anyPacket {
	t.Helper()
	send(t, admin.Conn, Envelope{Type: TypeStart})
	rounds := make(map[*testClient]anyPacket)
	for _, c := range all {
		expect(t, c.Conn, TypeGame)
		expect(t, c.Conn, TypeSystem)
		rounds[c] = expectRound(t, c)
	}
	return rounds
}

func expectRound(t *testing.T, c *testClient) anyPacket {
	t.Helper()
	expect(t, c.Conn, TypeClear)
	p := expect(t, c.Conn, TypeRound)
	expect(t, c.Conn, TypeSystem)
	return p
}
