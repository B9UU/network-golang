package unix

import (
	"context"
	"net"
	"os"
)

// streaming socket
func streamingEchoServer(ctx context.Context, network, addr string) (net.Addr, error) {
	// start the server
	s, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	// spin go rutine to handle the connections
	go func() {
		// new go rutine to block until it receives
		// a signal from ctx.Cancle. then close the server
		go func() {
			<-ctx.Done()
			_ = s.Close()
		}()

		// loop through the connections
		for {
			// block until new connection is established and accept
			conn, err := s.Accept()
			if err != nil {
				return
			}
			// spin new goroutine to handle the connection
			go func() {
				// defer connection closing
				defer func() { _ = conn.Close() }()
				// keep loping to handle multiple messages
				// from the same connection
				for {
					buf := make([]byte, 1024)
					// this will block until there's a new message
					// or the previous message was > 1024
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					// write to the buffer
					_, err = conn.Write(buf[:n])
					if err != nil {
						return
					}
				}
			}()
		}
	}()
	return s.Addr(), nil
}

// data gram socket
func datagramEchoServer(ctx context.Context, network, addr string) (net.Addr, error) {
	// create a server connectionless
	s, err := net.ListenPacket(network, addr)
	if err != nil {
		return nil, err
	}

	// Start goroutine to handle connections
	go func() {
		// goroutine for clean up
		go func() {
			// blocks until ctx.Close() sends a signal
			<-ctx.Done()
			// close the server
			_ = s.Close()
			// since we are using ListenPacket it will not remove the file for us
			if network == "unixgram" {
				_ = os.Remove(addr)
			}
		}()
		buf := make([]byte, 1024)
		// keep reading from the connection until EOF
		for {
			n, clientAddr, err := s.ReadFrom(buf)
			if err != nil {
				return
			}
			_, err = s.WriteTo(buf[:n], clientAddr)
			if err != nil {
				return
			}
		}
	}()
	return s.LocalAddr(), nil
}
