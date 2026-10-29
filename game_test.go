package main

import "testing"

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

	send(t, alice.Conn, Envelope{Type: TypeStart})
	for _, c := range []*testClient{alice, bob} {
		if got := expect(t, c.Conn, TypeGame).State; got != StatePlaying {
			t.Fatalf("got state %q", got)
		}
		expect(t, c.Conn, TypeSystem)
	}
	send(t, alice.Conn, Envelope{Type: TypeStart})
	if got := expect(t, alice.Conn, TypeError).Message; got != errAlreadyStarted.Error() {
		t.Fatalf("got error %q", got)
	}
}
