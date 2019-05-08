package main

import (
	"bufio"
	"fmt"
	"net"

	"golang.org/x/crypto/blake2b"
)

func list(config *Config, key []byte) {
	var status byte

	// parse server IP from config file
	ip := net.ParseIP(config.Host)
	addr := net.TCPAddr{
		IP:   ip,
		Port: config.Port,
	}

	// dial a connection
	conn, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		panic(err)
	}

	// specify op
	conn.Write([]byte{'l'})
	// get the response
	status, err = bufio.NewReader(conn).ReadByte()
	if err != nil || status != 0 {
		fmt.Println("server did not recognize the operation ", err)
		return
	}

	h, err := blake2b.New256(key)
	if err != nil {
		fmt.Println(err)
		return
	}
	h.Write([]byte(config.User))
	h.Write([]byte(config.Organization))
	userHash := h.Sum(nil)
	// write 32 bytes of user/org hash identifier
	conn.Write(userHash)
	// get the response
	status, err = bufio.NewReader(conn).ReadByte()
	if err != nil || status == 1 {
		fmt.Println("files could not be listed, check connection ", err)
		return
	}
	if status == 2 {
		fmt.Printf("username \"%s\" does not exist within organization \"%s\"\n", config.User, config.Organization)
		return
	}

	fmt.Printf("Listing files for \"%s@%s\":\n", config.User, config.Organization)

	file, err := bufio.NewReader(conn).ReadString('\n')
	for file != "EOF\n" {
		fmt.Print(file)
	}
}
