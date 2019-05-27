package main

import (
	"bufio"
	"crypto"
	"fmt"
	"io/ioutil"
	"net"

	hpke "github.com/danielhavir/go-hpke"
	"github.com/danielhavir/mailctl/internal/commons"
)

func recv(config *Config, messageID string, key []byte, prv crypto.PrivateKey) {
	var status byte
	params, _ := hpke.GetParams(commons.HpkeMode)
	userHash := commons.Hash([]byte(config.getUserOrg()))

	// dial a connection
	conn, err := net.DialTCP("tcp", nil, config.parseIP())
	if err != nil {
		panic(err)
	}

	// specify op
	conn.Write([]byte{'r'})
	r := bufio.NewReader(conn)
	// get the response
	status, err = r.ReadByte()
	if err != nil || status != 0 {
		fmt.Println("server did not recognize the operation ", err)
		return
	}

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

	fmt.Println("user found and authenticated")

	// write file name length
	conn.Write([]byte{uint8(len(messageID))})
	// write message id (i.e. filename)
	conn.Write([]byte(messageID))

	// get the response
	status, err = r.ReadByte()
	if err != nil || status == 1 {
		fmt.Println("files could not be listed, check connection ", err)
		return
	}
	if status == 2 {
		fmt.Printf("filename \"%s\" does not exist\n", messageID)
		return
	}

	msgLenBytes := make([]byte, 4)
	_, err = r.Read(msgLenBytes)
	if err != nil {
		fmt.Println("problem with connection", err)
		return
	}
	digest := make([]byte, commons.ByteToUint32(msgLenBytes))
	_, err = r.Read(digest)
	if err != nil {
		conn.Write([]byte{1})
		fmt.Println("problem when receiving file", err)
		return
	}
	enc := make([]byte, params.PubKeySize)
	copy(enc, digest[:params.PubKeySize])
	ct := make([]byte, len(digest)-params.PubKeySize)
	copy(ct, digest[params.PubKeySize:])

	msg, err := hpke.DecryptBase(params, prv, enc, ct, nil)
	if err != nil {
		conn.Write([]byte{1})
		fmt.Println("decryption failed", err)
		return
	}

	err = ioutil.WriteFile(messageID, msg, 0644)
	if err != nil {
		conn.Write([]byte{1})
		fmt.Println("file could not be saved", err)
		return
	}
	conn.Write([]byte{0})
	fmt.Printf("file %s successfully received\n", messageID)

}
