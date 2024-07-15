//go:build darwin || linux
// +build darwin linux

package unix

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestEchoServerUnix(t *testing.T) {
	// make echo_file in /tmp/ directory
	dir, err := os.MkdirTemp("", "echo_unix")
	if err != nil {
		t.Fatal(err)
	}
	// defering the cleaning
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()
	// create a context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// join the path
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	// call the streamingEchoServer
	rAddr, err := streamingEchoServer(ctx, "unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	// give read/write permision
	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("unix", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	msg := []byte("ping")
	// write 3 messages
	for i := 0; i < 3; i++ {
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf) // read once from the server
	if err != nil {
		t.Fatal(err)
	}
	expected := bytes.Repeat(msg, 3)
	if !bytes.Equal(expected, buf[:n]) {
		t.Fatalf("expected reply %q; actual reply %q", expected, buf[:n])
	}
	// NOTE: learn what is this
	// _ = closer.Close()
	// <-done
}

func TestEchoServerUnixDatagram(t *testing.T) {
	// initiate the server and create the socket file for the server
	// the server will handle the clean up
	dir, err := os.MkdirTemp("", "echo_unixgram")
	if err != nil {
		t.Fatal(err)
	}

	// we call os.RemoveAll() to remove the socket temp directory with all subdirectories
	// that means the clients and server socket files. otherwise you have to do it manually
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()
	// create the context with cancel and join the path and call the datagramEchoServer()
	ctx, cancel := context.WithCancel(context.Background())
	sSocket := filepath.Join(dir, fmt.Sprintf("s%d.sock", os.Getpid()))
	serverAddr, err := datagramEchoServer(ctx, "unixgram", sSocket)
	if err != nil {
		t.Fatal(err)
	}
	// defer cancelling
	defer cancel()
	// give read/write permision
	err = os.Chmod(sSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}
	cSocket := filepath.Join(dir, fmt.Sprintf("c%d.sock", os.Getpid()))
	client, err := net.ListenPacket("unixgram", cSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	err = os.Chmod(cSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("ping")
	for i := 0; i < 3; i++ { // writing 3 messages
		_, err = client.WriteTo(msg, serverAddr)
		if err != nil {
			t.Fatal(err)
		}
	}
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ { // read 3 messages
		n, addr, err := client.ReadFrom(buf)
		if err != nil {
			t.Fatal(err)
		}
		if addr.String() != serverAddr.String() {
			t.Fatalf("received reply from %q instead of %q", addr, serverAddr)
		}
		if !bytes.Equal(msg, buf[:n]) {
			t.Fatalf("expected reply %q; actual reply %q", msg, buf[:n])
		}
	}

}
