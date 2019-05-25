package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	hpke "github.com/danielhavir/go-hpke"
	utils "github.com/danielhavir/mailctl/internal/utils"
)

func send(config *Config, rcpt string, filepath string, subject string) {
	msg, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Printf("filepath \"%s\" does not exist\n", err)
		return
	}
	fmt.Printf("sending %s to %s\n", filepath, rcpt)

	userOrg := strings.Split(rcpt, "@")
	var status byte
	params, _ := hpke.GetParams(hpkeMode)

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
	conn.Write([]byte{'s'})
	r := bufio.NewReader(conn)
	// get the response
	status, err = r.ReadByte()
	if err != nil || status != 0 {
		fmt.Println("server did not recognize the operation ", err)
		return
	}

	userHash := utils.Hash([]byte(userOrg[0] + userOrg[1]))
	// write 32 bytes of user/org hash identifier
	conn.Write(userHash)
	// get the response
	status, err = r.ReadByte()
	if err != nil || status == 1 {
		fmt.Println("files could not be listed, check connection ", err)
		return
	}
	if status == 2 {
		fmt.Printf("username \"%s\" does not exist within organization \"%s\" or his public key was not registered\n", userOrg[0], userOrg[1])
		return
	}

	pBytes := make([]byte, params.PubKeySize)
	_, err = r.Read(pBytes)
	if err != nil || status == 1 {
		fmt.Println("public key could not be received ", err)
		return
	}
	pub, err := hpke.Unmarshall(params, pBytes)
	if err != nil {
		fmt.Println("incorrect public key received", err)
		return
	}

	conn.Write([]byte(subject + "\n"))

	ct, enc, err := hpke.EncryptBase(params, nil, pub, msg, nil)
	if err != nil {
		fmt.Println("message could not be encrypted", err)
		return
	}

	conn.Write(utils.Uint32ToByte(uint32(len(ct) + params.PubKeySize)))
	conn.Write(append(enc, ct...))

	// get the response
	status, err = r.ReadByte()
	if err != nil || status != 0 {
		fmt.Println("server did not receive the message ", err)
		return
	}

	fmt.Printf("file %s sent\n", filepath)

}
