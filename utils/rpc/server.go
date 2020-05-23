package rpc

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
)

type Server struct {
	Address     string
	rpcServer   *rpc.Server
	listener    net.Listener
	done        chan struct{}
	connections map[int]net.Conn
	lock        sync.Mutex
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

func (t *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	t.listener = listener
	t.Address = listener.Addr().String()
	return nil
}

func (t *Server) Serve() error {
	connectionIndex := 0
	for {
		conn, err := t.listener.Accept()
		// Check if closed
		select {
		case <-t.done:
			return nil
		default:
		}

		if err != nil {
			return err
		}
		connectionIndex++
		t.lock.Lock()
		t.connections[connectionIndex] = conn
		t.lock.Unlock()

		// t.Logf("Received client connection  %s->%s", conn.RemoteAddr(), conn.LocalAddr())
		index := connectionIndex
		go func() {
			t.rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
			t.lock.Lock()
			delete(t.connections, index)
			t.lock.Unlock()
		}()
	}
}

func (t *Server) Close() {
	select {
	case <-t.done:
		// Already closed
		return
	default:
	}
	close(t.done)

	t.listener.Close()

	// Close all current connections
	t.lock.Lock()
	for _, conn := range t.connections {
		conn.Close()
	}
	for k := range t.connections {
		delete(t.connections, k)
	}
	t.lock.Unlock()
}
