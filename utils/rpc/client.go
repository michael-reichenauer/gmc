package rpc

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/michael-reichenauer/gmc/utils"
	"github.com/michael-reichenauer/gmc/utils/log"
	"golang.org/x/net/websocket"
)

// ServiceClient supports making remote calls
type ServiceClient interface {
	Call(arg interface{}, rsp interface{}) error
	Events(url string) (chan string, error)
}

type serviceClient struct {
	serviceName string
	client      *Client
	isLogCalls  bool
}

// Client supports rpc client functionality
type Client struct {
	IsLogCalls        bool
	OnConnectionError func(err error)
	connection        *connection
	rpcClient         *rpc.Client
	done              chan struct{}
	uri               string
	eventsURI         string
	host              string
	Latency           time.Duration
	BandWithMbit      float32
}

// NewClient returns a new rpc client
func NewClient() *Client {
	return &Client{done: make(chan struct{})}
}

// Connect to the specified uri
func (t *Client) Connect(uri, eventsURI string) error {
	// if t.Latency != 0 {
	// 	time.Sleep(3 * t.Latency)
	// }
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	t.host = u.Host
	origin := fmt.Sprintf("http://%s", u.Host)

	conn, err := websocket.Dial(uri, "", origin)
	if err != nil {
		return err
	}
	t.uri = uri
	t.eventsURI = eventsURI
	log.Infof("Connected to  %q", uri)
	t.connection = &connection{
		conn:              conn,
		onConnectionError: t.OnConnectionError,
		latency:           t.Latency,
		bandWithMbit:      t.BandWithMbit,
	}
	t.rpcClient = jsonrpc.NewClient(t.connection)
	return nil
}

// ConnectEvents connect to the event channel
func (t *Client) events(id string) (chan string, error) {
	url := fmt.Sprintf("%s%s", t.eventsURI, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(log.Fatal(err))
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")

	httpClient := &http.Client{Transport: &http.Transport{Proxy: utils.GetHTTPProxy()}}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Infof("No contact with %s, %s", url, err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		message, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Error code: %d, %s", resp.StatusCode, message)
	}
	eventChan := make(chan string)

	go func() {
		defer func() { _ = resp.Body.Close() }()
		defer func() { eventChan <- "" }()

		r := bufio.NewReader(resp.Body)
		for {
			e, err := t.readEvent(r)
			if err == io.EOF {
				log.Infof("Event stream closed")
				return
			}
			if err != nil {
				log.Warnf("Error %v", err)
				return
			}
			eventChan <- e
			//log.Infof("Event %q", e)
		}
	}()

	return eventChan, nil
}

func (t *Client) readEvent(r *bufio.Reader) (string, error) {
	_, err := r.Peek(1)
	if err == io.ErrUnexpectedEOF {
		err = io.EOF
	}
	if err != nil {
		return "", err
	}

	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Skip last '\n' of the double '\n'
	_, err = r.ReadString('\n')
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(line, "data: ") {
		return "", fmt.Errorf("invalid event format")

	}
	line = line[6:]
	line = strings.TrimSuffix(line, "\n")

	return line, nil
}

// Close closes the connection
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

// Interrupt the connection (used in tests)
func (t *Client) Interrupt() {
	t.connection.conn.Close()
}

// ServiceClient returns service client, which supports making remote calls
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

func (t *serviceClient) Events(url string) (chan string, error) {
	return t.client.events(url)
}

func (*serviceClient) callerMethodName() string {
	pc := make([]uintptr, 1)
	n := runtime.Callers(3, pc[:])
	if n < 1 {
		return ""
	}
	frame, _ := runtime.CallersFrames(pc).Next()
	callerName := frame.Function
	i := strings.LastIndex(callerName, ".")
	if i != -1 {
		callerName = callerName[i+1:]
	}
	return callerName
}
