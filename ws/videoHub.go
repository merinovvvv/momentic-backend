package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/merinovvvv/momentic-backend/models"
)

// VideoHub управляет подписками на комментарии к конкретным видео
type VideoHub struct {
	// Карта: [video_id] -> [набор клиентов]
	rooms      map[int64]map[*CommentClient]bool
	roomsMutex sync.RWMutex

	// Канал для регистрации клиента в конкретной комнате
	Register chan *CommentClient

	// Канал для отмены регистрации
	Unregister chan *CommentClient

	// Канал для рассылки нового комментария
	BroadcastComment chan *models.Comment
}

type CommentClient struct {
	VideoID int64
	Conn    *websocket.Conn // Используйте существующий импорт gorilla/websocket
	Send    chan *models.CommentWSMessage
}

func NewVideoHub() *VideoHub {
	return &VideoHub{
		rooms:            make(map[int64]map[*CommentClient]bool),
		Register:         make(chan *CommentClient),
		Unregister:       make(chan *CommentClient),
		BroadcastComment: make(chan *models.Comment),
	}
}

func (h *VideoHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.roomsMutex.Lock()
			if h.rooms[client.VideoID] == nil {
				h.rooms[client.VideoID] = make(map[*CommentClient]bool)
			}
			h.rooms[client.VideoID][client] = true
			h.roomsMutex.Unlock()
			log.Printf("INFO: User joined room for video %d", client.VideoID)

		case client := <-h.Unregister:
			h.roomsMutex.Lock()
			if clients, ok := h.rooms[client.VideoID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.rooms, client.VideoID)
					}
				}
			}
			h.roomsMutex.Unlock()

		case comment := <-h.BroadcastComment:
			h.roomsMutex.RLock()
			msg := &models.CommentWSMessage{
				Type:    "new_comment",
				Payload: *comment,
			}
			// Рассылаем только тем, кто смотрит это видео
			if clients, ok := h.rooms[comment.VideoID]; ok {
				for client := range clients {
					select {
					case client.Send <- msg:
					default:
						// Если канал забит, удаляем клиента
						go func(c *CommentClient) { h.Unregister <- c }(client)
					}
				}
			}
			h.roomsMutex.RUnlock()
		}
	}
}
