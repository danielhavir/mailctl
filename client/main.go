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

func main() {
	configureCommand := flag.NewFlagSet("configure", flag.ExitOnError)
	configPathC := configureCommand.String("config-file", "", "Path to config file [optional] (default \"~/.mailctl/config.json\")")

	sendCommand := flag.NewFlagSet("send", flag.ExitOnError)
	rcpt := sendCommand.String("rcpt", "", "Recipient of the message (format: <user>@<organization>) [required]")
	file := sendCommand.String("file", "", "Path to file to be send [required]")
	subject := sendCommand.String("subject", "", "Path to file to be send [optional]")
	configPathS := sendCommand.String("config-file", "", "Path to config file [optional] (default \"~/.mailctl/config.json\")")

	recvCommand := flag.NewFlagSet("recv", flag.ExitOnError)
	messageID := recvCommand.String("message-id", "", "Message ID [required]")
	configPathR := recvCommand.String("config-file", "", "Path to config file [optional] (default \"~/.mailctl/config.json\")")

	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	configPathL := listCommand.String("config-file", "", "Path to config file [optional] (default \"~/.mailctl/config.json\")")

	if len(os.Args) < 2 {
		printHelp(configureCommand, sendCommand, recvCommand, listCommand)
		os.Exit(0)
	}

	var config *Config

	switch os.Args[1] {
	case "configure":
		configureCommand.Parse(os.Args[2:])
		err := configure(*configPathC)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	case "send":
		sendCommand.Parse(os.Args[2:])
		pswd, err := readPassword()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		config, err = readconfigfile(*configPathS, pswd)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		send(config, *rcpt, *file, *subject)
	case "recv":
		recvCommand.Parse(os.Args[2:])
		pswd, err := readPassword()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		config, err = readconfigfile(*configPathR, pswd)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		recv(config, *messageID)
	case "list":
		listCommand.Parse(os.Args[2:])
		pswd, err := readPassword()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		config, err = readconfigfile(*configPathL, pswd)
		list(config)
	case "h", "-h", "-help", "--help", "help":
		printHelp(configureCommand, sendCommand, recvCommand, listCommand)
		os.Exit(0)
	default:
		fmt.Printf("%q is not valid command.\nvalid commands and their usage\n", os.Args[1])
		printHelp(configureCommand, sendCommand, recvCommand, listCommand)
		os.Exit(0)
	}
}
