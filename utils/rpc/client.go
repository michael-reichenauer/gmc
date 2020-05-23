package rpc

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"runtime"
	"strings"
)

type Client struct {
	serviceName string
	conn        net.Conn
	rpcClient   *rpc.Client
	done        chan struct{}
}

func NewRpcClient(serviceName string) *Client {
	return &Client{serviceName: serviceName + ".", done: make(chan struct{})}
}

func (t *Client) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	t.conn = conn
	t.rpcClient = jsonrpc.NewClient(conn)

	return nil
}

func (t *Client) Close() {
	select {
	case <-t.done:
		// Already closed
		return
	default:
	}
	close(t.done)

	t.conn.Close()
}

func (t *Client) Call(arg interface{}, rsp interface{}) error {
	callerName := t.callerMethodName()
	return t.rpcClient.Call(t.serviceName+callerName, arg, rsp)
}

func (*Client) callerMethodName() string {
	rpc := make([]uintptr, 1)
	n := runtime.Callers(3, rpc[:])
	if n < 1 {
		return ""
	}
	frame, _ := runtime.CallersFrames(rpc).Next()
	callerName := frame.Function
	i := strings.LastIndex(callerName, ".")
	if i != -1 {
		callerName = callerName[i+1:]
	}
	return callerName
}
