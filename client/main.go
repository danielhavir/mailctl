package main

import (
	"flag"
	"fmt"
	"os"
)

func printHelp(flags ...*flag.FlagSet) {
	fmt.Println("\"configure\" (setting up configuration):")
	flags[0].PrintDefaults()
	fmt.Println("\"send\" (sending messages):")
	flags[1].PrintDefaults()
	fmt.Println("\"recv\" (receiving messages):")
	flags[2].PrintDefaults()
	fmt.Println("\"list\" (listing unread messages):")
	flags[3].PrintDefaults()
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func main() {
	configureCommand := flag.NewFlagSet("configure", flag.ExitOnError)
	confPathC := configureCommand.String("config-path", "", "Path to the directory with the config file [optional] (default \"~/.mailctl\")")

	sendCommand := flag.NewFlagSet("send", flag.ExitOnError)
	rcpt := sendCommand.String("rcpt", "", "Recipient of the message (format: <user>@<organization>) [required]")
	file := sendCommand.String("file", "", "Path to file to be send [required]")
	subject := sendCommand.String("subject", "", "Path to file to be send [optional]")
	confPathS := sendCommand.String("config-path", "", "Path to the directory with the config file [optional] (default \"~/.mailctl\")")

	recvCommand := flag.NewFlagSet("recv", flag.ExitOnError)
	messageID := recvCommand.String("message-id", "", "Message ID [required]")
	confPathR := recvCommand.String("config-path", "", "Path to the directory with the config file [optional] (default \"~/.mailctl\")")

	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	confPathL := listCommand.String("config-path", "", "Path to the directory with the config file [optional] (default \"~/.mailctl\")")

	if len(os.Args) < 2 {
		printHelp(configureCommand, sendCommand, recvCommand, listCommand)
		os.Exit(0)
	}

	var config *Config

	switch os.Args[1] {
	case "configure":
		configureCommand.Parse(os.Args[2:])
		err := configure(*confPathC)
		checkError(err)
	case "send":
		sendCommand.Parse(os.Args[2:])
		pswd, err := readPassword()
		checkError(err)
		config, err = readconfigfile(*confPathS, pswd)
		checkError(err)
		send(config, *rcpt, *file, *subject)
	case "recv":
		recvCommand.Parse(os.Args[2:])
		pswd, err := readPassword()
		checkError(err)
		config, err = readconfigfile(*confPathR, pswd)
		checkError(err)
		prv, err := readKey(config, *confPathL, pswd)
		checkError(err)
		recv(config, *messageID, pswd, prv)
	case "list":
		listCommand.Parse(os.Args[2:])
		pswd, err := readPassword()
		checkError(err)
		config, err = readconfigfile(*confPathL, pswd)
		checkError(err)
		prv, err := readKey(config, *confPathL, pswd)
		checkError(err)
		list(config, pswd, prv)
	case "h", "-h", "-help", "--help", "help":
		printHelp(configureCommand, sendCommand, recvCommand, listCommand)
		os.Exit(0)
	default:
		fmt.Printf("%q is not valid command.\nvalid commands and their usage\n", os.Args[1])
		printHelp(configureCommand, sendCommand, recvCommand, listCommand)
		os.Exit(0)
	}
}
