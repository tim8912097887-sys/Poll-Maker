package vote

import (
	"encoding/json"
	"sync"
)

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]struct{}
}


func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]map[*Client]struct{}),
	}
}

func (m *RoomManager) Join(
	pollID string,
	client *Client,
) {
	m.mu.Lock()

	room, ok := m.rooms[pollID]
	if !ok {
		room = make(map[*Client]struct{})
		m.rooms[pollID] = room
	}

	room[client] = struct{}{}

	message, _ := json.Marshal(map[string]any{"type": "join", "client_id": client.clientID})
	// Unlock before broadcasting to avoid potential deadlocks if Broadcast tries to acquire the lock again
	m.mu.Unlock()
	m.Broadcast(pollID, []byte(message))
}

func (m *RoomManager) Leave(
	pollID string,
	client *Client,
) {
	m.mu.Lock()

	room, ok := m.rooms[pollID]
	if !ok {
		return
	}

	delete(room, client)

	if len(room) == 0 {
		delete(m.rooms, pollID)
	}

	message, _ := json.Marshal(map[string]any{"type": "leave", "client_id": client.clientID})
	
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
		client.send <- message
	}
}