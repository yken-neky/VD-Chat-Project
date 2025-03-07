package main

import (
	"New_VDChat/internal/handlers"
	"github.com/gofiber/fiber/v2"
	_ "github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()

	// Configurar rutas WebSocket
	handlers.SetupWebSocketRoutes(app)

	err := app.Listen(":3000")
	if err != nil {
		return
	}
}
