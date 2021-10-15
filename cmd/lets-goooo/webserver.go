package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

const tempDefaultString = `<html>
<head>
<title>Lets Goooo</title>
</head>
<body style="text-align:center">
<img src="assets/logoooo.svg" style="max-width:500px;" alt="Logo" />
</body>
</html>`

var tempDefault = template.Must(template.New("lets goooo").Parse(tempDefaultString)) //TODO get from other file

func RunWebservers() {
	wait := new(sync.WaitGroup)
	wait.Add(2)
	go func() {
		handlers := map[string]http.HandlerFunc{"/": defaultHandler}
		CreateWebserver(443, handlers)
		wait.Done()
	}()
	time.AfterFunc(2*time.Second, func() {
		handler := map[string]http.HandlerFunc{"/": defaultHandler}
		CreateWebserver(4443, handler)
		wait.Done()
	})
	wait.Wait()
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	tempDefault.ExecuteTemplate(w, "lets goooo", nil)
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
