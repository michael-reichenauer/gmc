package server

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"net/http"

	"github.com/gorilla/websocket"
)

// define our WebSocket endpoint
func ServeWs(pool *Pool, w http.ResponseWriter, r *http.Request) {
	log.Infof("Connecting from %s", r.Host)

	conn, err := upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	client := &Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	log.Infof("Reading ... from %s", r.Host)
	client.Read()
	log.Infof("Done reading from %s", r.Host)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warnf("%v", err)
		return nil, err
	}

	return conn, nil
}
