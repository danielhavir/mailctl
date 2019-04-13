package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"

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
	fmt.Println()
	return
}

func writeconfigfile(config *Config, filepath string, key []byte) (err error) {
	if filepath == "" {
		filepath = path.Join(confDir, confFile)
	}
	confBytes, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return
	}
	h := hmac.New(sha256.New, key)
	h.Write(confBytes)
	hash := h.Sum(nil)
	confBytes = append(encodehex(hash[:]), confBytes...)
	err = writefile(confBytes, filepath)
	return
}

func readconfigfile(filepath string, key []byte) (config *Config, err error) {
	if filepath == "" {
		filepath = path.Join(confDir, confFile)
	}
	confBytes, err := readfile(filepath)
	if err != nil {
		return
	}
	// hex encoding doubles the size of the hash, hence 2*sha256.Size
	hash := decodehex(confBytes[:2*sha256.Size])
	confBytes = confBytes[2*sha256.Size:]
	h := hmac.New(sha256.New, key)
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
	if filepath == "" {
		filepath = path.Join(confDir, confFile)
	}
	var config *Config
	var key []byte

	if key, err = readPassword(); err != nil {
		return
	}
	if _, err = os.Stat(confDir); os.IsNotExist(err) {
		err = mkdir(confDir)
		if err != nil {
			return
		}
	} else if _, err = os.Stat(filepath); !os.IsNotExist(err) {
		config, err = readconfigfile(filepath, key)
		if err != nil {
			return
		}
		fmt.Println("overwriting existing configuration")
	} else {
		config = &Config{
			User:         "",
			Organization: "",
			Host:         "",
			Port:         1881,
		}
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter username [%s]: ", config.User)
	user, _ := reader.ReadString('\n')
	config.User = strings.TrimSuffix(user, "\n")
	fmt.Printf("Enter organization [%s]: ", config.Organization)
	org, _ := reader.ReadString('\n')
	config.Organization = strings.TrimSuffix(org, "\n")
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
	fmt.Printf("configuration was successfully created and stores in \"%s\"\n", filepath)
	return
}
