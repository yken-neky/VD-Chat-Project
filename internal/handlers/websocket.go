package handlers

import (
	"New_VDChat/internal/services/room"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
	"time"
)

var rm = room.NewRoomManager()

func SetupWebSocketRoutes(app *fiber.App) {
	// Middleware para WebSocket
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/create", websocket.New(handleCreateRoom))
	app.Get("/ws/join/:roomID", websocket.New(handleJoinRoom))
}

func handleCreateRoom(c *websocket.Conn) {
	roomID := rm.CreateRoom()
	participantID, err := rm.JoinRoom(roomID, c)
	if err != nil {
		c.WriteJSON(map[string]string{"error": err.Error()})
		c.Close()
		return
	}
	defer rm.LeaveRoom(roomID, participantID)

	// Notificar al cliente
	if err := c.WriteJSON(map[string]string{
		"type":           "room_created",
		"room_id":        roomID,
		"participant_id": participantID,
	}); err != nil {
		log.Printf("Error en handleCreateRoom: %v", err)
	}

	// Escuchar mensajes continuamente
	for {
		var msg map[string]interface{}
		if err := c.ReadJSON(&msg); err != nil {
			log.Printf("Error leyendo mensaje: %v", err)
			break
		}

		// Broadcast del mensaje a todos los participantes
		if message, ok := msg["message"].(string); ok {
			rm.BroadcastMessage(roomID, participantID, message)
		}
	}
}

// internal/handlers/websocket.go
func handleJoinRoom(c *websocket.Conn) {
	roomID := c.Params("roomID")
	participantID, err := rm.JoinRoom(roomID, c)
	if err != nil {
		c.WriteJSON(map[string]string{"error": err.Error()})
		c.Close()
		return
	}
	defer rm.LeaveRoom(roomID, participantID)

	// Mantener conexión activa
	c.SetReadDeadline(time.Now().Add(60 * time.Minute)) // Timeout de 1 hora

	for {
		var msg map[string]interface{}
		if err := c.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Conexión cerrada inesperadamente: %v", err)
			}
			break
		}

		// Procesar mensajes
		if message, ok := msg["message"].(string); ok {
			rm.BroadcastMessage(roomID, participantID, message)
		}
	}
}
