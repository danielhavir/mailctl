package main

import "fmt"

func send(config *Config, rcpt string, filepath string, subject string) {
	fmt.Printf("Sending %s to %s with subject \"%s\" and config ", filepath, rcpt, subject)
	fmt.Println(config)
}
