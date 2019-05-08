package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"path"

	"golang.org/x/crypto/blake2b"
)

const confDir = "mailctl"

func configure(path string) (exist bool, err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		return
	}
	return true, nil
}

func registerKey(r *bufio.Reader, conn net.Conn) {
	userHash := make([]byte, blake2b.Size256)
	if _, err := r.Read(userHash); err != nil {
		log.Fatal(err)
		conn.Write([]byte{1})
		return
	}

	// register user if not already registered and respond
	userDir := path.Join(storage, string(encodehex(userHash)))
	exist, err := configure(userDir)
	if err != nil {
		log.Fatal(err)
		conn.Write([]byte{1})
		return
	}
	if exist {
		conn.Write([]byte{2})
		return
	}

	conn.Write([]byte{0})

	pubBytes := make([]byte, 64)
	if _, err := r.Read(pubBytes); err != nil {
		conn.Write([]byte{1})
		return
	}
	// save key to appropriate directory
	pubPath := path.Join(userDir, "key.pub")
	if err := writefile(pubBytes, pubPath); err != nil {
		conn.Write([]byte{1})
		return
	}

	conn.Write([]byte{0})
	return
}
