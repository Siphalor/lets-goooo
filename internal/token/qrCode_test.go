// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package token

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetQrCode(t *testing.T) {

	_, err := GetQrCode("aProperURL", "")
	assert.Error(t, err, "token generation for QR Code did not fail with wrong location length")

}
