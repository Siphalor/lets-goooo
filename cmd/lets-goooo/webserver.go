package main

import (
	"fmt"
	"html/template"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/token"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var logIOUrl = "https://localhost"

type QrCodeUrl struct {
	PngUrl   string
	Location string
}

func RunWebservers() {
	wait := new(sync.WaitGroup)
	wait.Add(2)
	go func() {
		handlers := map[string]http.HandlerFunc{
			"/":       defaultHandler,
			"/login":  loginHandler,
			"/logout": logoutHandler,
		}
		CreateWebserver(443, handlers)
		wait.Done()
	}()
	time.AfterFunc(2*time.Second, func() {
		handler := map[string]http.HandlerFunc{
			"/":       defaultHandler,
			"/qr.png": qrPngHandler,
			"/qr":     qrHandler,
		}
		CreateWebserver(4443, handler)
		wait.Done()
	})
	wait.Wait()
}

func defaultHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil)
}

func loginHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil)
}

func logoutHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil)
}

func qrPngHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	location := strings.ToUpper(q.Get("location"))
	if len(location) != 3 {
		fmt.Printf("abbreviation has to be 3 characters, not %v", len(location))
		if _, err := w.Write([]byte("couldn't resolve location-abbreviation. Need 3 characters")); err != nil {
			fmt.Printf("failed to write qrcode to Response: %v", err)
			return
		}
		return
	}
	qrcode := token.GetQrCode(logIOUrl, location)
	/*
		if qrcodeString, err := token.GetQrCode(logIOUrl, location); err != nil {
			fmt.Printf("failed to get qrcode: %v", err)
			return
		}
	*/
	if _, err := w.Write(qrcode); err != nil {
		fmt.Printf("failed to write qrcode to Response: %v", err)
		return
	}
}

func qrHandler(w http.ResponseWriter, r *http.Request) {
	data := QrCodeUrl{fmt.Sprintf("%v.png", r.URL.Path), r.URL.Query().Get("location")}
	executeTemplate(w, "qr.html", data)
}

func executeTemplate(w http.ResponseWriter, file string, data interface{}) {
	tempDefault, err := template.ParseFiles(file)
	if err != nil {
		fmt.Printf("failed to parse template: %#v", err)
		return
	}
	err = tempDefault.Execute(w, data)
	if err != nil {
		fmt.Printf("failed to execute template: %#v", err)
		return
	}
}

func CreateWebserver(port int, handlers map[string]http.HandlerFunc) {
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	for key, handler := range handlers {
		mux.HandleFunc(key, handler)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Fatalln(server.ListenAndServeTLS("certification/cert.pem", "certification/key.pem"))

}
