package token

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var key = []byte("thisis32bitlongpassphraseimusing")

func TestEncryptAES(t *testing.T) {

	//Encryption fail check with wrong key length
	plain := "  1634639400:MOS"
	key = []byte("not32bitKey")
	_, err := EncryptAES(key, plain)
	assert.Error(t, err, "Encryption worked with wrong key")

	//Testing for proper function of EncryptAES
	key = []byte("thisis32bitlongpassphraseimusing")
	expected := "60fdbe3bb31230a70d2ffe5ebb9a4e3f"
	actual, err := EncryptAES(key, plain)

	assert.NoErrorf(t, err, "encryption did not work")
	assert.Equal(t, expected, actual, "encryption returned wrong cipher")
}

func TestDecryptAES(t *testing.T) {

	//Decryption fail check with wrong key length
	expectedPlain := "  1634639400:MOS"
	cipher, _ := EncryptAES(key, expectedPlain)
	key = []byte("not32bitKey")
	_, err := DecryptAES(key, cipher)
	assert.Error(t, err, "Decryption worked with wrong key")

	//Decryption fail check with wrong cipher length
	key = []byte("thisis32bitlongpassphraseimusing")
	_, err = DecryptAES(key, "tooShortCipher")
	assert.Error(t, err, "Decryption worked with wrong cipher")

	//Testing for proper function of DecryptAES
	actual, err := DecryptAES(key, cipher)

	assert.NoError(t, err, "decryption did not work")
	assert.Equal(t, expectedPlain, actual, "encrypted and afterwards decrypted string is not the same as at the beginning")
}

func TestCreateToken(t *testing.T) {
	location := "MOS"
	unencryptedExpectedToken := fmt.Sprintf("%12v:%s", int64(time.Now().Unix())/int64(ValidTime)*int64(ValidTime), location)
	expectedToken, _ := EncryptAES(key, unencryptedExpectedToken)

	actual, err := CreateToken(location)

	assert.NoErrorf(t, err, "token creation did not work")
	assert.Equal(t, expectedToken, actual, "Wrong Token created")
}
