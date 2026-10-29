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
	Nickname string       `json:"nickname"`
	Message  string       `json:"message"`
	Messages []ChatPacket `json:"messages"`
	Stroke   Stroke       `json:"stroke"`
	Strokes  []Stroke     `json:"strokes"`
}

func newTestServer(t *testing.T) (*Hub, *httptest.Server) {
	t.Helper()
	hub := newHub()
	srv := httptest.NewServer(websocket.Handler(hub.handleWS))
	t.Cleanup(srv.Close)
	return hub, srv
}

func expect(t *testing.T, ws *websocket.Conn, typ string) anyPacket {
	t.Helper()
	ws.SetReadDeadline(time.Now().Add(time.Second))
	var p anyPacket
	if err := websocket.JSON.Receive(ws, &p); err != nil {
		t.Fatal(err)
	}
	if p.Type != typ {
		t.Fatalf("got packet type %q, want %q", p.Type, typ)
	}
	return p
}

// connect dials and consumes the welcome packets, so the client is registered on return.
func connect(t *testing.T, srv *httptest.Server) (*websocket.Conn, string, []ChatPacket) {
	ws, nick, history, _ := connectWithDrawing(t, srv)
	return ws, nick, history
}

func connectWithDrawing(t *testing.T, srv *httptest.Server) (*websocket.Conn, string, []ChatPacket, []Stroke) {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	ws, err := websocket.Dial(url, "", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ws.Close() })
	nick := expect(t, ws, TypeNickname).Nickname
	history := expect(t, ws, TypeHistory).Messages
	drawing := expect(t, ws, TypeDrawing).Strokes
	return ws, nick, history, drawing
}

func send(t *testing.T, ws *websocket.Conn, v any) {
	t.Helper()
	if err := websocket.JSON.Send(ws, v); err != nil {
		t.Fatal(err)
	}
}

func TestChatBroadcastWithNickname(t *testing.T) {
	_, srv := newTestServer(t)
	alice, aliceNick, _ := connect(t, srv)
	bob, _, _ := connect(t, srv)

	send(t, alice, ChatPacket{Type: TypeMessage, Message: "hello"})
	for _, ws := range []*websocket.Conn{alice, bob} {
		p := expect(t, ws, TypeMessage)
		if p.Message != "hello" || p.Nickname != aliceNick {
			t.Fatalf("got %+v", p)
		}
	}
}

func TestRename(t *testing.T) {
	_, srv := newTestServer(t)
	alice, _, _ := connect(t, srv)
	bob, _, _ := connect(t, srv)

	send(t, alice, NicknamePacket{Type: TypeNickname, Nickname: "  alice "})
	if got := expect(t, alice, TypeNickname).Nickname; got != "alice" {
		t.Fatalf("got nickname %q", got)
	}
	expect(t, alice, TypeSystem)
	expect(t, bob, TypeSystem)

	send(t, alice, ChatPacket{Type: TypeMessage, Message: "hi"})
	if got := expect(t, bob, TypeMessage).Nickname; got != "alice" {
		t.Fatalf("got nickname %q", got)
	}
}

func TestInvalidNicknameIgnored(t *testing.T) {
	_, srv := newTestServer(t)
	alice, nick, _ := connect(t, srv)

	send(t, alice, NicknamePacket{Type: TypeNickname, Nickname: "   "})
	send(t, alice, ChatPacket{Type: TypeMessage, Message: "still here"})
	if got := expect(t, alice, TypeMessage).Nickname; got != nick {
		t.Fatalf("got nickname %q, want %q", got, nick)
	}
}

func TestHistorySentOnJoin(t *testing.T) {
	_, srv := newTestServer(t)
	alice, _, _ := connect(t, srv)
	send(t, alice, ChatPacket{Type: TypeMessage, Message: "first"})
	send(t, alice, ChatPacket{Type: TypeMessage, Message: "second"})
	expect(t, alice, TypeMessage)
	expect(t, alice, TypeMessage)

	_, _, history := connect(t, srv)
	if len(history) != 2 || history[0].Message != "first" || history[1].Message != "second" {
		t.Fatalf("got history %+v", history)
	}
}

func TestUnknownTypeKeepsConnection(t *testing.T) {
	_, srv := newTestServer(t)
	alice, _, _ := connect(t, srv)
	send(t, alice, Envelope{Type: "bogus"})
	send(t, alice, ChatPacket{Type: TypeMessage, Message: "ok"})
	expect(t, alice, TypeMessage)
}

func TestClientRemovedOnDisconnect(t *testing.T) {
	hub, srv := newTestServer(t)
	ws, _, _ := connect(t, srv)
	ws.Close()

	deadline := time.Now().Add(time.Second)
	for {
		hub.mu.Lock()
		n := len(hub.clients)
		hub.mu.Unlock()
		if n == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("client not removed, %d left", n)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

var testStroke = Stroke{From: Point{0.1, 0.1}, To: Point{0.5, 0.5}, Color: "#ff0000", Size: 4}

func TestDrawBroadcastToOthers(t *testing.T) {
	_, srv := newTestServer(t)
	alice, _, _ := connect(t, srv)
	bob, _, _ := connect(t, srv)

	send(t, alice, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	if got := expect(t, bob, TypeDraw).Stroke; got != testStroke {
		t.Fatalf("got stroke %+v", got)
	}

	send(t, alice, ChatPacket{Type: TypeMessage, Message: "done"})
	expect(t, alice, TypeMessage)
}

func TestDrawingSentOnJoin(t *testing.T) {
	_, srv := newTestServer(t)
	alice, _, _ := connect(t, srv)
	send(t, alice, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	send(t, alice, ChatPacket{Type: TypeMessage, Message: "sync"})
	expect(t, alice, TypeMessage)

	_, _, _, drawing := connectWithDrawing(t, srv)
	if len(drawing) != 1 || drawing[0] != testStroke {
		t.Fatalf("got drawing %+v", drawing)
	}
}

func TestClear(t *testing.T) {
	_, srv := newTestServer(t)
	alice, _, _ := connect(t, srv)
	bob, _, _ := connect(t, srv)
	send(t, alice, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	expect(t, bob, TypeDraw)

	send(t, alice, Envelope{Type: TypeClear})
	expect(t, alice, TypeClear)
	expect(t, bob, TypeClear)

	_, _, _, drawing := connectWithDrawing(t, srv)
	if len(drawing) != 0 {
		t.Fatalf("got drawing %+v", drawing)
	}
}

func TestInvalidStrokeIgnored(t *testing.T) {
	_, srv := newTestServer(t)
	alice, _, _ := connect(t, srv)
	bob, _, _ := connect(t, srv)

	bad := []Stroke{
		{From: Point{-1, 0}, To: Point{0.5, 0.5}, Color: "#000000", Size: 4},
		{From: Point{0, 0}, To: Point{0.5, 0.5}, Color: "red", Size: 4},
		{From: Point{0, 0}, To: Point{0.5, 0.5}, Color: "#000000", Size: 500},
	}
	for _, s := range bad {
		send(t, alice, DrawPacket{Type: TypeDraw, Stroke: s})
	}
	send(t, alice, ChatPacket{Type: TypeMessage, Message: "after"})
	expect(t, bob, TypeMessage)
}
