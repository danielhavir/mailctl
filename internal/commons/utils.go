package commons

import (
	"encoding/binary"
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

// DecodeHex decodes a hex-encoded byte array
func DecodeHex(src []byte) []byte {
	dst := make([]byte, hex.DecodedLen(len(src)))
	hex.Decode(dst, src)
	return dst
}

// EncodeHex encodes a byte array
func EncodeHex(src []byte) []byte {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst
}

// Hash for hashing a byte array using blake2b
func Hash(data []byte) []byte {
	h := blake2b.Sum256(data)
	return h[:]
}

// ByteToUint32 converts a byte array into a uint32 number
func ByteToUint32(in []byte) (out uint32) {
	out = binary.BigEndian.Uint32(in)
	return
}

// Uint32ToByt3 converts a uint32 into a byte array
func Uint32ToByte(in uint32) (out []byte) {
	out = make([]byte, 4)
	binary.BigEndian.PutUint32(out, in)
	return
}
