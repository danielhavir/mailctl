package main

import (
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
}
