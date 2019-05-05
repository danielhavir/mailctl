package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ssh/terminal"
)

// Config stores the client configuration
type Config struct {
	User         string `json:"user"`
	Organization string `json:"organization"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
}

const (
	confDir  = ".mailctl"
	confFile = "config.json"
)

func readPassword() (key []byte, err error) {
	fmt.Print("Enter Password: ")
	key, err = terminal.ReadPassword(int(syscall.Stdin))
	key = hash(key)
	fmt.Println()
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
	// hex encoding doubles the size of the hash, hence 2*sha256.Size
	hash := decodehex(confBytes[:2*sha256.Size])
	confBytes = confBytes[2*sha256.Size:]
	h, err := blake2b.New256(key)
	if err != nil {
		return
	}
	h.Write(confBytes)
	hash2 := h.Sum(nil)
	if !hmac.Equal(hash, hash2[:]) {
		err = errors.New("password was incorrect or the integrity of the configuration was compromised")
		return
	}
	err = json.Unmarshal(confBytes, &config)
	return
}

func configure(filepath string) (err error) {
	usr, err := user.Current()
	if err != nil {
		return
	}

	if filepath == "" {
		filepath = path.Join(usr.HomeDir, confDir, confFile)
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

	dir := path.Join(usr.HomeDir, confDir)
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = mkdir(dir)
		if err != nil {
			return
		}
	} else if _, err = os.Stat(filepath); !os.IsNotExist(err) {
		config, err = readconfigfile(filepath, key)
		if err != nil {
			return
		}
		fmt.Println("overwriting existing configuration")
	}

	reader := bufio.NewReader(os.Stdin)
	// username and organization cannot be overwritten
	if len(config.User) == 0 {
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
	err = writeconfigfile(config, filepath, key)
	if err != nil {
		return
	}
	fmt.Printf("configuration was successfully created and stored in \"%s\"\n", filepath)
	return
}
