package token

import (
	"fmt"
	qrcode "github.com/skip2/go-qrcode"
)

func GetQrCode(url string, location string) ([]byte, error) {

	token, err := CreateToken(location)
	if err != nil {
		return []byte{}, fmt.Errorf("could not create Token: %w", err)
	}

	var png []byte
	png, err = qrcode.Encode(url+"?token="+token, qrcode.Medium, 256)
	if err != nil {
		return []byte{}, fmt.Errorf("could not create QR Code: %w", err)
	}

	return png, nil
}
