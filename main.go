package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

const addr = ":8080"

func main() {
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.Handle("/ws", websocket.Handler(handleWS))

	fmt.Println("listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(84)
	}
}
