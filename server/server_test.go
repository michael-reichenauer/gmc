package server

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"net/rpc"
	"reflect"
	"testing"
)

type Srv int
type Args struct {
	A, B int
}
type Rsp struct {
	R int
}

func (t *Srv) Add(args Args, rsp *Rsp) error {
	rsp.R = args.A + args.B
	return nil

}

func TestNewServer(t *testing.T) {
	s := new(Srv)
	rpcSrv := rpc.NewServer()
	err := rpcSrv.Register(s)
	assert.NoError(t, err)
	sname := reflect.Indirect(reflect.ValueOf(s)).Type().Name()
	t.Logf("%q", sname)

	l, err := net.Listen("tcp", "127.0.0.1:6821")
	assert.NoError(t, err)
	srv := &http.Server{Handler: rpcSrv}
	var rsp Rsp
	done := make(chan struct{})
	go func() {
		defer srv.Shutdown(context.Background())
		defer close(done)
		//	time.Sleep(1 * time.Second)
		client, err := rpc.DialHTTP("tcp", "127.0.0.1:6821")
		assert.NoError(t, err)
		args := &Args{A: 2, B: 5}

		assert.NoError(t, client.Call("Srv.Add", args, &rsp))
		//assert.Error(t, client.Call("Srv.Add", args, &rsp))
	}()

	assert.Equal(t, http.ErrServerClosed, srv.Serve(l))
	<-done
	assert.Equal(t, 7, rsp.R)
}
