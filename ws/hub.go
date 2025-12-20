package ws

import (
	"log"
	"time"
)

type Hub struct {
	clients map[*Client]bool

	Broadcast chan *Message

	Register chan *Client

	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
			log.Printf("INFO: New WebSocket client registered. Total clients: %d", len(h.clients))

			client.send <- &Message{
				Sender:    "System",
				Content:   "Подключено к каналу технической поддержки.",
				Timestamp: time.Now(),
			}

		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("INFO: WebSocket client unregistered. Total clients: %d", len(h.clients))
			}

		case message := <-h.Broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					log.Printf("WARNING: Client send channel blocked, client removed.")
				}
			}
			log.Printf("INFO: Broadcast message sent to %d clients.", len(h.clients))
		}
	}
}
