package main

import (
	"strings"
	"testing"
)

func TestFirstClientIsAdmin(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)

	players := expect(t, alice.Conn, TypePlayers).Players
	if len(players) != 2 || !players[0].Admin || players[1].Admin || players[0].ID != alice.id || players[1].ID != bob.id {
		t.Fatalf("got players %+v", players)
	}
}

func TestAdminTransferredOnLeave(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	alice.Close()

	players := expect(t, bob.Conn, TypePlayers).Players
	if len(players) != 1 || players[0].ID != bob.id || !players[0].Admin {
		t.Fatalf("got players %+v", players)
	}
}

func TestStartRequiresAdmin(t *testing.T) {
	_, srv := newTestServer(t)
	connect(t, srv)
	bob := connect(t, srv)

	send(t, bob.Conn, Envelope{Type: TypeStart})
	if got := expect(t, bob.Conn, TypeError).Message; got != errNotAdmin.Error() {
		t.Fatalf("got error %q", got)
	}
}

func TestStartRequiresPlayers(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)

	send(t, alice.Conn, Envelope{Type: TypeStart})
	if got := expect(t, alice.Conn, TypeError).Message; got != errNotEnoughPlayers.Error() {
		t.Fatalf("got error %q", got)
	}
}

func TestStart(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)

	startGame(t, alice, alice, bob)
	send(t, alice.Conn, Envelope{Type: TypeStart})
	if got := expect(t, alice.Conn, TypeError).Message; got != errAlreadyStarted.Error() {
		t.Fatalf("got error %q", got)
	}
}

func TestWordOnlySentToDrawer(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	rounds := startGame(t, alice, alice, bob)

	drawer, guesser := rounds[alice], rounds[bob]
	if drawer.Drawer != alice.id || guesser.Drawer != alice.id {
		t.Fatalf("got drawers %d and %d", drawer.Drawer, guesser.Drawer)
	}
	if drawer.Word == "" || guesser.Word != "" {
		t.Fatalf("got words %q and %q", drawer.Word, guesser.Word)
	}
	if guesser.Hint != hint(drawer.Word) {
		t.Fatalf("got hint %q for %q", guesser.Hint, drawer.Word)
	}
}

func TestOnlyDrawerCanDraw(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	startGame(t, alice, alice, bob)

	send(t, bob.Conn, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	expect(t, bob.Conn, TypeError)
	send(t, bob.Conn, Envelope{Type: TypeClear})
	expect(t, bob.Conn, TypeError)

	send(t, alice.Conn, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	expect(t, bob.Conn, TypeDraw)
}

func TestDrawerCannotRevealWord(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	word := startGame(t, alice, alice, bob)[alice].Word

	say(t, alice.Conn, "it is a "+strings.ToUpper(word))
	expect(t, alice.Conn, TypeError)
	say(t, alice.Conn, "hint: it is big")
	expect(t, bob.Conn, TypeMessage)
}

func TestLateJoinerGetsRound(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	word := startGame(t, alice, alice, bob)[alice].Word
	send(t, alice.Conn, DrawPacket{Type: TypeDraw, Stroke: testStroke})
	expect(t, bob.Conn, TypeDraw)

	carol := connect(t, srv)
	if len(carol.drawing) != 1 {
		t.Fatalf("got drawing %+v", carol.drawing)
	}
	p := carol.round
	if p.Word != "" || p.Hint != hint(word) || p.Drawer != alice.id {
		t.Fatalf("got round %+v", p)
	}
}

func TestDrawerLeavingStopsGame(t *testing.T) {
	_, srv := newTestServer(t)
	alice := connect(t, srv)
	bob := connect(t, srv)
	carol := connect(t, srv)
	startGame(t, alice, alice, bob, carol)

	alice.Close()
	expect(t, bob.Conn, TypeSystem)
	if got := expect(t, bob.Conn, TypeGame).State; got != StateWaiting {
		t.Fatalf("got state %q", got)
	}
}
