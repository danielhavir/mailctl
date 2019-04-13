package main

import (
	"encoding/hex"
	"io/ioutil"
	"os"
)

func readfile(filepath string) (dat []byte, err error) {
	dat, err = ioutil.ReadFile(filepath)
	return
}

func writefile(text []byte, filepath string) (err error) {
	err = ioutil.WriteFile(filepath, text, 0664)
	return
}

func decodehex(src []byte) []byte {
	dst := make([]byte, hex.DecodedLen(len(src)))
	hex.Decode(dst, src)
	return dst
}

func encodehex(src []byte) []byte {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst
}

func mkdir(dirpath string) (err error) {
	err = os.Mkdir(dirpath, 0755)
	return
}
