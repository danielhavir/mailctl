package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"net"
	"path"
	"strings"

	"github.com/danielhavir/mailctl/internal/commons"
	"golang.org/x/crypto/blake2b"
)

func recvFromClient(r *bufio.Reader, conn net.Conn) {
	userHash := make([]byte, blake2b.Size256)
	if _, err := r.Read(userHash); err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}

	userDir := path.Join(storage, string(commons.EncodeHex(userHash)))
	pBytes, err := ioutil.ReadFile(path.Join(userDir, "key.pub"))
	if err != nil {
		conn.Write([]byte{2})
		return
	}
	conn.Write([]byte{0})

	conn.Write(commons.DecodeHex(pBytes))

	messageID, err := r.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}
	messageID = strings.TrimSuffix(messageID, "\n")

	ctLenBytes := make([]byte, 4)
	_, err = r.Read(ctLenBytes)
	if err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}

	ct := make([]byte, commons.ByteToUint32(ctLenBytes))
	_, err = r.Read(ct)
	if err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}

	err = ioutil.WriteFile(path.Join(userDir, messageID), ct, 0644)
	if err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}
	conn.Write([]byte{0})
}
