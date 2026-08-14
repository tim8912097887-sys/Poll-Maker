package websocket_server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/features/vote"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared"
	websocketresponse "github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/response/websocket_response"
)

type WebSocketHandler struct {
	logger *slog.Logger
	rooms  *vote.RoomManager
	voteCache vote.VoteCache
}

type WebSocketHandlerConfig struct {
	Logger *slog.Logger
	VoteCache vote.VoteCache
	Rooms *vote.RoomManager
}

func NewWebSocketHandler(
   webSocketHandlerConfig *WebSocketHandlerConfig,
) *WebSocketHandler {
	return &WebSocketHandler{
		logger: webSocketHandlerConfig.Logger,
		rooms:  webSocketHandlerConfig.Rooms,
		voteCache: webSocketHandlerConfig.VoteCache,
	}
}

func (h *WebSocketHandler) HandleConnection(ctx context.Context, conn *websocket.Conn, pollID string) {
		

	    h.logger.Debug(
			"websocket connection attempt",
			slog.String("poll_id", pollID),
		)

		// Close the connection if the poll ID is empty or invalid
		defer func() {
			_ = conn.Close()
		}()

		if pollID == "" {
			h.logger.Error("missing poll ID in request")
			h.writeError(conn, "invalid_poll", "missing poll ID in request")
			return
		}

		// Validate the poll
		err := h.validatePoll(ctx, pollID)
		if err != nil {
			h.logger.Error("failed to validate poll", slog.Any("error", err))
			h.writeError(conn, "invalid_poll", err.Error())
			return
		}

		client := vote.NewClient(
			pollID,
			conn,
			h.logger,
		)


		err = h.rooms.Join(pollID, client)

		if err != nil {
			if err == shared.ErrRoomFull {
				h.logger.Error("room is full", slog.String("poll_id", pollID))
				h.writeError(conn, "room_full", "the room is full")
				return
			}

			h.logger.Error("failed to join room", slog.Any("error", err))
			h.writeError(conn, "internal_error", "failed to join room")
			return
		}

		defer func() {
			h.rooms.Leave(pollID, client)
			
			_ = conn.Close()
		}()

		client.Run(ctx)
}

func (h *WebSocketHandler) validatePoll(
	ctx context.Context,
	pollID string,
) error {
	meta, err := h.voteCache.GetPollMeta(ctx, pollID)
	if err != nil {
		return fmt.Errorf("get poll meta: %w", err)
	}

	now := time.Now()

	if meta.StartedAt.After(now) {
		return shared.ErrPollNotStarted
	}

	if !meta.ExpiredAt.After(now) {
		return shared.ErrPollExpired
	}

	return nil
}

func (h *WebSocketHandler) writeError(
	conn *websocket.Conn,
	code string,
	message string,
) {
	if err := conn.WriteJSON(
		websocketresponse.NewWSErrorResponse(code,message),
	); err != nil {
		h.logger.Debug(
			"failed to write websocket error",
			slog.String("code", code),
			slog.Any("error", err),
		)
	}
}