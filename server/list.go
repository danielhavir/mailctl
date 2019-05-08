package main

import (
	"bufio"
	"net"
)

func listFiles(r *bufio.Reader, conn net.Conn) {
	user, _ := r.ReadString('\n')
	conn.Write([]byte("listing unread files for user " + user))
	return
}
