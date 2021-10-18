package token

import (
	"fmt"
	qrcode "github.com/skip2/go-qrcode"
)

func GetQrCode(url string, location string) []byte {
	token := CreateToken(location)
	var png []byte
	//ToDo: Dynamisch logIn/Out durch Cookie
	png, err := qrcode.Encode(url+"/login/"+token, qrcode.Medium, 256)
	if err != nil {
		fmt.Printf("Could not create QR Code: %v", err)
	}
	return png
}
