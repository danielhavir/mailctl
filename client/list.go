package main

import (
	"bufio"
	"fmt"
	"net"
)

func list(config *Config) {
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
	conn.Write([]byte(config.User + "\n"))
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(message)
		fmt.Println(err)
		return
	}
	fmt.Print("Message from server: " + message)
}
