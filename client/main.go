package main

import (
	"fmt"
	"log"
)

func main() {
	//_, err := readconfigfile(".mailctl/config.json")
	err := configure("")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("all good")
}
