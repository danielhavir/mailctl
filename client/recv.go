package main

import (
	"bufio"
	"fmt"
	"net"
)

func recv(config *Config, messageID string) {
	ip := net.ParseIP(config.Host)
	addr := net.TCPAddr{
		IP:   ip,
		Port: config.Port,
	}
	fmt.Printf("Downloading message %s for %s from %s\n", messageID, config.User, addr.String())
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		panic(err)
	}

	conn.Write([]byte{'l'})
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print("Message from server: " + message)
}
