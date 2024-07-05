package main

import (
	"io"
	"net"
)

func proxyConn(source, destination string) error {
	connSource, err := net.Dial("tcp", source)
	if err != nil {
		return err
	}
	defer connSource.Close()
	connDestination, err := net.Dial("tcp", destination)
	if err != nil {
		return err
	}
	defer connDestination.Close()

	go func() {
		// - we don't have to handle any error since we are just a proxy
		// - io.Copy does not return EOF error. so probably we're only
		//	getting a error when either connection is closed
		_, _ = io.Copy(connSource, connDestination)
	}()
	_, err = io.Copy(connDestination, connSource)

	return err
}
