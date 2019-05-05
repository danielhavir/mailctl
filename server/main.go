package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func listFiles(r *bufio.Reader, conn net.Conn) {
	user, _ := r.ReadString('\n')
	conn.Write([]byte("listing unread files for user " + user))
	return
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	fmt.Println("New connection with " + conn.RemoteAddr().String())

	op, err := r.ReadByte()
	if err != nil {
		fmt.Println("Empty connection: ", err)
		return
	}

	switch op {
	case 'l':
		listFiles(r, conn)
		break
	default:
		conn.Write([]byte("error"))
	}
}

func main() {
	fmt.Println("Launching server...")

	ln, err := net.Listen("tcp", ":1881")
	if err != nil {
		fmt.Println("Error launching server: ", err)
		os.Exit(1)
	}

	defer ln.Close()

	for {
		// accept connection on port
		conn, err := ln.Accept()

		if err != nil {
			fmt.Println("Error with connection: ", err)
			continue
		}

		go handleConnection(conn)
	}
}
