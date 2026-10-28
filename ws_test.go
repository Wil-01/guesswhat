package main

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func dial(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	ws, err := websocket.Dial(url, "", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return ws
}

func waitClients(t *testing.T, n int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for hub.count() != n {
		if time.Now().After(deadline) {
			t.Fatalf("expected %d clients, got %d", n, hub.count())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestChatBroadcast(t *testing.T) {
	srv := httptest.NewServer(websocket.Handler(handleWS))
	defer srv.Close()
	alice := dial(t, srv)
	defer alice.Close()
	bob := dial(t, srv)
	defer bob.Close()
	waitClients(t, 2)

	if err := websocket.JSON.Send(alice, Message{Message: "hello"}); err != nil {
		t.Fatal(err)
	}
	for _, ws := range []*websocket.Conn{alice, bob} {
		var got Message
		if err := websocket.JSON.Receive(ws, &got); err != nil {
			t.Fatal(err)
		}
		if got.Message != "hello" {
			t.Fatalf("got %q, want %q", got.Message, "hello")
		}
	}
}

func TestClientRemovedOnDisconnect(t *testing.T) {
	srv := httptest.NewServer(websocket.Handler(handleWS))
	defer srv.Close()
	ws := dial(t, srv)
	waitClients(t, 1)
	ws.Close()
	waitClients(t, 0)
}
