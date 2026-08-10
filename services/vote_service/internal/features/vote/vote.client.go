package vote

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/google/uuid"
)

type Client struct {
	clientID string
	pollID string
	conn   *websocket.Conn
	logger *slog.Logger
	send chan []byte
}

func NewClient(
	pollID string,
	conn *websocket.Conn,
	logger *slog.Logger,
) *Client {
	return &Client{
		clientID: uuid.New().String(),
		pollID: pollID,
		conn:   conn,
		logger: logger,
		send:   make(chan []byte, 32),
	}
}

func (c *Client) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go c.writeLoop(ctx)

	c.ReadLoop(ctx)
}

func (c *Client) ReadLoop(ctx context.Context) {
    defer func() {
        c.conn.Close()
    }()

    // Configure connection deadlines and heartbeats
    c.conn.SetReadLimit(512)
    _ = c.conn.SetReadDeadline(time.Now().Add(PongWait))
    
    // Reset read deadline every time a Pong is received from the client
    c.conn.SetPongHandler(func(string) error {
        _ = c.conn.SetReadDeadline(time.Now().Add(PongWait))
        return nil
    })

    for {
        // Discard any incoming bytes while letting the underlying library process Ping/Pong/Close
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            c.logger.Debug("websocket read error", slog.Any("error", err))
			break
        }
    }
}

func (c *Client) writeLoop(ctx context.Context) {

	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			_ = c.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			return

		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))

			if !ok {
				// The send channel was closed by the manager/hub
				_ = c.conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				return
			}

			// Write text message using Fiber/Gorilla signature
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.logger.Debug(
					"websocket write failed",
					slog.Any("error", err),
				)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Debug("websocket ping failed", slog.Any("error", err))
				return
			}
		}
	}
}