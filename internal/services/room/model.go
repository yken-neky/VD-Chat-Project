package room

import (
	"github.com/gofiber/websocket/v2"
	"sync"
)

type Participant struct {
	ID       string
	Conn     *websocket.Conn
	IsActive bool
}

type Room struct {
	ID           string
	Participants map[string]*Participant
	mu           sync.RWMutex
}
