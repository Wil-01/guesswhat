package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

const (
	codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	codeLength   = 5
)

var errRoomNotFound = errors.New("room not found")

// Lock order: a room may lock the lobby, never the other way around.
type Lobby struct {
	mu       sync.Mutex
	rooms    map[string]*Room
	emptyTTL time.Duration
	setup    func(*Room)
}

func newLobby() *Lobby {
	return &Lobby{rooms: make(map[string]*Room), emptyTTL: defaultEmptyTTL}
}

func (l *Lobby) create(private bool, s Settings) *Room {
	room := newRoom()
	room.private = private
	room.applySettings(s)
	room.lobby = l

	l.mu.Lock()
	room.code = l.newCode()
	room.emptyTTL = l.emptyTTL
	if l.setup != nil {
		l.setup(room)
	}
	l.rooms[room.code] = room
	l.mu.Unlock()

	room.mu.Lock()
	room.scheduleCleanup()
	room.mu.Unlock()
	return room
}

func (l *Lobby) newCode() string {
	for {
		b := make([]byte, codeLength)
		for i := range b {
			b[i] = codeAlphabet[rand.IntN(len(codeAlphabet))]
		}
		if _, ok := l.rooms[string(b)]; !ok {
			return string(b)
		}
	}
}

func (l *Lobby) get(code string) *Room {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rooms[strings.ToUpper(code)]
}

func (l *Lobby) remove(code string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.rooms, code)
}

func (l *Lobby) publicRooms() []RoomInfo {
	l.mu.Lock()
	var rooms []*Room
	for _, room := range l.rooms {
		if !room.private {
			rooms = append(rooms, room)
		}
	}
	l.mu.Unlock()

	infos := make([]RoomInfo, len(rooms))
	for i, room := range rooms {
		infos[i] = room.info()
	}
	slices.SortFunc(infos, func(a, b RoomInfo) int { return cmp.Compare(a.Code, b.Code) })
	return infos
}

func (l *Lobby) handleList(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, l.publicRooms())
}

func (l *Lobby) handleCreate(w http.ResponseWriter, req *http.Request) {
	body := struct {
		Private bool `json:"private"`
		Settings
	}{Settings: defaultSettings()}
	if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 1024)).Decode(&body); err != nil && err != io.EOF {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := body.Settings.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	room := l.create(body.Private, body.Settings)
	writeJSON(w, http.StatusCreated, map[string]string{"code": room.code})
}

func handleThemes(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, themeNames())
}

func (l *Lobby) handleWS(conn *websocket.Conn) {
	room := l.get(conn.Request().URL.Query().Get("room"))
	if room == nil {
		sendError(conn, errRoomNotFound)
		conn.Close()
		return
	}
	room.handleWS(conn)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Fprintln(os.Stderr, "write error:", err)
	}
}
