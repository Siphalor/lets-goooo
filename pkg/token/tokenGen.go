package token

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"time"
)

func CreateToken(location string) (string, error) {

	ValidTime := 120

	timestamp := fmt.Sprintf("%12v:%s", int64(time.Now().Unix())/int64(ValidTime)*int64(ValidTime), location)

	//ToDO: Startparameter KEY
	keyn := "thisis32bitlongpassphraseimusing"

	return EncryptAES([]byte(keyn), timestamp)
}

func EncryptAES(key []byte, plaintext string) (string, error) {

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	out := make([]byte, len(plaintext))

	c.Encrypt(out, []byte(plaintext))

	//Returns a Cipher, which is the token for the URL
	return hex.EncodeToString(out), nil
}

func DecryptAES(key []byte, ciphertext string) (string, error) {

	cipher, _ := hex.DecodeString(ciphertext)

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	plain := make([]byte, len(cipher))
	c.Decrypt(plain, cipher)

	//Returns time:location
	return string(plain), nil
}
