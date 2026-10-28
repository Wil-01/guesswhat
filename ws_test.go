package main

import (
	"net/http/httptest"
	"strings"
	"testing"

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

func TestEchoUppercase(t *testing.T) {
	srv := httptest.NewServer(websocket.Handler(handleWS))
	defer srv.Close()
	ws := dial(t, srv)
	defer ws.Close()

	if err := websocket.Message.Send(ws, "hello"); err != nil {
		t.Fatal(err)
	}
	var got string
	if err := websocket.Message.Receive(ws, &got); err != nil {
		t.Fatal(err)
	}
	if got != "HELLO" {
		t.Fatalf("got %q, want %q", got, "HELLO")
	}
}
