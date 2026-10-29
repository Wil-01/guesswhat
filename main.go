package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

//go:embed web
var webFiles embed.FS

func newServer(l *Lobby) (http.Handler, error) {
	static, err := fs.Sub(webFiles, "web")
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(static))
	mux.HandleFunc("GET /api/rooms", l.handleList)
	mux.HandleFunc("POST /api/rooms", l.handleCreate)
	mux.Handle("/ws", websocket.Handler(l.handleWS))
	return mux, nil
}

func main() {
	if len(words) == 0 {
		fmt.Fprintln(os.Stderr, "no words to play with")
		os.Exit(84)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	handler, err := newServer(newLobby())
	if err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(84)
	}

	fmt.Println("listening on :" + port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(84)
	}
}
