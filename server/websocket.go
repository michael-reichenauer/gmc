package server

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

// define our WebSocket endpoint
func ServeWs(w http.ResponseWriter, r *http.Request) {
	log.Infof("Connected %s", r.Host)

	// upgrade this connection to a WebSocket connection
	ws, err := upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}
	go writer(ws)
	reader(ws)
}

// We'll need to define an upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// We'll need to check the origin of our connection
	// this will allow us to make requests from our React
	// development server to here.
	// For now, we'll do no checking and just allow any connection
	CheckOrigin: func(r *http.Request) bool { return true },
}

func upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warnf("%v", err)
		return ws, err
	}
	return ws, nil
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Warnf("%v", err)
			return
		}
		// print out that message for clarity
		log.Infof("Message %q", string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Warnf("%v", err)
			return
		}

	}
}
func writer(conn *websocket.Conn) {
	for {
		log.Infof("Sending")
		messageType, r, err := conn.NextReader()
		if err != nil {
			log.Warnf("%v", err)
			return
		}
		w, err := conn.NextWriter(messageType)
		if err != nil {
			log.Warnf("%v", err)
			return
		}
		if _, err := io.Copy(w, r); err != nil {
			log.Warnf("%v", err)
			return
		}
		if err := w.Close(); err != nil {
			log.Warnf("%v", err)
			return
		}
	}
}
