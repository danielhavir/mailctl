package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"

	"github.com/danielhavir/mailctl/internal/utils"
	"golang.org/x/crypto/blake2b"
)

func sendToClient(r *bufio.Reader, conn net.Conn) {
	userHash := make([]byte, blake2b.Size256)
	if _, err := r.Read(userHash); err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}

	userDir := path.Join(storage, string(utils.EncodeHex(userHash)))
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		conn.Write([]byte{2})
		return
	}
	conn.Write([]byte{0})

	err := verify(r, conn, userHash)
	if err != nil {
		log.Println(err)
		return
	}

	l, err := r.ReadByte()
	if err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}
	messageID := make([]byte, l)
	_, err = r.Read(messageID)
	if err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}
	ct, err := ioutil.ReadFile(path.Join(userDir, string(messageID)))
	if err != nil {
		log.Println(err)
		conn.Write([]byte{2})
		return
	}
	conn.Write([]byte{0})

	conn.Write(utils.Uint32ToByte(uint32(len(ct))))
	conn.Write(ct)

	status, err := r.ReadByte()
	if err != nil || status == 1 {
		log.Println(err)
	}
}
