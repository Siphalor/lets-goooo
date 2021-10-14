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
	tempDefault.ExecuteTemplate(w, "lets goooo", nil)

	//q := r.URL.Query()
	//name := q.Get("name")
	//if name == "" {
	//	name = "World"
	//}
	//responseString := "<html><body>Hello " + name + "</body></html>"
	//w.Write([]byte(responseString)) // unbedingt Templates verwenden!
}

func CreateWebserver(port int) {
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	mux.HandleFunc("/", mainHandler) //TODO mainHandler für LogoSeite
	//http.HandleFunc("/login/", loginHandler)
	//http.HandleFunc("/logout/", logoutHandler)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Fatalln(server.ListenAndServeTLS("certification/cert.pem", "certification/key.pem"))

}
