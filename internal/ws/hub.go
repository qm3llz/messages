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

func (h *Hub) Run(ctx context.Context) {
	for {
		select {

		case c := <-h.register:
			if h.clients[c.UserID] == nil {
				h.clients[c.userID] = make(map[*Client]bool)
			}

		case c := <-h.unregister:

		case ev := <-h.broadcast:

		case <-ctx.Done():
			return
		}
	}
}
