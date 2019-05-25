package main

import (
	"bufio"
	"crypto"
	"fmt"
	"net"

	"github.com/danielhavir/mailctl/internal/commons"
)

func list(config *Config, key []byte, prv crypto.PrivateKey) {
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
	r := bufio.NewReader(conn)
	// get the response
	status, err = r.ReadByte()
	if err != nil || status != 0 {
		fmt.Println("server did not recognize the operation ", err)
		return
	}

	userHash := commons.Hash([]byte(config.User + config.Organization))
	// write 32 bytes of user/org hash identifier
	conn.Write(userHash)
	// get the response
	status, err = r.ReadByte()
	if err != nil || status == 1 {
		fmt.Println("files could not be listed, check connection ", err)
		return
	}
	if status == 2 {
		fmt.Printf("username \"%s\" does not exist within organization \"%s\"\n", config.User, config.Organization)
		return
	}

	err = verify(r, conn, prv)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Listing files for \"%s@%s\":\n", config.User, config.Organization)
	file, err := r.ReadString('\n')
	for file != "EOF\n" && err == nil {
		fmt.Print(file)
		file, err = r.ReadString('\n')
		if err != nil {
			fmt.Println("error: ", err, file)
		}
	}
}
