package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"log"
	"net"
	"os"
	"path"

	hpke "github.com/danielhavir/go-hpke"
	"golang.org/x/crypto/blake2b"
)

const (
	confDir  = "mailctl"
	hpkeMode = hpke.BASE_X25519_SHA256_XChaCha20Blake2bSIV
)

func configure(path string) (exist bool, err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		return
	}
	return true, nil
}

func verify(r *bufio.Reader, conn net.Conn, userHash []byte) (err error) {
	userDir := path.Join(storage, string(encodehex(userHash)))
	pubPath := path.Join(userDir, "key.pub")
	pBytes, err := readfile(pubPath)
	if err != nil {
		return
	}

	pBytes = decodehex(pBytes)
	params, _ := hpke.GetParams(hpkeMode)
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
	userDir := path.Join(storage, string(encodehex(userHash)))
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

	params, _ := hpke.GetParams(hpkeMode)
	pubBytes := make([]byte, 2*params.PubKeySize)
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
