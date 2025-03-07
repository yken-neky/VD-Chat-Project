package room

import (
	"errors"
	"fmt"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"log"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewRoomManager crear un nuevo manager
func NewRoomManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// Función para generar logs de la sala
func (r *Room) generateRoomLog() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participants := make([]string, 0, len(r.Participants))
	for _, p := range r.Participants {
		participants = append(participants, p.ID)
	}

	return fmt.Sprintf(
		"Sala ID: %s | Usuarios: %d | Participantes: [%s]",
		r.ID,
		len(r.Participants),
		strings.Join(participants, ", "),
	)
}

// JoinRoom Function para unir a las salas existentes
func (rm *Manager) JoinRoom(roomID string, ws *websocket.Conn) (string, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return "", errors.New("la sala no existe")
	}

	participantID := uuid.New().String()
	room.Participants[participantID] = &Participant{
		ID:   participantID,
		Conn: ws,
	}

	// Log de unión a sala
	log.Printf("[%s] Nuevo usuario en sala: %s\n%s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		participantID,
		room.generateRoomLog(),
	)
	log.Printf("[DEBUG] Usuario %s unido a sala %s (Conexión activa: %v)",
		participantID,
		roomID,
		ws.Conn != nil)

	return participantID, nil
}

// CreateRoom Function para crear salas
func (rm *Manager) CreateRoom() string {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	roomID := uuid.New().String()
	rm.rooms[roomID] = &Room{
		ID:           roomID,
		Participants: make(map[string]*Participant),
	}

	// Log de creación de sala
	log.Printf("[%s] Nueva sala creada\n%s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		rm.rooms[roomID].generateRoomLog(),
	)

	return roomID
}

// LeaveRoom saca de la sala al participante
func (rm *Manager) LeaveRoom(roomID, participantID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[roomID]; exists {
		delete(room.Participants, participantID)
		if len(room.Participants) == 0 {
			delete(rm.rooms, roomID)
		}
	}
}

func (rm *Manager) BroadcastMessage(roomID, senderID, message string) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if room, exists := rm.rooms[roomID]; exists {
		room.mu.RLock()
		defer room.mu.RUnlock()

		for _, participant := range room.Participants {
			if participant.ID != senderID {
				participant.Conn.WriteJSON(map[string]interface{}{
					"type":    "message",
					"from":    senderID,
					"content": message,
					"time":    time.Now().Format("2006-01-02 15:04:05"),
				})
			}
		}
		log.Printf("[%s] Mensaje en sala %s: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			roomID,
			message)
	}
}

func (rm *Manager) StartHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rm.mu.RLock()
		for _, room := range rm.rooms {
			room.mu.RLock()
			for _, participant := range room.Participants {
				participant.Conn.WriteJSON(map[string]string{
					"type": "heartbeat",
					"time": time.Now().Format(time.RFC3339),
				})
			}
			room.mu.RUnlock()
		}
		rm.mu.RUnlock()
	}
}
