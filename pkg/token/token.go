package token

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

//ToDO: Startparameter KEY
const ValidTime = 120
const keyn = "thisis32bitlongpassphraseimusing"

func CreateToken(location string) (string, error) {

	if len(location) != 3 {
		return "", fmt.Errorf("Token creation failed, because location had wrong length: %v", len(location))
	}
	unencryptedToken := fmt.Sprintf("%12v:%s", int64(time.Now().Unix())/int64(ValidTime)*int64(ValidTime), location)

	return EncryptAES([]byte(keyn), unencryptedToken)
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

	if len(key) != 32 {
		return "", fmt.Errorf("key has wrong length")
	} else if len(ciphertext) != 32 {
		return "", fmt.Errorf("cipher has wrong length (%v)", len(ciphertext))
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	plain := make([]byte, len(cipher))
	c.Decrypt(plain, cipher)

	//Returns time:location
	return string(plain), nil
}

func CheckValidTime(token string) (bool, error) {

	plainToken, err := DecryptAES([]byte(keyn), token)
	if err != nil {
		return false, fmt.Errorf("decryption failed: %w", err)
	}
	tokenTime, err := strconv.ParseInt(strings.TrimSpace(strings.Split(plainToken, ":")[0]), 10, 64)
	if err != nil {
		return false, fmt.Errorf("splitting token failed: %w", err)
	}
	if time.Now().Unix()-tokenTime < (2 * ValidTime) {
		return true, nil
	}
	return false, nil
}
