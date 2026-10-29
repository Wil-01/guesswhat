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

func newServer(l *Lobby, a *Admin) (http.Handler, error) {
	static, err := fs.Sub(webFiles, "web")
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(static))
	mux.HandleFunc("GET /api/rooms", l.handleList)
	mux.HandleFunc("POST /api/rooms", l.handleCreate)
	mux.HandleFunc("GET /api/themes", l.handleThemes)
	mux.Handle("/ws", websocket.Handler(l.handleWS))
	a.routes(mux)
	return mux, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func fail(msg string, err error) {
	fmt.Fprintln(os.Stderr, msg, err)
	os.Exit(84)
}

func main() {
	port := getenv("PORT", "8080")
	words, err := newWordStore(getenv("WORDS_DIR", "data/words"))
	if err != nil {
		fail("cannot load words:", err)
	}
	admin := &Admin{words: words, password: os.Getenv("ADMIN_PASSWORD")}
	if admin.password == "" {
		fmt.Println("admin page disabled: set ADMIN_PASSWORD to enable it")
	}
	handler, err := newServer(newLobby(words), admin)
	if err != nil {
		fail("server error:", err)
	}

	fmt.Println("listening on :" + port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fail("server error:", err)
	}
}
