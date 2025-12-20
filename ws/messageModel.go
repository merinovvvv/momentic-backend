package ws

import "time"

type Message struct {
	Sender    string    `json:"sender"`    // Кто отправил
	Content   string    `json:"content"`   // Текст сообщения
	Timestamp time.Time `json:"timestamp"` // Время отправки
}
