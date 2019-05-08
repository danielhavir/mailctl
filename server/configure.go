package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"

	"golang.org/x/crypto/blake2b"
)

const confDir = "mailctl"

func configure(path string) (configured bool, err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		return
	}
	return true, nil
}

func registerKey(r *bufio.Reader, conn net.Conn) {
	userHash := make([]byte, blake2b.Size256)
	if _, err := r.Read(userHash); err != nil {
		fmt.Println(err)
		conn.Write([]byte{1})
		return
	}

	// register user if not already registered
	userDir := path.Join(confPath, string(encodehex(userHash)))
	if _, err := configure(userDir); err != nil {
		fmt.Println(err)
		conn.Write([]byte{1})
		return
	}

	pubBytes := make([]byte, 64)
	if _, err := r.Read(pubBytes); err != nil {
		conn.Write([]byte{1})
		return
	}
	// save key to appropriate directory
	pubPath := path.Join(userDir, "key.pub")
	writefile(pubBytes, pubPath)

	conn.Write([]byte{0})
	return
}
