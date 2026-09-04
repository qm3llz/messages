package ws

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
)

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, userID int) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*"},
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
