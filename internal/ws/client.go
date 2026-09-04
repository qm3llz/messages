package ws

import (
	"context"
	"time"

	"github.com/coder/websocket"
)

const (
	writeWait  = 5 * time.Second
	pingPeriod = 30 * time.Second
	pingWait   = 10 * time.Second
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int
}

func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.CloseNow()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}

			wctx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Write(wctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return
			}

		case <-ticker.C:
			pctx, cancel := context.WithTimeout(ctx, pingWait)
			err := c.conn.Ping(pctx)
			cancel()
			if err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) readPump(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		c.hub.unregister <- c
		cancel()
	}()

	ctx = c.conn.CloseRead(ctx)
	<-ctx.Done()
}
