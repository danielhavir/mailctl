package main

import (
	"bufio"
	"bytes"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"syscall"

	hpke "github.com/danielhavir/go-hpke"
	"github.com/danielhavir/mailctl/internal/commons"
	"github.com/danielhavir/xchacha20blake2b"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ssh/terminal"
)

// Config stores the client configuration
type Config struct {
	User         string `json:"user"`
	Organization string `json:"organization"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Status       uint8  `json:"status"`
}

const (
	confDir  = ".mailctl"
	confFile = "config.json"
)

func readPassword() (key []byte, err error) {
	fmt.Print("Enter Password: ")
	key, err = terminal.ReadPassword(int(syscall.Stdin))
	key = commons.Hash(key)
	fmt.Println()
	return
}

func generateKey(config *Config, confPath string, key []byte) (pBytes []byte, err error) {
	params, _ := hpke.GetParams(commons.HpkeMode)
	prv, pub, err := hpke.GenerateKeyPair(params, nil)
	if err != nil {
		return
	}

	pBytes, err = hpke.Marshall(params, pub)
	if err != nil {
		return
	}
	pBytes = commons.EncodeHex(pBytes)
	pubPath := path.Join(confPath, "key.pub")
	ioutil.WriteFile(pubPath, pBytes, 0644)

	sBytes, err := hpke.MarshallPrivate(params, prv)
	if err != nil {
		return
	}

	cipher, err := xchacha20blake2b.New(key)
	if err != nil {
		return
	}

	encBytes := cipher.Seal(nil, nil, sBytes, append([]byte(config.User), []byte(config.Organization)...))
	encBytes = commons.EncodeHex(encBytes)
	prvPath := path.Join(confPath, "key.pem")
	ioutil.WriteFile(prvPath, encBytes, 0644)

	return
}

func registerKey(config *Config, key, pub []byte) (status byte) {
	userHash := commons.Hash([]byte(config.User + config.Organization))

	// parse server IP from config file
	ip := net.ParseIP(config.Host)
	addr := net.TCPAddr{
		IP:   ip,
		Port: config.Port,
	}

	// dial a connection
	conn, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		fmt.Println("public key could not be registered, server could not be reached ", err)
		return 1
	}

	// specify op
	conn.Write([]byte{'c'})
	// get the response
	status, err = bufio.NewReader(conn).ReadByte()
	if err != nil || status != 0 {
		fmt.Println("server did not recognize the operation ", err)
		return
	}

	// write 32 bytes of user/org hash identifier
	conn.Write(userHash)
	// get the response
	status, err = bufio.NewReader(conn).ReadByte()
	if err != nil || status == 1 {
		fmt.Println("user could not be registered, check connection ", err)
		return
	}
	if status == 2 {
		fmt.Printf("username \"%s\" already exists within organization \"%s\"\n", config.User, config.Organization)
		return
	}

	// write 64 bytes hex encoded
	conn.Write(pub)
	status, err = bufio.NewReader(conn).ReadByte()
	if err != nil {
		fmt.Println("public key could not be registered, check connection ", err)
		return 1
	}
	return
}

func readKey(config *Config, confPath string, key []byte) (prv crypto.PrivateKey, err error) {
	if confPath == "" {
		usr, _ := user.Current()
		confPath = path.Join(usr.HomeDir, confDir)
	}

	encBytes, err := ioutil.ReadFile(path.Join(confPath, "key.pem"))
	if err != nil {
		return
	}
	encBytes = commons.DecodeHex(encBytes)
	cipher, err := xchacha20blake2b.New(key)
	if err != nil {
		return
	}

	sBytes, err := cipher.Open(nil, nil, encBytes, append([]byte(config.User), []byte(config.Organization)...))
	if err != nil {
		return
	}

	params, _ := hpke.GetParams(commons.HpkeMode)
	prv, err = hpke.UnmarshallPrivate(params, sBytes)
	return
}

func writeconfigfile(config *Config, confPath string, key []byte) (err error) {
	if confPath == "" {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		confPath = path.Join(usr.HomeDir, confDir)
	}
	confBytes, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return
	}
	h, err := blake2b.New256(key)
	if err != nil {
		return
	}
	h.Write(confBytes)
	hash := h.Sum(nil)
	confBytes = append(commons.EncodeHex(hash[:]), confBytes...)
	err = ioutil.WriteFile(path.Join(confPath, confFile), confBytes, 0644)
	return
}

func readconfigfile(confPath string, key []byte) (config *Config, err error) {
	if confPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		confPath = path.Join(usr.HomeDir, confDir)
	}
	confBytes, err := ioutil.ReadFile(path.Join(confPath, confFile))
	if err != nil {
		return
	}
	// hex encoding doubles the size of the hash, hence 2*blake2b.Size256
	hash := commons.DecodeHex(confBytes[:2*blake2b.Size256])
	confBytes = confBytes[2*blake2b.Size256:]
	h, err := blake2b.New256(key)
	if err != nil {
		return
	}
	h.Write(confBytes)
	hash2 := h.Sum(nil)
	if !bytes.Equal(hash, hash2[:]) {
		err = errors.New("password was incorrect or the integrity of the configuration was compromised")
		return
	}
	err = json.Unmarshal(confBytes, &config)
	return
}

func configure(confPath string) (err error) {
	usr, err := user.Current()
	if err != nil {
		return
	}

	if confPath == "" {
		confPath = path.Join(usr.HomeDir, confDir)
	}

	config := &Config{
		User:         "",
		Organization: "",
		Host:         "",
		Port:         1881,
	}
	var key []byte

	if key, err = readPassword(); err != nil {
		return
	}

	var overwritting bool
	overwritting = false

	filepath := path.Join(confPath, confFile)
	if _, err = os.Stat(confPath); os.IsNotExist(err) {
		err = os.Mkdir(confPath, 0755)
		if err != nil {
			return
		}
	} else if _, err = os.Stat(filepath); !os.IsNotExist(err) {
		config, err = readconfigfile(confPath, key)
		if err != nil {
			return
		}
		overwritting = true
		fmt.Println("overwriting existing configuration")
	}

	reader := bufio.NewReader(os.Stdin)
	// username and organization cannot be overwritten
	if !overwritting {
		fmt.Print("Enter username: ")
		user, _ := reader.ReadString('\n')
		if len(user) > 1 {
			config.User = strings.TrimSuffix(user, "\n")
		} else {
			err = errors.New("username cannot be empty")
			return
		}
		fmt.Print("Enter organization: ")
		org, _ := reader.ReadString('\n')
		if len(org) > 1 {
			config.Organization = strings.TrimSuffix(org, "\n")
		} else {
			err = errors.New("username cannot be empty")
			return
		}
	}
	fmt.Printf("Enter host address [%s]: ", config.Host)
	host, _ := reader.ReadString('\n')
	config.Host = strings.TrimSuffix(host, "\n")
	fmt.Printf("Enter port number [%d]: ", config.Port)
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSuffix(portStr, "\n")
	config.Port, err = strconv.Atoi(portStr)
	if err != nil {
		err = errors.New("port must be a valid number")
		return
	}

	if !overwritting {
		fmt.Println("generating keys...")
		pub, err := generateKey(config, confPath, key)
		if err != nil {
			return err
		}
		config.Status = registerKey(config, key, pub)
	}

	if config.Status == 2 {
		return errors.New("configure with another username")
	}

	err = writeconfigfile(config, confPath, key)
	if err != nil {
		return
	}
	fmt.Printf("configuration was successfully created and stored in \"%s\"\n", filepath)
	return
}

func verify(r *bufio.Reader, conn net.Conn, prv crypto.PrivateKey) (err error) {
	var len uint8
	len, err = r.ReadByte()
	if err != nil {
		return
	}

	params, _ := hpke.GetParams(commons.HpkeMode)
	ciphertext := make([]byte, len)
	rec, err := r.Read(ciphertext)
	if uint8(rec) < len {
		err = errors.New("expected and received ciphertext lenghts do not match")
		return
	}

	enc := make([]byte, params.PubKeySize)
	rec, err = r.Read(enc)
	if rec < params.PubKeySize {
		err = errors.New("expected and received ephemeral key's lenght do not match")
		return
	}

	msg, err := hpke.DecryptBase(params, prv, enc, ciphertext, nil)
	if err != nil {
		return
	}
	conn.Write(msg)

	return
}
