// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package main

import (
	"fmt"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/token"
	"log"
	"net/http"
	"strings"
)

// qrPngHandler returns a picture of the qrCode
func qrPngHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	location := strings.ToUpper(q.Get("location"))

	if location == "" {
		log.Printf("there was no given location: %s\n", location)
		writeError(w, 400, "no given location")
		return
	}

	_, exists := journal.Locations[location]
	if !exists {
		fmt.Printf("failed to resolve location: %v\n", location)
		writeError(w, 400, "unknown location")
		return
	}

	qrcode, err := token.GetQrCode(logIOUrl, location)
	if err != nil {
		log.Printf("failed to get qrcode: %v\n", err)
		writeError(w, 400, "failed to generate qr code")
		return
	}

	if _, err := w.Write(qrcode); err != nil {
		log.Printf("failed to write qrcode to Response: %v\n", err)
		writeError(w, 400, "failed to send qr code")
		return
	}
}

// qrHandler creates thw qrCode response with data (qrCode is generated with location in the template)
func qrHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PngUrl   string
		Location string
	}{
		PngUrl:   fmt.Sprintf("%s.png", r.URL.Path),
		Location: r.URL.Query().Get("location"),
	}
	executeTemplate(w, "qr.html", data, false)
}
