package handler

import (
	"net/http"

	"messages/internal/ws"
)

type WSHandler struct {
	hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, errNoUserInContext)
		return
	}

	ws.ServeWS(h.hub, w, r, userID)
}
