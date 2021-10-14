package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func RunWebservers() {
	wait := new(sync.WaitGroup)
	wait.Add(2)
	go func() {
		CreateWebserver(443)
		wait.Done()
	}()
	time.AfterFunc(20*time.Second, func() {
		CreateWebserver(4443)
		wait.Done()
	})
	wait.Wait()
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	if name == "" {
		name = "World"
	}
	responseString := "<html><body>Hello " + name + "</body></html>"
	w.Write([]byte(responseString)) // unbedingt Templates verwenden!
}

func CreateWebserver(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", mainHandler) //TODO mainHandler für LogoSeite
	//http.HandleFunc("/login/", loginHandler)
	//http.HandleFunc("/logout/", logoutHandler)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	server.ListenAndServeTLS("certification/cert.pem", "certification/key.pem")

}
