package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newLobbyServer(t *testing.T) (*Lobby, *httptest.Server) {
	t.Helper()
	lobby := newLobby()
	lobby.setup = func(r *Room) {
		r.rounds = 1
		r.pauseDuration = 10 * time.Millisecond
	}
	handler, err := newServer(lobby)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return lobby, srv
}

func createRoom(t *testing.T, srv *httptest.Server, private bool) string {
	t.Helper()
	body, _ := json.Marshal(map[string]bool{"private": private})
	res, err := http.Post(srv.URL+"/api/rooms", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("got status %d", res.StatusCode)
	}
	var out struct{ Code string }
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Code) != codeLength {
		t.Fatalf("got code %q", out.Code)
	}
	return out.Code
}

func listRooms(t *testing.T, srv *httptest.Server) []RoomInfo {
	t.Helper()
	res, err := http.Get(srv.URL + "/api/rooms")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var rooms []RoomInfo
	if err := json.NewDecoder(res.Body).Decode(&rooms); err != nil {
		t.Fatal(err)
	}
	return rooms
}

func joinRoom(t *testing.T, srv *httptest.Server, code string) *testClient {
	t.Helper()
	return join(t, dial(t, srv, "/ws?room="+code))
}

func TestCreateAndListRooms(t *testing.T) {
	_, srv := newLobbyServer(t)
	public := createRoom(t, srv, false)
	createRoom(t, srv, true)

	rooms := listRooms(t, srv)
	if len(rooms) != 1 || rooms[0].Code != public || rooms[0].Players != 0 || rooms[0].State != StateWaiting {
		t.Fatalf("got rooms %+v", rooms)
	}
	joinRoom(t, srv, public)
	if rooms := listRooms(t, srv); rooms[0].Players != 1 {
		t.Fatalf("got rooms %+v", rooms)
	}
}

func TestRoomsAreIsolated(t *testing.T) {
	_, srv := newLobbyServer(t)
	alice := joinRoom(t, srv, createRoom(t, srv, false))
	bob := joinRoom(t, srv, createRoom(t, srv, true))

	say(t, alice.Conn, "only for room one")
	expect(t, alice.Conn, TypeMessage)
	say(t, bob.Conn, "only for room two")
	if got := expect(t, bob.Conn, TypeMessage).Message; got != "only for room two" {
		t.Fatalf("got message %q", got)
	}
}

func TestJoinUnknownRoom(t *testing.T) {
	_, srv := newLobbyServer(t)
	ws := dial(t, srv, "/ws?room=NOPE1")
	if got := expect(t, ws, TypeError).Message; got != errRoomNotFound.Error() {
		t.Fatalf("got error %q", got)
	}
	expectClosed(t, ws)
}

func TestRoomCodeIsCaseInsensitive(t *testing.T) {
	_, srv := newLobbyServer(t)
	code := createRoom(t, srv, false)
	joinRoom(t, srv, string(bytes.ToLower([]byte(code))))
}

func TestRoomClosedAfterGame(t *testing.T) {
	lobby, srv := newLobbyServer(t)
	code := createRoom(t, srv, false)
	alice := joinRoom(t, srv, code)
	bob := joinRoom(t, srv, code)
	r := startGame(t, alice, alice, bob)

	say(t, r.guesser().Conn, r.word)
	for _, c := range []*testClient{alice, bob} {
		expect(t, c.Conn, TypeSystem)
		expect(t, c.Conn, TypeRoundEnd)
		expect(t, c.Conn, TypeGameEnd)
		expectClosed(t, c.Conn)
	}
	if lobby.get(code) != nil {
		t.Fatal("room still registered")
	}
	ws := dial(t, srv, "/ws?room="+code)
	expect(t, ws, TypeError)
}

func TestEmptyRoomCleanedUp(t *testing.T) {
	lobby, srv := newLobbyServer(t)
	lobby.mu.Lock()
	lobby.emptyTTL = 50 * time.Millisecond
	lobby.mu.Unlock()
	code := createRoom(t, srv, false)

	deadline := time.Now().Add(2 * time.Second)
	for lobby.get(code) != nil {
		if time.Now().After(deadline) {
			t.Fatal("empty room not cleaned up")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestRoomKeptWhilePlayersConnected(t *testing.T) {
	lobby, srv := newLobbyServer(t)
	lobby.mu.Lock()
	lobby.emptyTTL = 50 * time.Millisecond
	lobby.mu.Unlock()
	code := createRoom(t, srv, false)
	joinRoom(t, srv, code)

	time.Sleep(150 * time.Millisecond)
	if lobby.get(code) == nil {
		t.Fatal("room removed while a player is connected")
	}
}

func TestStaticFilesServed(t *testing.T) {
	_, srv := newLobbyServer(t)
	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("got status %d", res.StatusCode)
	}
}
