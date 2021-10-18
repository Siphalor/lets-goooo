package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

func RunWebservers() {
	wait := new(sync.WaitGroup)
	wait.Add(2)
	go func() {
		handlers := map[string]http.HandlerFunc{"/": defaultHandler}
		CreateWebserver(443, handlers)
		wait.Done()
	}()
	time.AfterFunc(2*time.Second, func() {
		handler := map[string]http.HandlerFunc{
			"/":   defaultHandler,
			"/qr": qrHandler,
		}
		CreateWebserver(4443, handler)
		wait.Done()
	})
	wait.Wait()
}

func defaultHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil)
}

func qrHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	location := q.Get("location")
	_ = location
}

func executeTemplate(w http.ResponseWriter, file string, data interface{}) {
	tempDefault, parseErr := template.ParseFiles(file)
	if parseErr != nil {
		fmt.Printf("failed to parse template: %#v", parseErr)
		return
	}
	exErr := tempDefault.Execute(w, data)
	if exErr != nil {
		fmt.Printf("failed to execute template: %#v", exErr)
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
