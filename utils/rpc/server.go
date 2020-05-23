package rpc

import (
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
)

type Server struct {
	Address   string
	rpcServer *rpc.Server

	done         chan struct{}
	connections  map[int]net.Conn
	lock         sync.Mutex
	httpServer   *http.Server
	listener     net.Listener
	connectionID int
}

func NewRpcServer() *Server {
	return &Server{
		rpcServer:   rpc.NewServer(),
		done:        make(chan struct{}),
		connections: make(map[int]net.Conn),
	}
}

func (t *Server) RegisterName(name string, service interface{}) error {
	return t.rpcServer.RegisterName(name, service)
}

func (t *Server) Start(address string, path string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc(path, t.httpRpcHandler)
	t.httpServer = &http.Server{Addr: "127.0.0.1:1234", Handler: mux}
	t.listener = listener
	t.Address = listener.Addr().String()
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

	// Keep track of current connections so they can be closed when closing server
	id := t.storeConnection(conn)
	t.rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
	t.removeConnection(id)
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
