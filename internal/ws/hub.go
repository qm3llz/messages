package ws

import "context"

type Event struct {
	UserIDs []int
	Payload []byte
}

type Hub struct {
	clients    map[int]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event),
	}
}

func (h *Hub) Broadcast(ev Event) {
	h.broadcast <- ev
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {

		case c := <-h.register:
			if h.clients[c.userID] == nil {
				h.clients[c.userID] = make(map[*Client]bool)
			}
			h.clients[c.userID][c] = true

		case c := <-h.unregister:
			if _, ok := h.clients[c.userID][c]; ok {
				delete(h.clients[c.userID], c)
				close(c.send)
				if len(h.clients[c.userID]) == 0 {
					delete(h.clients, c.userID)
				}
			}

		case ev := <-h.broadcast:
			for _, uid := range ev.UserIDs {
				for c := range h.clients[uid] {
					select {

					case c.send <- ev.Payload:

					default:
						close(c.send)
						delete(h.clients[uid], c)
					}
				}
			}

		case <-ctx.Done():
			return
		}
	}
}
