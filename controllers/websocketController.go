package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/merinovvvv/momentic-backend/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketController struct {
	Hub *ws.Hub
}

func NewWebSocketController(hub *ws.Hub) *WebSocketController {
	return &WebSocketController{Hub: hub}
}

func (wc *WebSocketController) ServeWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("FATAL: Failed to upgrade HTTP connection to WebSocket: %v", err)
		return
	}

	client := ws.NewClient(wc.Hub, conn)
	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()

}
