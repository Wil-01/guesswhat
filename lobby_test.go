package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
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

func postRoom(t *testing.T, srv *httptest.Server, body any) *http.Response {
	t.Helper()
	data, _ := json.Marshal(body)
	res, err := http.Post(srv.URL+"/api/rooms", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func createRoom(t *testing.T, srv *httptest.Server, private bool) string {
	t.Helper()
	return createThemedRoom(t, srv, private, "")
}

func createThemedRoom(t *testing.T, srv *httptest.Server, private bool, theme string) string {
	t.Helper()
	res := postRoom(t, srv, map[string]any{"private": private, "theme": theme})
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

func TestCreateThemedRoom(t *testing.T) {
	_, srv := newLobbyServer(t)
	code := createThemedRoom(t, srv, false, "nature")
	if rooms := listRooms(t, srv); rooms[0].Theme != "nature" {
		t.Fatalf("got rooms %+v", rooms)
	}
	if got := joinRoom(t, srv, code).settings.Theme; got != "nature" {
		t.Fatalf("got theme %q", got)
	}
}

func TestCreateRoomInvalidSettings(t *testing.T) {
	_, srv := newLobbyServer(t)
	for _, body := range []map[string]any{
		{"theme": "nope"},
		{"rounds": 11},
		{"duration": 10},
	} {
		res := postRoom(t, srv, body)
		res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("%v: got status %d", body, res.StatusCode)
		}
	}
}

func TestCreateRoomWithSettings(t *testing.T) {
	lobby, srv := newLobbyServer(t)
	lobby.mu.Lock()
	lobby.setup = nil
	lobby.mu.Unlock()
	res := postRoom(t, srv, map[string]any{"rounds": 5, "duration": 120})
	defer res.Body.Close()
	var out struct{ Code string }
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	p := joinRoom(t, srv, out.Code).settings
	if p.Rounds != 5 || p.Duration != 120 || p.Theme != "" {
		t.Fatalf("got settings %+v", p)
	}
}

func TestListThemes(t *testing.T) {
	_, srv := newLobbyServer(t)
	res, err := http.Get(srv.URL + "/api/themes")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var got []string
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, themeNames()) {
		t.Fatalf("got themes %v", got)
	}
}
