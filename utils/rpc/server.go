package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"strings"
	"sync"

	"github.com/imkira/go-observer"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/rs/cors"
	"golang.org/x/net/websocket"
)

const defaultServiceName = "api"

// Server supports rpc server functionality
type Server struct {
	RPCURL    string
	EventsURL string
	rpcServer *rpc.Server

	done          chan struct{}
	connections   map[int]io.ReadWriteCloser
	lock          sync.Mutex
	httpServer    *http.Server
	listener      net.Listener
	connectionID  int
	eventChannels map[string]observer.Property
	eventsPath    string
}

// NewServer returns a new rpc server
func NewServer() *Server {
	return &Server{
		rpcServer:     rpc.NewServer(),
		done:          make(chan struct{}),
		connections:   make(map[int]io.ReadWriteCloser),
		eventChannels: make(map[string]observer.Property),
	}
}

// RegisterService to register a rpc service
func (t *Server) RegisterService(serviceName string, service interface{}) error {
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	return t.rpcServer.RegisterName(serviceName, service)
}

// Start to start rpc server
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

	t.RPCURL = fmt.Sprintf("ws://%s%s", listener.Addr().String(), u.Path)
	t.EventsURL = fmt.Sprintf("http://%s%s", listener.Addr().String(), eventsPath)

	// Websocket
	mux.Handle(u.Path, websocket.Handler(t.webSocketHandler))
	mux.HandleFunc(eventsPath, t.eventHandler)

	handler := cors.Default().Handler(mux)

	t.httpServer = &http.Server{Handler: handler}
	t.listener = listener
	log.Infof("Started rpc server on %s", t.RPCURL)
	return nil
}

// Serve runs the https server
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

// CreateEvents creates event channel
func (t *Server) CreateEvents(id string) {
	log.Infof("Create events %s", id)
	t.lock.Lock()
	t.eventChannels[id] = observer.NewProperty(nil)
	t.lock.Unlock()
}

func (t *Server) CloseEvents(id string) {
	log.Infof("Close events %s", id)
	t.lock.Lock()
	events, ok := t.eventChannels[id]
	if ok {
		events.Update(nil)
		delete(t.eventChannels, id)
	}
	t.lock.Unlock()
}

// PostEvent to post events for the specified id
func (t *Server) PostEvent(id string, event interface{}) {
	//log.Infof("Post events %s %v", id, event)
	if event == nil {
		return
	}

	t.lock.Lock()
	events, ok := t.eventChannels[id]
	if ok {
		events.Update(event)
	}
	t.lock.Unlock()
}

// Close the rpc server
func (t *Server) Close() {
	log.Infof("Closing %s ...", t.RPCURL)
	select {
	case <-t.done:
		// Already closed
		return
	default:
	}

	close(t.done)

	t.lock.Lock()
	for id, events := range t.eventChannels {
		events.Update(nil)
		delete(t.eventChannels, id)
	}
	t.lock.Unlock()

	// Close server for new connections
	t.httpServer.Close()
	t.closeAllCurrentConnections()
	log.Infof("Closed %s", t.RPCURL)
}

func (t *Server) webSocketHandler(conn *websocket.Conn) {
	log.Infof("Connected %s->%s", conn.RemoteAddr(), t.RPCURL)

	// Keep track of current connections so they can be closed when closing server
	connection := &connection{conn: conn}
	id := t.storeConnection(connection)

	t.rpcServer.ServeCodec(jsonrpc.NewServerCodec(connection))
	t.removeConnection(id)
	log.Infof("Disconnected %s->%s", conn.RemoteAddr(), t.RPCURL)
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
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	log.Infof("Events request: %q", req.RequestURI)
	index := strings.LastIndex(req.RequestURI, t.eventsPath)
	if index == -1 || len(req.RequestURI) == index+len(t.eventsPath) {
		log.Warnf("Invalid id")
		http.Error(rw, "invalid id", http.StatusBadRequest)
		return
	}
	id := req.RequestURI[index+len(t.eventsPath):]

	var eventsStream observer.Stream
	t.lock.Lock()
	events, ok := t.eventChannels[id]
	if ok {
		eventsStream = events.Observe()
	}
	t.lock.Unlock()

	if eventsStream == nil {
		log.Warnf("Unknown events id: %s", id)
		http.Error(rw, "Unknown id", http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Remove this eventChannel from the map of eventChannels when this handler exits.
	defer func() {
		//	broker.closingClients <- messageChan
		log.Infof("Closed eventHandler %q", id)
	}()

	// flush to signal to client that event stream is open (no need to wait for first event)
	flusher.Flush()
	for {
		select {
		case <-req.Context().Done():
			log.Infof("Request was closed")
			return
		case <-t.done:
			log.Infof("Server was closed")
			return
		case <-eventsStream.Changes():
			eventsStream.Next()
			event := eventsStream.Value()
			if event == nil {
				log.Infof("Event stream was closed")
				return
			}

			eventBytes, err := json.Marshal(event)
			if err != nil {
				panic(log.Fatal(err))
			}
			//log.Infof("Send %s", string(eventBytes))
			fmt.Fprintf(rw, "data: %s\n\n", string(eventBytes))
		}

		flusher.Flush()
	}
}
