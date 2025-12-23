package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatMessage struct {
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	AvatarURL *string   `json:"avatarURL,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type chatClient struct {
	conn *websocket.Conn
	send chan []byte
}

type ChatHub struct {
	clients    map[*chatClient]bool
	broadcast  chan []byte
	register   chan *chatClient
	unregister chan *chatClient
	mu         sync.Mutex
}

var chatHubInstance *ChatHub
var chatOnce sync.Once

func GetChatHub() *ChatHub {
	chatOnce.Do(func() {
		chatHubInstance = &ChatHub{
			clients:    make(map[*chatClient]bool),
			broadcast:  make(chan []byte),
			register:   make(chan *chatClient),
			unregister: make(chan *chatClient),
		}
		go chatHubInstance.run()
	})
	return chatHubInstance
}

func (hub *ChatHub) run() {
	for {
		select {
		case client := <-hub.register:
			hub.mu.Lock()
			hub.clients[client] = true
			hub.mu.Unlock()
		case client := <-hub.unregister:
			hub.mu.Lock()
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				close(client.send)
			}
			hub.mu.Unlock()
		case msg := <-hub.broadcast:
			hub.mu.Lock()
			for client := range hub.clients {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(hub.clients, client)
				}
			}
			hub.mu.Unlock()
		}
	}
}

var chatUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeChatWs(c *gin.Context) {
	conn, err := chatUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	client := &chatClient{conn: conn, send: make(chan []byte, 256)}
	hub := GetChatHub()
	hub.register <- client

	go func() {
		for msg := range client.send {
			if err := client.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				break
			}
		}
		client.conn.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			break
		}
		var chatMsg ChatMessage
		if err := json.Unmarshal(message, &chatMsg); err != nil {
			log.Println("Invalid chat message", err)
			continue
		}
		if chatMsg.CreatedAt.IsZero() {
			chatMsg.CreatedAt = time.Now().UTC()
		}
		msgOut, _ := json.Marshal(chatMsg)
		hub.broadcast <- msgOut
	}

	hub.unregister <- client
	conn.Close()
}
