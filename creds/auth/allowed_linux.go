package auth

import (
	"fmt"
	"log"
	"net"
	"os/user"
	"strconv"

	"golang.org/x/sys/unix"
)

// check if the connection user in the groups -- authenticatedk
func Allowed(conn *net.UnixConn, groups map[string]struct{}) bool {
	// check for connection and groups and len of groups
	if conn == nil || groups == nil || len(groups) == 0 {
		return false
	}

	// getting the connection file -- *.sock
	file, _ := conn.File()
	// defer file closing
	defer func() { _ = file.Close() }()
	var (
		err   error
		ucred *unix.Ucred
	)
	for {
		// getting the user credentials
		ucred, err = unix.GetsockoptUcred(int(file.Fd()), // file descriptor is a unique refrence that the os assigns to the file
			unix.SOL_SOCKET, unix.SO_PEERCRED)
		// syscall interrupted, try again
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			log.Println(err)
			return false
		}
		break
	}
	// check if the id we have is an actual user and get the user
	u, err := user.LookupId(strconv.Itoa(int(ucred.Uid)))
	if err != nil {
		log.Println(err)
		return false
	}
	// get all groupids for the user
	gids, err := u.GroupIds()
	if err != nil {
		log.Println(err)
		return false
	}
	// then create a map of all grpids the user in
	for _, gid := range gids {
		if _, ok := groups[gid]; ok {
			return true
		}
	}
	return false
}
