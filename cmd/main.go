package main

import (
	"log"
	"net/http"

	"github.com/RX90/Chat/internal/ws"
	"github.com/RX90/Chat/web"
)

func main() {
	hub := ws.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	fs := http.FileServer(web.TemplatesFS())
	http.Handle("/", fs)

	log.Println("server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("server exited with error:", err)
	}
}
