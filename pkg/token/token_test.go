package token

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"testing"
	"time"
)

var key = []byte("thisis32bitlongpassphraseimusing")

func TestEncryptAES(t *testing.T) {

	//Encryption fail check with wrong key length
	plain := "  1634639400:MOS"
	key = []byte("not32bitKey")
	_, err := EncryptAES(key, plain)
	assert.Error(t, err, "encryption worked with wrong key")

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
	assert.Error(t, err, "decryption worked with wrong key")

	//Decryption fail check with wrong cipher length
	key = []byte("thisis32bitlongpassphraseimusing")
	_, err = DecryptAES(key, "tooShortCipher")
	assert.Error(t, err, "decryption worked with wrong cipher")

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
	assert.Equal(t, expectedToken, actual, "wrong Token created")
}

func TestValidate(t *testing.T) {
	journal.Locations = map[string]*journal.Location{
		"MOS": {Code: "MOS", Name: "Mosbach"},
	}

	//Wrong token length
	location, err := Validate(encrypt("NotACorrectTokenLength"))
	assert.Error(t, err, "false length of token did not create an error during decryption")

	//text in place of timestamp
	location, err = Validate(encrypt("CorrectLengh:123"))
	assert.Error(t, err, "no fail with string in token")

	//No ":" for splitting
	location, err = Validate(encrypt("1234567891012MOS"))
	assert.Error(t, err, "no fail without : for splitting")

	//Outdated token
	location, err = Validate(encrypt("000000000001:MOS"))
	assert.Error(t, err)

	//Unknown location
	incorrectToken, _ := CreateToken("ZZZ")
	location, err = Validate(incorrectToken)
	assert.Error(t, err)

	correctToken, _ := CreateToken("MOS")
	location, err = Validate(correctToken)
	assert.Equal(t, journal.Locations["MOS"], location, "incorrect location from token")
	assert.NoError(t, err)
}

func encrypt(plain string) string {
	cipher, _ := EncryptAES(key, plain)
	return cipher
}
