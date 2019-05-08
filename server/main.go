package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/user"
	"path"
)

// define path to files in a global scope
var storage string

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
