package rpc

import (
	"net"
	"sync/atomic"
	"time"
)

type connection struct {
	conn              net.Conn
	onConnectionError func(err error)
	connErrors        int32
	latency           time.Duration
	bandWithMpbs      float32
	onRead            func(p []byte)
	onWrite           func(p []byte)
}

func (t *connection) Read(p []byte) (n int, err error) {
	n, err = t.conn.Read(p)
	if err != nil && t.onConnectionError != nil {
		if c := atomic.AddInt32(&t.connErrors, 1); c == 1 {
			t.onConnectionError(err)
		}
	}
	if t.latency != 0 {
		time.Sleep(t.latency)
	}
	if t.onRead != nil {
		t.onRead(p)
	}
	return
}

func (t *connection) Write(p []byte) (n int, err error) {
	if t.latency != 0 {
		time.Sleep(t.latency)
	}
	n, err = t.conn.Write(p)
	if err != nil && t.onConnectionError != nil {
		if c := atomic.AddInt32(&t.connErrors, 1); c == 1 {
			t.onConnectionError(err)
		}
	}
	if t.onWrite != nil {
		t.onWrite(p)
	}
	return
}

func (t *connection) Close() error {
	return t.conn.Close()
}
