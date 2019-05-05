package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"syscall"

	hpke "github.com/danielhavir/go-hpke"
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
	hpkeMode = hpke.BASE_X25519_SHA256_XChaCha20Blake2bSIV
)

func readPassword() (key []byte, err error) {
	fmt.Print("Enter Password: ")
	key, err = terminal.ReadPassword(int(syscall.Stdin))
	key = hash(key)
	fmt.Println()
	return
}

func generateKey(confPath string) (pBytes []byte, err error) {
	params, _ := hpke.GetParams(hpkeMode)
	prv, pub, err := hpke.GenerateKeyPair(params, nil)
	if err != nil {
		return
	}

	pBytes, err = hpke.Marshall(params, pub)
	if err != nil {
		return
	}
	pBytes = encodehex(pBytes)
	pubPath := path.Join(confPath, "key.pub")
	writefile(pBytes, pubPath)

	sBytes, err := hpke.MarshallPrivate(params, prv)
	if err != nil {
		return
	}
	sBytes = encodehex(sBytes)
	prvPath := path.Join(confPath, "key.pem")
	writefile(sBytes, prvPath)

	return
}

func registerKey(config *Config, key, pub []byte) (status byte, err error) {
	h, err := blake2b.New256(key)
	if err != nil {
		return
	}
	h.Write([]byte(config.User))
	h.Write([]byte(config.Organization))
	userHash := h.Sum(nil)

	// parse server IP from config file
	ip := net.ParseIP(config.Host)
	addr := net.TCPAddr{
		IP:   ip,
		Port: config.Port,
	}

	// dial a connection
	conn, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		return
	}

	// specify op
	conn.Write([]byte{'c'})
	conn.Write(userHash)
	conn.Write(pub)
	status, err = bufio.NewReader(conn).ReadByte()
	return
}

func writeconfigfile(config *Config, filepath string, key []byte) (err error) {
	if filepath == "" {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		filepath = path.Join(usr.HomeDir, confDir, confFile)
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
	confBytes = append(encodehex(hash[:]), confBytes...)
	err = writefile(confBytes, filepath)
	return
}

func readconfigfile(filepath string, key []byte) (config *Config, err error) {
	if filepath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		filepath = path.Join(usr.HomeDir, confDir, confFile)
	}
	confBytes, err := readfile(filepath)
	if err != nil {
		return
	}
	// hex encoding doubles the size of the hash, hence 2*blake2b.Size256
	hash := decodehex(confBytes[:2*blake2b.Size256])
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

	filepath := path.Join(confPath, confFile)

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

	if _, err = os.Stat(confPath); os.IsNotExist(err) {
		err = mkdir(confPath)
		if err != nil {
			return
		}
	} else if _, err = os.Stat(filepath); !os.IsNotExist(err) {
		config, err = readconfigfile(filepath, key)
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
		pub, err := generateKey(confPath)
		if err != nil {
			return err
		}
		config.Status, err = registerKey(config, key, pub)
		if err != nil {
			return err
		}
	}

	err = writeconfigfile(config, filepath, key)
	if err != nil {
		return
	}
	fmt.Printf("configuration was successfully created and stored in \"%s\"\n", filepath)
	return
}
