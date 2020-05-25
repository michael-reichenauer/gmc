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
)

type ServiceClient interface {
	Call(arg interface{}, rsp interface{}) error
}

type serviceClient struct {
	serviceName string
	client      *Client
}

type Client struct {
	conn      net.Conn
	rpcClient *rpc.Client
	done      chan struct{}
	uri       string
}

func NewClient() *Client {
	return &Client{done: make(chan struct{})}
}

func (t *Client) Connect(uri string) error {
	log.Infof("Connecting to %s", uri)
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	t.uri = uri

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return err
	}
	t.conn = conn
	io.WriteString(conn, "CONNECT "+u.Path+" HTTP/1.0\n\n")

	// Require successful HTTP response before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.StatusCode == http.StatusOK {
		// OK response, create json rpc client
		log.Infof("Connected %s->%s", conn.LocalAddr(), uri)
		t.rpcClient = jsonrpc.NewClient(conn)
		return nil
	}

	// Some error occurred, either connect error or response error
	if err == nil {
		// Some response error
		err = fmt.Errorf("invalid http response, code: %d, %s", resp.StatusCode, resp.Status)
	}
	log.Infof("Failed to connect to %s, %v", uri, err)
	t.Close()
	return err
}

func (t *Client) Close() {
	select {
	case <-t.done:
		// Already closed
		return
	default:
	}
	log.Infof("Closing %s", t.uri)
	close(t.done)

	t.conn.Close()
}

func (t *Client) ServiceClient(serviceName string) ServiceClient {
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	return &serviceClient{client: t, serviceName: serviceName + "."}
}

func (t *Client) call(serviceMethod string, arg interface{}, rsp interface{}) error {
	return t.rpcClient.Call(serviceMethod, arg, rsp)
}

func (t *serviceClient) Call(arg interface{}, rsp interface{}) error {
	callerName := t.callerMethodName()
	name := t.serviceName + callerName
	log.Debugf("%s >", name)
	defer log.Debugf("%s <", name)
	err := t.client.call(t.serviceName+callerName, arg, rsp)
	if err != nil {
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
