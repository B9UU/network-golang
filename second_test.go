package main

import (
	"net"
	"syscall"
	"testing"
	"time"
)

func DialTimeout(timeout time.Duration, network, address string) (net.Conn, error) {
	d := net.Dialer{
		Control: func(_, addr string, c syscall.RawConn) error {
			return &net.DNSError{
				Err:         "connection timed out",
				Name:        addr,
				Server:      "127.0.0.1",
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
		Timeout: timeout,
	}
	return d.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
	c, err := DialTimeout(5*time.Second, "tcp", "10.0.0.0:http")
	if err == nil {
		c.Close()
		t.Fatal("connection did not time out")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err)
	}
	if !nErr.Timeout() {
		t.Fatal("error is not a timeout")
	}
}
