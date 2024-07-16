package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"network-golang/creds/auth"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
)

func init() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n\t%s <groups name>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

// look if the group passed existes and app the groupid as a key into the map
func parseGroupNames(args []string) map[string]struct{} {
	groups := make(map[string]struct{})
	for _, arg := range args {
		grp, err := user.LookupGroup(arg)
		if err != nil {
			log.Println(err)
			continue
		}
		groups[grp.Gid] = struct{}{}
	}
	return groups
}

func main() {
	flag.Parse()
	// groups allowed to access this
	groups := parseGroupNames(flag.Args())
	// create tempdir with creds.sock
	socket := filepath.Join(os.TempDir(), "creds.sock")
	// make the path unix endpoint
	addr, err := net.ResolveUnixAddr("unix", socket)
	if err != nil {
		log.Fatal(err)
	}
	// start the server
	s, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.Fatal(err)
	}
	// make a channel that listens to the sys/os interrupts
	c := make(chan os.Signal, 1)
	// creates a goroutine to watch for signals and pass the signal we want into the channel
	signal.Notify(c, os.Interrupt)
	go func() {
		// blocks until we recieve a signal
		<-c
		_ = s.Close()
	}()
	fmt.Printf("Listening on %s ...\n", socket)

	for {
		// blocks until we a get a connection
		conn, err := s.AcceptUnix()
		if err != nil {
			break
		}
		// checks if the conn is authenticated
		if auth.Allowed(conn, groups) {
			// success you're allowed
			_, err = conn.Write([]byte("Welcome\n"))
			if err != nil {
				// handle the connection in a goroutine here
				continue
			}
		} else {
			// denied, you're disallowed
			_, err = conn.Write([]byte("Access denied\n"))
		}
		if err != nil {
			log.Println(err)
		}
		_ = conn.Close()
	}
}
