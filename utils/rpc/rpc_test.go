package rpc_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/michael-reichenauer/gmc/utils/log"
	"github.com/michael-reichenauer/gmc/utils/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type None *int

var (
	Nil None = &no
	no  int
)

// The service interface implemented by both service client and server
type Api interface {
	Add(arg Args, rsp *int) error
	Sub(arg Args, rsp *int) error
	Set(arg Args, rsp None) error
	Get(arg None, rsp *int) error
	Trigger(arg None, rsp None) error
	OpenEvents(arg None, rsp *string) error
	GenerateEvents(count int, rsp None) error
	CloseEvents(arg string, rsp None) error
}

type Event struct {
	Name  string
	Count int
}

// The service client
type ApiClient struct {
	client rpc.ServiceClient
}

func NewApiClient(client rpc.ServiceClient) *ApiClient {
	return &ApiClient{client: client}
}

func (t *ApiClient) Events(id string) (chan Event, error) {
	rawEvents, err := t.client.Events(id)
	if err != nil {
		return nil, err
	}
	events := make(chan Event)

	go func() {
		defer close(events)
		for e := range rawEvents {
			if e == "" {
				break
			}
			var event Event
			err := json.Unmarshal([]byte(e), &event)
			if err != nil {
				panic(log.Fatal(err))
			}
			events <- event
		}
	}()

	return events, nil
}

func (t *ApiClient) Add(args Args, rsp *int) error {
	// All implementations are the same, client will call different methods of the server
	return t.client.Call(args, rsp)
}

func (t *ApiClient) Sub(args Args, rsp *int) error {
	// All implementations are the same, client will call different methods of the server
	return t.client.Call(args, rsp)
}

func (t *ApiClient) Set(args Args, rsp None) error {
	return t.client.Call(args, rsp)
}

func (t *ApiClient) Get(args None, rsp *int) error {
	return t.client.Call(args, rsp)
}

func (t *ApiClient) Trigger(args None, rsp None) error {
	return t.client.Call(args, rsp)
}

func (t *ApiClient) OpenEvents(args None, rsp *string) error {
	return t.client.Call(args, rsp)
}

func (t *ApiClient) GenerateEvents(arg int, rsp None) error {
	return t.client.Call(arg, rsp)
}

func (t *ApiClient) CloseEvents(arg string, rsp None) error {
	return t.client.Call(arg, rsp)
}

type ApiServer struct {
	rpcServer *rpc.Server
	eventsID  string
}

func NewApiServer(rpcServer *rpc.Server) *ApiServer {
	return &ApiServer{rpcServer: rpcServer}
}

type Args struct {
	A, B int
}

func (t *ApiServer) Add(args Args, rsp *int) error {
	if args.A == 5 {
		// Handle special arg case to test error handling
		return fmt.Errorf("failed for 5")
	}
	*rsp = args.A + args.B
	return nil
}

func (t *ApiServer) Sub(args Args, rsp *int) error {
	*rsp = args.A - args.B
	return nil
}

func (t *ApiServer) Set(arg Args, rsp None) error {
	return nil
}

func (t *ApiServer) Get(arg None, rsp *int) error {
	*rsp = 5
	return nil
}

func (t *ApiServer) Trigger(_ None, _ None) error {
	return nil
}

func (t *ApiServer) OpenEvents(_ None, rsp *string) error {
	t.eventsID = uuid.New().String()
	t.rpcServer.CreateEvents(t.eventsID)
	*rsp = t.eventsID
	return nil
}

func (t *ApiServer) GenerateEvents(args int, rsp None) error {
	go func() {
		for i := 0; i < args; i++ {
			t.rpcServer.PostEvent(t.eventsID, Event{Name: fmt.Sprintf("%d", i), Count: i})
		}
	}()
	return nil
}

func (t *ApiServer) CloseEvents(arg string, _ None) error {
	t.rpcServer.CloseEvents(arg)
	return nil
}

func TestRpc(t *testing.T) {
	// Create rpc server and register service server
	rpcServer := rpc.NewServer()
	apiServer := NewApiServer(rpcServer)
	err := rpcServer.RegisterService("api", apiServer)
	assert.NoError(t, err)

	// Start rpc sever and serve rpc requests
	assert.NoError(t, rpcServer.Start("http://127.0.0.1:0/api/ws", "/api/events"))
	defer rpcServer.Close()
	go func() {
		err := rpcServer.Serve()
		if err != nil {
			panic(err)
		}
	}()

	// Create rpc client and create service client
	rpcClient := rpc.NewClient()
	assert.NoError(t, rpcClient.Connect(rpcServer.RPCURL, rpcServer.EventsURL))
	defer rpcClient.Close()
	apiClient := NewApiClient(rpcClient.ServiceClient("api"))

	// Make rpc requests
	var rsp int
	for i := 0; i < 1000; i++ {
		if i == 5 {
			// Verify tha call for arg 5 will fail (server returns error)
			require.Error(t, apiClient.Add(Args{A: i, B: i}, &rsp), "Add Call: %d", i)
			continue
		}
		require.NoError(t, apiClient.Add(Args{A: i, B: i}, &rsp), "Add Call: %d", i)
		require.Equal(t, i*2, rsp)
		require.NoError(t, apiClient.Sub(Args{A: i * 2, B: i}, &rsp), "Sub Call: %d", i)
		require.Equal(t, i, rsp)
	}

	require.NoError(t, apiClient.Set(Args{A: 5, B: 3}, Nil))
	require.NoError(t, apiClient.Get(Nil, &rsp))
	require.Equal(t, 5, rsp)
	require.NoError(t, apiClient.Trigger(Nil, Nil))

	var eventsID string
	require.NoError(t, apiClient.OpenEvents(Nil, &eventsID))
	events, err := apiClient.Events(eventsID)
	require.NoError(t, err)

	apiClient.GenerateEvents(5, Nil)
	count := 0
	for e := range events {
		require.Equal(t, fmt.Sprintf("%d", count), e.Name)
		require.Equal(t, count, e.Count)
		count++
		if count == 5 {
			require.NoError(t, apiClient.CloseEvents(eventsID, Nil))
		}
	}
	require.Equal(t, 5, count)
}

func TestRpcWithCloseServer(t *testing.T) {
	// Create rpc server and register service server
	rpcServer := rpc.NewServer()
	err := rpcServer.RegisterService("api", NewApiServer(rpcServer))
	assert.NoError(t, err)

	// Start rpc sever and serve rpc requests
	assert.NoError(t, rpcServer.Start("http://127.0.0.1:0/api/ws", "/api/events"))
	defer rpcServer.Close()
	go func() {
		err := rpcServer.Serve()
		if err != nil {
			panic(err)
		}
	}()

	// Create rpc client and create service client
	rpcClient := rpc.NewClient()
	rpcClient.OnConnectionError = func(err error) { log.Warnf("Connection error: %v", err) }
	assert.NoError(t, rpcClient.Connect(rpcServer.RPCURL, rpcServer.EventsURL))
	defer rpcClient.Close()
	apiClient := NewApiClient(rpcClient.ServiceClient("api"))

	// Make rpc requests
	for i := 0; i < 1000; i++ {
		var rsp int
		if i == 5 {
			// Verify tha call for arg 5 will fail (server returns error)
			require.Error(t, apiClient.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
			continue
		}
		if i == 20 {
			rpcServer.Close()
			// Verify that calls will fail after server is closed
			require.Error(t, apiClient.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
			break
		}
		require.NoError(t, apiClient.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
		require.Equal(t, i*2, rsp)
	}
}

func TestRpcWithCloseClient(t *testing.T) {
	// Create rpc server and register service server
	rpcServer := rpc.NewServer()
	err := rpcServer.RegisterService("api", NewApiServer(rpcServer))
	assert.NoError(t, err)

	// Start rpc sever and serve rpc requests
	assert.NoError(t, rpcServer.Start("http://127.0.0.1:0/api/ws", "/api/events"))
	defer rpcServer.Close()
	go func() {
		err := rpcServer.Serve()
		if err != nil {
			panic(err)
		}
	}()

	// Create rpc client and create service client
	rpcClient := rpc.NewClient()
	rpcClient.OnConnectionError = func(err error) { log.Warnf("Connection error: %v", err) }
	assert.NoError(t, rpcClient.Connect(rpcServer.RPCURL, rpcServer.EventsURL))
	defer rpcClient.Close()
	apiClient := NewApiClient(rpcClient.ServiceClient("api"))

	// Make rpc requests
	for i := 0; i < 1000; i++ {
		var rsp int
		if i == 5 {
			// Verify tha call for arg 5 will fail (server returns error)
			require.Error(t, apiClient.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
			continue
		}
		if i == 20 {
			rpcClient.Interrupt()
			// Verify that calls will fail after client is closed
			require.Error(t, apiClient.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
			break
		}
		require.NoError(t, apiClient.Add(Args{A: i, B: i}, &rsp), "Call: %d", i)
		require.Equal(t, i*2, rsp)
	}
}
