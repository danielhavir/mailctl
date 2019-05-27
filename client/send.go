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

func getPublicKey(config *Config, userHash []byte, params *hpke.Params) (pub crypto.PublicKey, err error) {
	var status byte
	// byte array for public key bytes
	pBytes := make([]byte, params.PubKeySize)

	conn, err := net.DialTCP("tcp", nil, config.parseIP())
	if err != nil {
		panic(err)
	}

	// specify op
	conn.Write([]byte{'g'})
	r := bufio.NewReader(conn)
	// get the response
	status, err = r.ReadByte()
	if err != nil || status != 0 {
		err = fmt.Errorf("server did not recognize the operation %v", err)
		return
	}

	// write 32 bytes of user/org hash identifier
	conn.Write(userHash)
	// get the response
	status, err = r.ReadByte()
	if err != nil || status == 1 {
		err = fmt.Errorf("username does not exist or his public key was not registered")
		return
	}

	_, err = r.Read(pBytes)
	if err != nil || status == 1 {
		err = fmt.Errorf("public key could not be received %v", err)
		return
	}
	conn.Close()

	// unmarshall public key from bytes
	pub, err = hpke.Unmarshall(params, pBytes)
	if err != nil {
		err = fmt.Errorf("incorrect public key received %v", err)
	}
	return
}

func send(config *Config, rcpt string, filepath string, subject string) {
	params, _ := hpke.GetParams(commons.HpkeMode)
	userHash := commons.Hash([]byte(rcpt))

	// get public key for recipient
	pub, err := getPublicKey(config, userHash, params)
	if err != nil {
		fmt.Println(err)
		return
	}

	// read file
	msg, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Printf("filepath \"%s\" does not exist\n", err)
		return
	}
	fmt.Printf("sending %s to %s\n", filepath, rcpt)

	// encrypt the message
	ct, enc, err := hpke.EncryptBase(params, nil, pub, msg, nil)
	if err != nil {
		fmt.Println("message could not be encrypted", err)
		return
	}

	var status byte
	conn, err := net.DialTCP("tcp", nil, config.parseIP())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// specify op
	conn.Write([]byte{'s'})
	r := bufio.NewReader(conn)
	// get the response
	status, err = r.ReadByte()
	if err != nil || status != 0 {
		err = fmt.Errorf("server did not recognize the operation %v", err)
		return
	}

	// write 32 bytes of user/org hash identifier
	conn.Write(userHash)
	// get the response
	status, err = r.ReadByte()
	if err != nil || status == 1 {
		fmt.Println("error during connection")
		return
	} else if status == 2 {
		fmt.Printf("username \"%s\" does not exist or his public key was not registered\n", rcpt)
		return
	}

	conn.Write([]byte(subject + "\n"))
	conn.Write(commons.Uint32ToByte(uint32(len(ct) + params.PubKeySize)))
	conn.Write(append(enc, ct...))

	// get the response
	status, err = r.ReadByte()
	if err != nil || status != 0 {
		fmt.Println("server did not receive the message ", err)
		return
	}

	fmt.Printf("file %s sent\n", filepath)

}
