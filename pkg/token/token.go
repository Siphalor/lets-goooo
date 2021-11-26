package token

import (
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
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

// Validate validates the given token and returns the contained journal.Location on success.
func Validate(token string) (*journal.Location, error) {
	plainToken, err := DecryptAES([]byte(keyn), token)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	parts := strings.SplitN(plainToken, ":", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid token data: not enough parts")
	}

	tokenTime, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time from token: %w", err)
	}

	locCode := strings.TrimSpace(parts[1])
	location, exists := journal.Locations[locCode]
	if !exists {
		return nil, fmt.Errorf("unknown location code: %v", locCode)
	}

	if time.Now().Unix()-tokenTime < (2 * ValidTime) {
		return location, nil
	}
	return nil, fmt.Errorf("token has timed out: token timestamp: %v", tokenTime)
}
