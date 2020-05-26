package rpc

import (
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"sync"
)

const defaultServiceName = "api"

type Server struct {
	URL       string
	rpcServer *rpc.Server

	done         chan struct{}
	connections  map[int]net.Conn
	lock         sync.Mutex
	httpServer   *http.Server
	listener     net.Listener
	connectionID int
}

func NewServer() *Server {
	return &Server{
		rpcServer:   rpc.NewServer(),
		done:        make(chan struct{}),
		connections: make(map[int]net.Conn),
	}
}

func (t *Server) RegisterService(serviceName string, service interface{}) error {
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	return t.rpcServer.RegisterName(serviceName, service)
}

func (t *Server) Start(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", u.Host)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc(u.Path, t.httpRpcHandler)
	t.httpServer = &http.Server{Handler: mux}
	t.listener = listener
	t.URL = fmt.Sprintf("http://%s%s", listener.Addr().String(), u.Path)
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

func (t *Server) Close() {
	log.Infof("Closing %s ...", t.URL)
	select {
	case <-t.done:
		// Already closed
		return
	default:
	}
	close(t.done)

	// Close server for new connections
	t.httpServer.Close()
	t.closeAllCurrentConnections()
	log.Infof("Closed %s", t.URL)
}

func (t *Server) httpRpcHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		http.Error(w, "method must be connect", 405)
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	defer conn.Close()

	// Return response that connection was accepted
	_, _ = io.WriteString(conn, "HTTP/1.0 200 Connected\r\n\r\n")

	log.Infof("Connected %s->%s", req.RemoteAddr, t.URL)
	// Keep track of current connections so they can be closed when closing server
	id := t.storeConnection(conn)
	t.rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
	t.removeConnection(id)
	log.Infof("Disconnected %s->%s", req.RemoteAddr, t.URL)
}

func (t *Server) storeConnection(conn net.Conn) int {
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
