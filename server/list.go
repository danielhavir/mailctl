package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"path"

	"github.com/danielhavir/mailctl/internal/commons"
	"golang.org/x/crypto/blake2b"
)

func listFiles(r *bufio.Reader, conn net.Conn) {
	userHash := make([]byte, blake2b.Size256)
	if _, err := r.Read(userHash); err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}

	// register user if not already registered and respond
	userDir := path.Join(storage, string(commons.EncodeHex(userHash)))
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

	f, err := os.Open(userDir)
	if err != nil {
		log.Println(err)
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Println(err)
	}

	for _, file := range files {
		if file.Name() != "key.pub" {
			conn.Write([]byte(file.Name() + "\n"))
		}
	}
	conn.Write([]byte("EOF\n"))

	return
}
