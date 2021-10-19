package token

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"time"
)

func CreateToken(location string) (token string) {
	timestamp := fmt.Sprintf("%12d:%s", time.Now().Unix(), location)

	//ToDO: Startparameter KEY
	keyn := "thisis32bitlongpassphraseimusing"

	return EncryptAES([]byte(keyn), timestamp)
}

func EncryptAES(key []byte, plaintext string) string {

	c, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	out := make([]byte, len(plaintext))

	c.Encrypt(out, []byte(plaintext))

	return hex.EncodeToString(out)
}

func DecryptAES(key []byte, ciphertext string) string {

	cipher, _ := hex.DecodeString(ciphertext)

	c, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	plain := make([]byte, len(cipher))
	c.Decrypt(plain, cipher)

	return string(plain)
}
