package vote

import (
	"context"
	"log/slog"
	"sync"
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
	done     chan struct{}
	closeOnce sync.Once
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
		done:     make(chan struct{}),
	}
}

func (c *Client) Run(ctx context.Context) {

	go c.writeLoop(ctx)

	c.ReadLoop(ctx)

	c.Close()
}

func (c *Client) ReadLoop(ctx context.Context) {
    defer func() {
        c.conn.Close()
    }()

    // Configure connection deadlines and heartbeats
    c.conn.SetReadLimit(512)
    err := c.conn.SetReadDeadline(time.Now().Add(PongWait))
    if err != nil {
		c.logger.Debug(
			"failed to set websocket read deadline",
			slog.Any("error", err),
		)
		return
	}

    // Reset read deadline every time a Pong is received from the client
    c.conn.SetPongHandler(func(string) error {
        _ = c.conn.SetReadDeadline(time.Now().Add(PongWait))
        return nil
    })

    for {
        // Discard any incoming bytes while letting the underlying library process Ping/Pong/Close
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
			) {
				c.logger.Debug(
					"websocket client disconnected",
					slog.String("client_id", c.clientID),
				)
			} else {
				c.logger.Debug(
					"websocket read failed",
					slog.String("client_id", c.clientID),
					slog.Any("error", err),
				)
			}
			return
        }
    }
}

func (c *Client) writeLoop(ctx context.Context) {

	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := c.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			) ; err != nil {
				c.logger.Debug(
					"websocket write failed",
					slog.Any("error", err),
				)
			}
			return

		case <-c.done:
			if err := c.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			); err != nil {
				c.logger.Debug(
					"websocket write failed",
					slog.Any("error", err),
				)
			}
			return

		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(
				time.Now().Add(WriteWait),
			); err != nil {
				c.logger.Debug(
					"failed to set websocket write deadline",
					slog.Any("error", err),
				)
				return
			}
			if !ok {
				// The send channel was closed by the manager/hub
				if err := c.conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				) ; err != nil {
					c.logger.Debug(
						"websocket write failed",
						slog.Any("error", err),
					)
				}
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
			if err := c.conn.SetWriteDeadline(
				time.Now().Add(WriteWait),
			); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Debug("websocket ping failed", slog.Any("error", err))
				return
			}
		}
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		_ = c.conn.Close()
	})
}