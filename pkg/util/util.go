package util

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/gob"
	"io"
	"log"
)

// Hash creates a SHA1 from the byte representation of a value
func Hash(val interface{}) []byte {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(val)
	if err != nil {
		log.Panicf("failed to hash value %T: %#v - %#v", val, val, err)
	}
	hash := sha1.New()
	hash.Write(b.Bytes())
	return hash.Sum(nil)
}

// Base64Encode base64 encodes bytes to a string.
func Base64Encode(val []byte) string {
	return base64.StdEncoding.EncodeToString(val)
}

// Base64Decode base64 decodes a string to raw bytes.
func Base64Decode(val string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(val)
}

func WriteString(writer io.Writer, text string) error {
	textBytes := []byte(text)
	for len(textBytes) > 0 {
		length, err := writer.Write(textBytes)
		if err != nil {
			return err
		}
		textBytes = textBytes[length:]
	}
	return nil
}
