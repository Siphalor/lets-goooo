package main

import (
	"github.com/stretchr/testify/assert"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"testing"
)

func TestValidate(t *testing.T) {
	expectedUser := journal.User{
		Name:    "Tom",
		Address: "Mosbach",
	}

	//Blank String to be validated
	_, err := Validate("")
	assert.Error(t, err)

	//String does not contain separator :
	_, err = Validate("notAProperInput")
	assert.Error(t, err)

	//Incorrect Secret
	cookieSecret = "secret"
	data := util.Base64Encode([]byte("Tom\tMosbach"))
	hash := util.Base64Encode(util.HashString(data + "\t" + "wrongCookie"))
	incorrectCookie := data + ":" + hash

	_, err = Validate(incorrectCookie)
	assert.Error(t, err)

	//Incorrect Hash
	cookieSecret = "secret"
	data = util.Base64Encode([]byte("Tom\tMosbach"))
	hash = util.Base64Encode([]byte(data + "\t" + cookieSecret))
	incorrectCookie = data + ":" + hash

	_, err = Validate(incorrectCookie)
	assert.Error(t, err)

	//Not Base64Encoded
	cookieSecret = "secret"
	unencodedData := "Tom\tMosbach"
	unencodedHash := util.HashString(unencodedData + "\t" + cookieSecret)
	correctCookie := unencodedData + ":" + string(unencodedHash)
	expectedUser = journal.User{
		Name:    "Tom",
		Address: "Mosbach",
	}

	_, err = Validate(correctCookie)
	assert.Error(t, err)

	//Correct Input
	cookieSecret = "secret"
	data = util.Base64Encode([]byte("Tom\tMosbach"))
	hash = util.Base64Encode(util.HashString(data + "\t" + cookieSecret))
	correctCookie = data + ":" + hash
	expectedUser = journal.User{
		Name:    "Tom",
		Address: "Mosbach",
	}

	user, err := Validate(correctCookie)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}
