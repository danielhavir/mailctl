package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"path"

	hpke "github.com/danielhavir/go-hpke"
	"golang.org/x/crypto/blake2b"
)

func sendToClient(r *bufio.Reader, conn net.Conn) {
	params, _ := hpke.GetParams(hpkeMode)

	userHash := make([]byte, blake2b.Size256)
	if _, err := r.Read(userHash); err != nil {
		log.Fatal(err)
		conn.Write([]byte{1})
		return
	}

	// register user if not already registered and respond
	userDir := path.Join(storage, string(encodehex(userHash)))
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		conn.Write([]byte{2})
		return
	}
	conn.Write([]byte{0})

	err := verify(r, conn, userHash)
	if err != nil {
		log.Fatal(err)
		return
	}

	l, err := r.ReadByte()
	if err != nil {
		log.Fatal(err)
		conn.Write([]byte{1})
		return
	}
	messageID := make([]byte, l)
	_, err = r.Read(messageID)
	if err != nil {
		log.Fatal(err)
		conn.Write([]byte{1})
		return
	}
	ct, err := readfile(path.Join(userDir, string(messageID)))
	if err != nil {
		log.Fatal(err)
		conn.Write([]byte{2})
		return
	}
	conn.Write([]byte{0})

	conn.Write(uint32ToByte(uint32(len(ct))))
	conn.Write(ct[params.PubKeySize:])
	conn.Write(ct[:params.PubKeySize])

	status, err := r.ReadByte()
	if err != nil || status == 1 {
		log.Fatal(err)
	}
}
