package token

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetQrCode(t *testing.T) {

	_, err := GetQrCode("aProperURL", "")
	assert.Error(t, err, "token generation for QR Code did not fail with wrong location length")

}
