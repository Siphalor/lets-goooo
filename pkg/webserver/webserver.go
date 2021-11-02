package webserver

import (
	"log"
	"net/http"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	if name == "" {
		name = "World"
	}
	responseString := "<html><body>Hello " + name + "</body></html>"
	w.Write([]byte(responseString)) // unbedingt Templates verwenden!
}

func StartWebserver() {
	http.HandleFunc("/", mainHandler) //TODO mainHandler für LogoSeite
	//http.HandleFunc("/login/", loginHandler)
	//http.HandleFunc("/logout/", logoutHandler)
	log.Fatalln(http.ListenAndServeTLS(":4443", "certification/cert.pem", "certification/key.pem", nil))
}
