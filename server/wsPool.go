package server

import (
	"github.com/michael-reichenauer/gmc/utils/log"
)

type wsPool struct {
	Register   chan *wsClient
	Unregister chan *wsClient
	Clients    map[*wsClient]bool
	Broadcast  chan Message
}

func NewWsPool() *wsPool {
	return &wsPool{
		Register:   make(chan *wsClient),
		Unregister: make(chan *wsClient),
		Clients:    make(map[*wsClient]bool),
		Broadcast:  make(chan Message),
	}
}

func (pool *wsPool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			log.Infof("Size of Connection Pool: %d", len(pool.Clients))
			for client, _ := range pool.Clients {
				// log.Infof("%s", client)
				client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
			}
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			log.Infof("Size of Connection Pool: %d", len(pool.Clients))
			for client, _ := range pool.Clients {
				client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
			}
		case message := <-pool.Broadcast:
			log.Infof("Sending message to all clients in Pool")
			for client, _ := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					log.Warnf("%s", err)
					return
				}
			}
		}
	}
}
