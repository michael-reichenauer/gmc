package server

import (
	"github.com/michael-reichenauer/gmc/utils/log"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	ID   string
	Conn *websocket.Conn
	Pool *wsPool
}

type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

func (c *wsClient) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Warnf("%v", err)
			return
		}
		message := Message{Type: messageType, Body: string(p)}
		log.Infof("Message Received: %+v\n", message)
		c.Pool.Broadcast <- message
		log.Infof("Message broadcast %+v\n", message)
	}
}
