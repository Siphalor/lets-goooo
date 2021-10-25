package main

import (
	"fmt"
	"html/template"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/token"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

//TODO get from config
var logIOUrl = "https://localhost:4443/"

type QrCodeUrl struct {
	PngUrl   string
	Location string
}

func RunWebservers() {
	wait := new(sync.WaitGroup)
	wait.Add(2)
	go func() {
		handler := map[string]http.HandlerFunc{
			"/":       defaultHandler,
			"/login":  loginHandler,
			"/logout": logoutHandler,
		}
		RunWebserver(CreateWebserver(4443, handler))
		wait.Done()
	}()
	time.AfterFunc(2*time.Second, func() {
		handler := map[string]http.HandlerFunc{
			"/":       defaultHandler,
			"/qr.png": qrPngHandler,
			"/qr":     qrHandler,
		}
		RunWebserver(CreateWebserver(443, handler))
		wait.Done()
	})
	wait.Wait()
}

func defaultHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil)
}

func loginHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil) //TODO
}

func logoutHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil) //TODO
}

func qrPngHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	location := strings.ToUpper(q.Get("location"))

	if location == "" {
		log.Printf("there was no given location: %#v \n", location)
		if _, err := w.Write([]byte("no given location")); err != nil {
			log.Printf("failed to write qrcode to Response: %#v \n %#v \n", err, r)
			return
		}
		return
	}

	if len(location) != 3 {
		log.Printf("abbreviation has to be 3 characters, not %#v \n", len(location))
		if _, err := w.Write([]byte("couldn't resolve location-abbreviation. Need 3 characters")); err != nil {
			log.Printf("failed to write qrcode to Response: %#v \n", err)
			return
		}
		return
	}

	qrcode, err := token.GetQrCode(logIOUrl, location)
	if err != nil {
		log.Printf("failed to get qrcode: %#v \n", err)
		return
	}

	if _, err := w.Write(qrcode); err != nil {
		log.Printf("failed to write qrcode to Response: %#v \n", err)
		return
	}
}

func qrHandler(w http.ResponseWriter, r *http.Request) {
	data := QrCodeUrl{fmt.Sprintf("%v.png", r.URL.Path), r.URL.Query().Get("location")}
	executeTemplate(w, "qr.html", data)
}

//creates a Template from a given file and responses with the
func executeTemplate(w http.ResponseWriter, file string, data interface{}) {
	tempDefault, err := template.ParseFiles(file)
	if err != nil {
		log.Printf("failed to parse template: %#v \n", err)
		return
	}

	if err := tempDefault.Execute(w, data); err != nil {
		log.Printf("failed to execute template: %#v \n", err)
		return
	}
}

func CreateWebserver(port int, handlers map[string]http.HandlerFunc) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	for key, handler := range handlers {
		mux.HandleFunc(key, handler)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return &server

}

func RunWebserver(server *http.Server) {
	workdir := GetFilePath("cmd")
	log.Fatalln(server.ListenAndServeTLS(workdir+"certification/cert.pem", workdir+"certification/key.pem"))
}

func GetFilePath(search string) string {
	_, filename, _, _ := runtime.Caller(0)
	index := strings.LastIndex(filename, "cmd")
	workdir := filename[0:index]
	return workdir
}
