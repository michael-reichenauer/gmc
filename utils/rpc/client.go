package rpc

import (
	"bufio"
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"runtime"
	"strings"
	"sync/atomic"
)

type ServiceClient interface {
	Call(arg interface{}, rsp interface{}) error
}

type serviceClient struct {
	serviceName string
	client      *Client
	isLogCalls  bool
}

type Client struct {
	IsLogCalls        bool
	OnConnectionError func(err error)
	connection        *connection
	rpcClient         *rpc.Client
	done              chan struct{}
	uri               string
}

func NewClient() *Client {
	return &Client{done: make(chan struct{})}
}

func (t *Client) Connect(uri string) error {
	conn, err := t.connect(uri)
	if err != nil {
		return err
	}
	t.uri = uri
	t.connection = &connection{conn: conn, onConnectionError: t.OnConnectionError}
	t.rpcClient = jsonrpc.NewClient(t.connection)
	return nil
}

type connection struct {
	conn              net.Conn
	onConnectionError func(err error)
	connErrors        int32
}

func (t *connection) Read(p []byte) (n int, err error) {
	n, err = t.conn.Read(p)
	if err != nil && t.onConnectionError != nil {
		if c := atomic.AddInt32(&t.connErrors, 1); c == 1 {
			t.onConnectionError(err)
		}
	}
	return
}

func (t *connection) Write(p []byte) (n int, err error) {
	n, err = t.conn.Write(p)
	if err != nil && t.onConnectionError != nil {
		if c := atomic.AddInt32(&t.connErrors, 1); c == 1 {
			t.onConnectionError(err)
		}
	}
	return
}

func (t *connection) Close() error {
	return t.conn.Close()
}

func (*Client) connect(uri string) (net.Conn, error) {
	log.Infof("Connecting to %s", uri)
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return nil, err
	}
	_, err = io.WriteString(conn, "CONNECT "+u.Path+" HTTP/1.0\n\n")
	if err != nil {
		log.Warnf("Failed to write CONNECT request to %s, %v", uri, err)
		_ = conn.Close()
		return nil, err
	}

	// Require successful HTTP response before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err != nil {
		log.Warnf("Failed to read CONNECT response from %s, %v", uri, err)
		_ = conn.Close()
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// OK response, create json rpc client
		err = fmt.Errorf("invalid http response, code: %d, %s", resp.StatusCode, resp.Status)
		log.Warnf("Failed to read CONNECT response from %s, %v", uri, err)
		_ = conn.Close()
		return nil, err
	}
	log.Infof("Connected %s->%s", conn.LocalAddr(), uri)
	return conn, nil
}

func (t *Client) Close() {
	select {
	case <-t.done:
		// Already closed
		return
	default:
	}
	log.Infof("Closing %s ...", t.uri)
	close(t.done)
	t.connection.Close()
	log.Infof("Closed %s", t.uri)
}

func (t *Client) Interrupt() {
	t.connection.conn.Close()
}

func (t *Client) ServiceClient(serviceName string) ServiceClient {
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	return &serviceClient{client: t, serviceName: serviceName + ".", isLogCalls: t.IsLogCalls}
}

func (t *Client) call(serviceMethod string, arg interface{}, rsp interface{}) error {
	return t.rpcClient.Call(serviceMethod, arg, rsp)
}

func (t *serviceClient) Call(arg interface{}, rsp interface{}) error {
	callerName := t.callerMethodName()
	name := t.serviceName + callerName
	if t.isLogCalls {
		log.Debugf("%s >", name)
		defer log.Debugf("%s <", name)
	}

	err := t.client.call(t.serviceName+callerName, arg, rsp)
	if err != nil && t.isLogCalls {
		log.Warnf("%s error: %v", name, err)
	}
	return err
}

func (*serviceClient) callerMethodName() string {
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
