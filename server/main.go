package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/user"
	"path"
	"time"
)

// define path to files in a global scope
var storage string

func handleConnection(conn net.Conn) {
	start := time.Now()
	defer conn.Close()
	r := bufio.NewReader(conn)

	op, err := r.ReadByte()
	if err != nil {
		fmt.Println("Empty connection: ", err)
		return
	}

	switch op {
	case 'l':
		conn.Write([]byte{0})
		listFiles(r, conn)
	case 'c':
		conn.Write([]byte{0})
		registerKey(r, conn)
	case 'r':
		conn.Write([]byte{0})
		sendToClient(r, conn)
	case 's':
		conn.Write([]byte{0})
		recvFromClient(r, conn)
	default:
		conn.Write([]byte{1})
	}
	elapsed := time.Since(start)
	fmt.Printf("New connection with %s resolved in %s\n", conn.RemoteAddr().String(), elapsed)
}

func main() {
	usr, err := user.Current()
	if err != nil {
		return
	}
	storage = path.Join(usr.HomeDir, confDir)

	flag.StringVar(&storage, "storage-path", storage, "Path to the directory with the config file [optional] (default \"~/.mailctl\")")
	address := flag.String("address", ":1881", "address to host on (default \":1881\")")
	flag.Parse()

	configured, err := configure(storage)
	if err != nil {
		fmt.Println("couldn't configure organization", err)
		os.Exit(1)
	}
	if !configured {
		fmt.Println("configured organization storage at", storage)
	}

	fmt.Println("Launching server...")

	ln, err := net.Listen("tcp", *address)
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

		// handle connection in parallel process
		go handleConnection(conn)
	}
}
