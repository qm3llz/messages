package ws

import (
	"context"
	"net/http"
	"os"

	"github.com/coder/websocket"
)

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, userID int) {
	domen := os.Getenv("WS_ALLOWED_ORIGINS")
	if domen == "" {
		domen = "localhost:*"
	} // TODO: add split for many patterns

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{domen},
	})
	if err != nil {
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 16),
		userID: userID,
	}
	hub.register <- client

	ctx, cancel := context.WithCancel(r.Context())
	go client.writePump(ctx)
	client.readPump(ctx, cancel)
}
