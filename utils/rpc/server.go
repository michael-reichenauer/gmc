package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/rs/cors"
	"golang.org/x/net/websocket"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultServiceName = "api"

type Server struct {
	URL       string
	EventsURL string
	rpcServer *rpc.Server

	done          chan struct{}
	connections   map[int]io.ReadWriteCloser
	lock          sync.Mutex
	httpServer    *http.Server
	listener      net.Listener
	connectionID  int
	eventChannels map[chan []byte]string
	eventsPath    string
}

func NewServer() *Server {
	return &Server{
		rpcServer:     rpc.NewServer(),
		done:          make(chan struct{}),
		connections:   make(map[int]io.ReadWriteCloser),
		eventChannels: make(map[chan []byte]string),
	}
}

func (t *Server) RegisterService(serviceName string, service interface{}) error {
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	return t.rpcServer.RegisterName(serviceName, service)
}

func (t *Server) Start(uri, eventsPath string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(eventsPath, "/") {
		eventsPath = eventsPath + "/"
	}
	t.eventsPath = eventsPath

	listener, err := net.Listen("tcp", u.Host)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	t.URL = fmt.Sprintf("ws://%s%s", listener.Addr().String(), u.Path)

	// Websocket
	mux.Handle(u.Path, websocket.Handler(t.webSocketHandler))
	mux.HandleFunc(eventsPath, t.eventHandler)

	handler := cors.Default().Handler(mux)

	t.httpServer = &http.Server{Handler: handler}
	t.listener = listener
	log.Infof("Started rpc server on %s", t.URL)
	return nil
}

func (t *Server) Serve() error {
	err := t.httpServer.Serve(t.listener)

	select {
	case <-t.done:
		// Closed, no error
		return nil
	default:
	}

	return err
}

func (t *Server) PostEvent(id string, event interface{}) {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		panic(log.Fatal(err))
	}

	var eventChannels []chan []byte
	t.lock.Lock()
	for eventChannel, channelID := range t.eventChannels {
		if id == channelID {
			eventChannels = append(eventChannels, eventChannel)
		}
	}
	t.lock.Unlock()

	for _, eventChannel := range eventChannels {
		select {
		case eventChannel <- eventBytes:
		case <-t.done:
			// Closed, no error
			return
		}
	}
}

func (t *Server) Close() {
	log.Infof("Closing %s ...", t.URL)
	select {
	case <-t.done:
		// Already closed
		return
	default:
	}

	close(t.done)

	t.lock.Lock()
	for eventChannel := range t.eventChannels {
		close(eventChannel)
		delete(t.eventChannels, eventChannel)
	}
	t.lock.Unlock()

	// Close server for new connections
	t.httpServer.Close()
	t.closeAllCurrentConnections()
	log.Infof("Closed %s", t.URL)
}

func (t *Server) webSocketHandler(conn *websocket.Conn) {
	log.Infof("Connected %s->%s", conn.RemoteAddr(), t.URL)

	// Keep track of current connections so they can be closed when closing server
	connection := &connection{conn: conn}
	id := t.storeConnection(connection)

	t.rpcServer.ServeCodec(jsonrpc.NewServerCodec(connection))
	t.removeConnection(id)
	log.Infof("Disconnected %s->%s", conn.RemoteAddr(), t.URL)
}

func (t *Server) storeConnection(conn io.ReadWriteCloser) int {
	var id int
	t.lock.Lock()
	t.connectionID++
	id = t.connectionID
	t.connections[id] = conn
	t.lock.Unlock()
	return id
}

func (t *Server) removeConnection(id int) {
	t.lock.Lock()
	delete(t.connections, id)
	t.lock.Unlock()
}

func (t *Server) closeAllCurrentConnections() {
	// Close and delete all current connections
	t.lock.Lock()
	for _, conn := range t.connections {
		conn.Close()
	}
	for k := range t.connections {
		delete(t.connections, k)
	}
	t.lock.Unlock()
}

func (t *Server) eventHandler(rw http.ResponseWriter, req *http.Request) {
	log.Warnf("Events start")
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	log.Warnf("Request: %q", req.RequestURI)
	index := strings.LastIndex(req.RequestURI, t.eventsPath)
	if index == -1 || len(req.RequestURI) == index+len(t.eventsPath) {
		log.Warnf("Invalid id")
		http.Error(rw, "invalid id", http.StatusBadRequest)
		return
	}
	id := req.RequestURI[index+len(t.eventsPath):]
	log.Infof("id: %d %q %q", index, req.RequestURI, id)

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	eventChannel := make(chan []byte)
	t.lock.Lock()
	t.eventChannels[eventChannel] = id
	t.lock.Unlock()

	// Remove this eventChannel from the map of eventChannels when this handler exits.
	defer func() {
		//	broker.closingClients <- messageChan
		t.lock.Lock()
		delete(t.eventChannels, eventChannel)
		t.lock.Unlock()
		log.Warnf("Closed")
	}()

	time.AfterFunc(5*time.Second, func() {
		t.PostEvent(id, "some event")
	})

	// flush to signal to client that event stream is open (no need to wait for first event)
	flusher.Flush()
	for {
		select {
		case <-req.Context().Done():
			return
		case event, ok := <-eventChannel:
			if !ok {
				return
			}
			log.Warnf("Send %s", string(event))
			fmt.Fprintf(rw, "data: %s\n\n", string(event))
		}

		flusher.Flush()
	}
}
