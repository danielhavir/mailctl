package main

import (
	"io/ioutil"
	"os"
)

func readfile(path string) (dat []byte, err error) {
	dat, err = ioutil.ReadFile(path)
	return
}

func writefile(text []byte, path string) (err error) {
	err = ioutil.WriteFile(path, text, 0664)
	return
}

func mkdir(path string) (err error) {
	err = os.Mkdir(path, 755)
	return
}
