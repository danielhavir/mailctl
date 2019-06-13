package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"

	hpke "github.com/danielhavir/go-hpke"
	"github.com/danielhavir/mailctl/internal/commons"
	"golang.org/x/crypto/blake2b"
)

const (
	confDir = "mailctl"
)

func configure(path string) (exist bool, err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		return
	}
	return true, nil
}

func verify(r *bufio.Reader, conn net.Conn, userHash []byte) (err error) {
	userDir := path.Join(storage, string(commons.EncodeHex(userHash)))
	pubPath := path.Join(userDir, "key.pub")
	pBytes, err := ioutil.ReadFile(pubPath)
	if err != nil {
		return
	}

	pBytes = commons.DecodeHex(pBytes)
	params, _ := hpke.GetParams(commons.HpkeMode)
	pub, err := hpke.Unmarshall(params, pBytes)
	if err != nil {
		return
	}

	msg := make([]byte, 128)
	rand.Read(msg)
	ciphertext, enc, err := hpke.EncryptBase(params, nil, pub, msg, nil)
	if err != nil {
		return
	}
	conn.Write([]byte{uint8(len(ciphertext))})
	conn.Write(ciphertext)
	conn.Write(enc)

	received := make([]byte, len(msg))
	_, err = r.Read(received)
	if err != nil {
		return
	}

	if !bytes.Equal(msg, received) {
		err = errors.New("authentication failed, original message and decrypted message are different")
	}
	return
}

func registerKey(r *bufio.Reader, conn net.Conn) {
	userHash := make([]byte, blake2b.Size256)
	if _, err := r.Read(userHash); err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}

	// register user if not already registered and respond
	userDir := path.Join(storage, string(commons.EncodeHex(userHash)))
	exist, err := configure(userDir)
	if err != nil {
		log.Println(err)
		conn.Write([]byte{1})
		return
	}
	if exist {
		conn.Write([]byte{2})
		return
	}

	conn.Write([]byte{0})

	params, _ := hpke.GetParams(commons.HpkeMode)
	pubBytes := make([]byte, 2*params.PubKeySize)
	if _, err := r.Read(pubBytes); err != nil {
		conn.Write([]byte{1})
		return
	}
	// save key to appropriate directory
	pubPath := path.Join(userDir, "key.pub")
	if err := ioutil.WriteFile(pubPath, pubBytes, 0400); err != nil {
		conn.Write([]byte{1})
		return
	}

	conn.Write([]byte{0})
	return
}
