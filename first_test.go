package main

import (
	"io"
	"net"
	"testing"
)

func TestListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}
			go func(c net.Conn) {
				t.Log("I got a connection")
				defer func() {
					c.Close()
					done <- struct{}{}
				}()
				buf := make([]byte, 1024)
				n, err := c.Read(buf)
				if err != nil {
					if err != io.EOF {
						t.Log(err)
					}
					return
				}
				t.Logf("received: %q", buf[:n])
			}(conn)
		}
	}()

	randomAddress := listener.Addr().String()
	conn, err := net.Dial("tcp", randomAddress)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("I sent a connection on ", randomAddress)
	conn.Close()
	<-done
	listener.Close()
	<-done

}
