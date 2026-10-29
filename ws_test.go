package main

import (
	"testing"
	"time"
)

func TestChatBroadcastWithNickname(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)

	say(t, alice.Conn, "hello")
	for _, c := range []*testClient{alice, bob} {
		p := expect(t, c.Conn, TypeMessage)
		if p.Message != "hello" || p.Nickname != alice.nickname {
			t.Fatalf("got %+v", p)
		}
	}
}

func TestRename(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)

	send(t, alice.Conn, NicknamePacket{Type: TypeNickname, Nickname: "  alice "})
	if got := expect(t, alice.Conn, TypeNickname).Nickname; got != "alice" {
		t.Fatalf("got nickname %q", got)
	}
	expect(t, alice.Conn, TypeSystem)
	expect(t, bob.Conn, TypeSystem)
	players := expect(t, bob.Conn, TypePlayers).Players
	if players[0].Nickname != "alice" {
		t.Fatalf("got players %+v", players)
	}

	say(t, alice.Conn, "hi")
	if got := expect(t, bob.Conn, TypeMessage).Nickname; got != "alice" {
		t.Fatalf("got nickname %q", got)
	}
}

func TestInvalidNicknameRejected(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)

	send(t, alice.Conn, NicknamePacket{Type: TypeNickname, Nickname: "   "})
	expect(t, alice.Conn, TypeError)
	say(t, alice.Conn, "still here")
	if got := expect(t, alice.Conn, TypeMessage).Nickname; got != alice.nickname {
		t.Fatalf("got nickname %q, want %q", got, alice.nickname)
	}
}

func TestHistorySentOnJoin(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	say(t, alice.Conn, "first")
	say(t, alice.Conn, "second")
	expect(t, alice.Conn, TypeMessage)
	expect(t, alice.Conn, TypeMessage)

	h := connect(t, srv).history
	if len(h) != 2 || h[0].Message != "first" || h[1].Message != "second" {
		t.Fatalf("got history %+v", h)
	}
}

func TestUnknownTypeRejected(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	send(t, alice.Conn, Envelope{Type: "bogus"})
	expect(t, alice.Conn, TypeError)
	say(t, alice.Conn, "ok")
	expect(t, alice.Conn, TypeMessage)
}

func TestClientRemovedOnDisconnect(t *testing.T) {
	hub, srv := newTestServer(t)
	alice := connect(t, srv)
	alice.Close()

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
	alice := connect(t, srv)
	bob := connect(t, srv)

	send(t, alice.Conn, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	if got := expect(t, bob.Conn, TypeDraw).Stroke; got != testStroke {
		t.Fatalf("got stroke %+v", got)
	}
	say(t, alice.Conn, "done")
	expect(t, alice.Conn, TypeMessage)
}

func TestDrawingSentOnJoin(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	send(t, alice.Conn, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	say(t, alice.Conn, "sync")
	expect(t, alice.Conn, TypeMessage)

	d := connect(t, srv).drawing
	if len(d) != 1 || d[0] != testStroke {
		t.Fatalf("got drawing %+v", d)
	}
}

func TestClear(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	send(t, alice.Conn, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	expect(t, bob.Conn, TypeDraw)

	send(t, alice.Conn, Envelope{Type: TypeClear})
	expect(t, alice.Conn, TypeClear)
	expect(t, bob.Conn, TypeClear)
	if d := connect(t, srv).drawing; len(d) != 0 {
		t.Fatalf("got drawing %+v", d)
	}
}

func TestInvalidStrokeRejected(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)

	bad := []Stroke{
		{From: Point{-1, 0}, To: Point{0.5, 0.5}, Color: "#000000", Size: 4},
		{From: Point{0, 0}, To: Point{0.5, 0.5}, Color: "red", Size: 4},
		{From: Point{0, 0}, To: Point{0.5, 0.5}, Color: "#000000", Size: 500},
	}
	for _, s := range bad {
		send(t, alice.Conn, DrawPacket{Type: TypeDraw, Stroke: s})
		expect(t, alice.Conn, TypeError)
	}
	say(t, alice.Conn, "after")
	expect(t, bob.Conn, TypeMessage)
}
