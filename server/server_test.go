package server

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"
)

type Service interface {
	Add(args Args, rsp *Rsp) error
}

type ServiceClient struct {
	address string
	client  *rpc.Client
}

func NewServiceClient(address string) Service {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}
	//defer conn.Close()

	client := jsonrpc.NewClient(conn)

	return &ServiceClient{address: address, client: client}
}

func (t ServiceClient) Add(args Args, rsp *Rsp) error {
	return t.client.Call("Srv.Add", args, rsp)
}

type ServiceServer struct {
}

func NewServiceServer() Service {
	return &ServiceServer{}
}

type Args struct {
	A, B int
}
type Rsp struct {
	R int
}

func (t *ServiceServer) Add(args Args, rsp *Rsp) error {
	rsp.R = args.A + args.B
	return nil
}

func TestNewJsonRpcServer(t *testing.T) {
	address := "127.0.0.1:0"
	service := NewServiceServer()

	rpcSrv := rpc.NewServer()
	err := rpcSrv.RegisterName("Srv", service)
	assert.NoError(t, err)

	l, err := net.Listen("tcp", address)
	assert.NoError(t, err)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				panic(err)
			}
			// t.Logf("Recevied client connection  %s->%s", conn.RemoteAddr(), conn.LocalAddr())
			go rpcSrv.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	client := NewServiceClient(l.Addr().String())

	for i := 0; i < 1000; i++ {
		var rsp Rsp
		require.NoError(t, client.Add(Args{A: i, B: i}, &rsp))
		require.Equal(t, i*2, rsp.R)
	}
}

// func TestNewRpcServer(t *testing.T) {
// 	service := new(Srv)
// 	rpcSrv := rpc.NewServer()
// 	err := rpcSrv.Register(service)
// 	assert.NoError(t, err)
// 	serviceName := reflect.Indirect(reflect.ValueOf(service)).Type().Name()
// 	t.Logf("%q", serviceName)
//
// 	l, err := net.Listen("tcp", "127.0.0.1:0")
// 	assert.NoError(t, err)
// 	srv := &http.Server{Handler: rpcSrv}
// 	var rsp Rsp
// 	done := make(chan struct{})
// 	go func() {
// 		defer srv.Shutdown(context.Background())
// 		defer close(done)
// 		//	time.Sleep(1 * time.Second)
// 		client, err := rpc.DialHTTP("tcp", l.Addr().String())
// 		assert.NoError(t, err)
// 		args := &Args{A: 2, B: 5}
//
// 		assert.NoError(t, client.Call("Srv.Add", args, &rsp))
// 		//assert.Error(t, client.Call("Srv.Add", args, &rsp))
// 	}()
//
// 	assert.Equal(t, http.ErrServerClosed, srv.Serve(l))
// 	<-done
// 	assert.Equal(t, 7, rsp.R)
// }

// func GetFunctionName(i interface{}) string {
// 	v := reflect.ValueOf(i)
// 	p := v.Pointer()
// 	r := runtime.FuncForPC(p)
// 	return r.Name()
// }
