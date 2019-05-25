package commons

import (
	"encoding/binary"
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

func DecodeHex(src []byte) []byte {
	dst := make([]byte, hex.DecodedLen(len(src)))
	hex.Decode(dst, src)
	return dst
}

func EncodeHex(src []byte) []byte {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst
}

func Hash(data []byte) []byte {
	h := blake2b.Sum256(data)
	return h[:]
}

func ByteToUint32(in []byte) (out uint32) {
	out = binary.BigEndian.Uint32(in)
	return
}

func Uint32ToByte(in uint32) (out []byte) {
	out = make([]byte, 4)
	binary.BigEndian.PutUint32(out, in)
	return
}
