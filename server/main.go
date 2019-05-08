package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"golang.org/x/crypto/blake2b"
)

func listFiles(r *bufio.Reader, conn net.Conn) {
	user, _ := r.ReadString('\n')
	conn.Write([]byte("listing unread files for user " + user))
	return
}

func registerKey(r *bufio.Reader, conn net.Conn) {
	userHash := make([]byte, blake2b.Size256)
	if _, err := r.Read(userHash); err != nil {
		conn.Write([]byte{1})
		return
	}

	pubBytes := make([]byte, 64)
	if _, err := r.Read(pubBytes); err != nil {
		conn.Write([]byte{1})
		return
	}
	// save key to appropriate directory

	conn.Write([]byte{0})
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
		conn.Write([]byte{0})
		listFiles(r, conn)
		break
	case 'c':
		conn.Write([]byte{0})
		registerKey(r, conn)
		break
	default:
		conn.Write([]byte{1})
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
