package vote

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared"
)

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]struct{}
}


func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms:  make(map[string]map[*Client]struct{}),
	}
}

func (m *RoomManager) Join(
	pollID string,
	client *Client,
) error {
	m.mu.Lock()

	room, ok := m.rooms[pollID]
	if !ok {
		room = make(map[*Client]struct{})
		m.rooms[pollID] = room
	}

	room[client] = struct{}{}

	if len(room) >= MaxClientsPerRoom {
		return shared.ErrRoomFull
	}

	message, err := json.Marshal(map[string]any{"type": "join", "client_id": client.clientID})
	if err != nil {
		client.logger.Error("Failed to marshal join message", slog.Any("error", err))
		m.mu.Unlock()
		return err
	}
	// Unlock before broadcasting to avoid potential deadlocks if Broadcast tries to acquire the lock again
	m.mu.Unlock()
	m.Broadcast(pollID, []byte(message))

	return nil
}

func (m *RoomManager) Leave(
	pollID string,
	client *Client,
) {
	m.mu.Lock()

	room, ok := m.rooms[pollID]
	if !ok {
		m.mu.Unlock()
		return
	}

	delete(room, client)

	if len(room) == 0 {
		delete(m.rooms, pollID)
	}

	message, err := json.Marshal(map[string]any{"type": "leave", "client_id": client.clientID})
	if err != nil {
		client.logger.Error("Failed to marshal leave message", slog.Any("error", err))
		m.mu.Unlock()
		return
	}

	// Unlock before broadcasting to avoid potential deadlocks if Broadcast tries to acquire the lock again
	m.mu.Unlock()
	m.Broadcast(pollID, []byte(message))
}

func (m *RoomManager) Broadcast(pollId string, message []byte) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	room, ok := m.rooms[pollId]
	if !ok {
		return
	}

	for client := range room {
		select {
		case client.send <- message:
		// Don't block if the client's send channel is full; log a warning and close the client
		default:
			client.logger.Warn(
				"client send buffer full",
				slog.String("client_id", client.clientID),
				slog.String("poll_id", pollId),
			)

			go client.Close()
		}
	}
}