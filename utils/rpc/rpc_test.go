package rpc

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// The service interface implemented by both service client and server
type Service interface {
	Add(args Args, rsp *int) error
	Sub(args Args, rsp *int) error
}

// The service client
type ServiceClient struct {
	client *Client
}

func NewServiceClient(client *Client) Service {
	return &ServiceClient{client: client}
}

func (t ServiceClient) Add(args Args, rsp *int) error {
	// All implementations are the same, client will call different methods of the server
	return t.client.Call(args, rsp)
}

func (t ServiceClient) Sub(args Args, rsp *int) error {
	// All implementations are the same, client will call different methods of the server
	return t.client.Call(args, rsp)
}

type ServiceServer struct {
}

func NewServiceServer() Service {
	return &ServiceServer{}
}

type Args struct {
	A, B int
}

func (t *ServiceServer) Add(args Args, rsp *int) error {
	if args.A == 5 {
		// Handle special arg case to test error handling
		return fmt.Errorf("failed for 5")
	}
	*rsp = args.A + args.B
	return nil
}

func (t *ServiceServer) Sub(args Args, rsp *int) error {
	*rsp = args.A - args.B
	return nil
}

func TestNewJsonRpcServer(t *testing.T) {
	// Create rpc server and register service server
	serviceServer := NewServiceServer()
	rpcServer := NewRpcServer()
	err := rpcServer.RegisterName("Service", serviceServer)
	assert.NoError(t, err)

	// Start rpc sever and serve rpc requests
	assert.NoError(t, rpcServer.Start("http://127.0.0.1:0/rpc"))
	defer rpcServer.Close()
	go func() {
		err := rpcServer.Serve()
		if err != nil {
			panic(err)
		}
	}()

	// Create rpc client and create service client
	rpcClient := NewRpcClient("Service")
	assert.NoError(t, rpcClient.Connect(rpcServer.URL))
	defer rpcClient.Close()
	client := NewServiceClient(rpcClient)

	// Make rpc requests
	for i := 0; i < 1000; i++ {
		var rsp int
		if i == 5 {
			// Verify tha call for arg 5 will fail (server returns error)
			require.Error(t, client.Add(Args{A: i, B: i}, &rsp), "Add Call: %d", i)
			continue
		}
		require.NoError(t, client.Add(Args{A: i, B: i}, &rsp), "Add Call: %d", i)
		require.Equal(t, i*2, rsp)
		require.NoError(t, client.Sub(Args{A: i * 2, B: i}, &rsp), "Sub Call: %d", i)
		require.Equal(t, i, rsp)
	}
}

func TestClosingServer(t *testing.T) {
	// Create rpc server and register service server
	serviceServer := NewServiceServer()
	rpcServer := NewRpcServer()
	err := rpcServer.RegisterName("Service", serviceServer)
	assert.NoError(t, err)

	// Start rpc sever and serve rpc requests
	assert.NoError(t, rpcServer.Start("http://127.0.0.1:0/rpc"))
	defer rpcServer.Close()
	go func() {
		err := rpcServer.Serve()
		if err != nil {
			panic(err)
		}
	}()

	// Create rpc client and create service client
	rpcClient := NewRpcClient("Service")
	assert.NoError(t, rpcClient.Connect(rpcServer.URL))
	defer rpcClient.Close()
	client := NewServiceClient(rpcClient)

	// Make rpc requests
	for i := 0; i < 1000; i++ {
		var rsp int
		if i == 5 {
			// Verify tha call for arg 5 will fail (server returns error)
			require.Error(t, client.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
			continue
		}
		if i == 20 {
			rpcServer.Close()
			// Verify that calls will fail after server is closed
			require.Error(t, client.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
			break
		}
		require.NoError(t, client.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
		require.Equal(t, i*2, rsp)
	}
}

func TestClosingClient(t *testing.T) {
	// Create rpc server and register service server
	serviceServer := NewServiceServer()
	rpcServer := NewRpcServer()
	err := rpcServer.RegisterName("Service", serviceServer)
	assert.NoError(t, err)

	// Start rpc sever and serve rpc requests
	assert.NoError(t, rpcServer.Start("http://127.0.0.1:0/rpc"))
	defer rpcServer.Close()
	go func() {
		err := rpcServer.Serve()
		if err != nil {
			panic(err)
		}
	}()

	// Create rpc client and create service client
	rpcClient := NewRpcClient("Service")
	assert.NoError(t, rpcClient.Connect(rpcServer.URL))
	defer rpcClient.Close()
	client := NewServiceClient(rpcClient)

	// Make rpc requests
	for i := 0; i < 1000; i++ {
		var rsp int
		if i == 5 {
			// Verify tha call for arg 5 will fail (server returns error)
			require.Error(t, client.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
			continue
		}
		if i == 20 {
			rpcClient.Close()
			// Verify that calls will fail after client is closed
			require.Error(t, client.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
			break
		}
		require.NoError(t, client.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
		require.Equal(t, i*2, rsp)
	}
}
