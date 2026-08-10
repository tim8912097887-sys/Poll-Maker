package websocket_server

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/features/vote"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/response"
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
		

	    h.logger.Info(
			"websocket connection attempt",
			slog.String("poll_id", pollID),
		)

		if pollID == "" {
			h.logger.Error("missing poll ID in request")
			err := conn.WriteJSON(response.NewErrorResponse("missing_poll_id", "poll ID is required", nil))
			if err != nil {
				h.logger.Error(
					"failed to write error response to websocket",
					slog.Any("error", err),
				)
			}
			return
		}

		// Check if the poll exists and is valid for voting
		pollMeta, err := h.voteCache.GetPollMeta(ctx, pollID)
		if err != nil {
		    h.logger.Error(
				"failed to get poll meta",
				slog.Any("error", err),
			)
			err = conn.WriteJSON(response.NewErrorResponse("failed_to_get_poll_meta", "failed to get poll meta", nil))
			if err != nil {
				h.logger.Error(
					"failed to write error response to websocket",
					slog.Any("error", err),
				)
			}
			conn.Close()
			return
		}
		if pollMeta.ExpiredAt.Before(time.Now()) {
			h.logger.Error(
				"poll has expired",
				slog.String("poll_id", pollID),
			)
			err = conn.WriteJSON(response.NewErrorResponse("poll_expired", "poll has expired", nil))
			if err != nil {
				h.logger.Error(
					"failed to write error response to websocket",
					slog.Any("error", err),
				)
			}
			conn.Close()
			return
		}
		if pollMeta.StartedAt.After(time.Now()) {
			h.logger.Error(
				"poll has not started yet",
				slog.String("poll_id", pollID),
			)
			err = conn.WriteJSON(response.NewErrorResponse("poll_not_started", "poll has not started yet", nil))
			if err != nil {
				h.logger.Error(
					"failed to write error response to websocket",
					slog.Any("error", err),
				)
			}
			conn.Close()
			return
		}
	

		client := vote.NewClient(
			pollID,
			conn,
			h.logger,
		)

		h.rooms.Join(pollID, client)

		defer func() {
			h.rooms.Leave(pollID, client)
			
			_ = conn.Close()
		}()

		client.Run(ctx)
}
