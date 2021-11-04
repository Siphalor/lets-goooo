package main

import (
	"context"
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

var logIOUrl = "https://localhost:4443/"

type QrCodeUrl struct {
	PngUrl   string
	Location string
}

// RunWebservers opening login/out and qrCode webservers at the given ports
func RunWebservers(portLogin int, portQr int) error {
	if portLogin == portQr {
		return fmt.Errorf("can't use the same port for two webservers")
	}

	//waitGroup to keep the method open until both servers were shut down
	wait := new(sync.WaitGroup)
	wait.Add(2)

	//creating webserver for QrCode
	handlerQR := map[string]http.HandlerFunc{
		"/":       defaultHandler,
		"/qr":     qrHandler,
		"/qr.png": qrPngHandler,
	}
	server, destroy := CreateWebserver(portQr, handlerQR)

	//starting webserver for QrCode
	go func() {
		if err := RunWebserver(server); err != http.ErrServerClosed {
			log.Printf("SSL server ListenAndServe: %v", err)
		}
		destroy()
		wait.Done()
	}()

	time.Sleep(time.Second) // To be sure that the server is up (or start failed)

	//creating webserver for LogIO
	handlerLogIO := map[string]http.HandlerFunc{
		"/":       defaultHandler,
		"/login":  loginHandler,
		"/logout": logoutHandler,
	}
	server, destroy = CreateWebserver(portLogin, handlerLogIO)

	//starting webserver for LogIO
	go func() {
		if err := RunWebserver(server); err != http.ErrServerClosed {
			log.Printf("SSL server ListenAndServe: %v", err)
		}
		destroy()
		wait.Done()
	}()

	time.Sleep(time.Second) // To be sure that the server is up (or start failed)

	wait.Wait()
	return nil
}

func defaultHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil, false)
}

func loginHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil, false) //TODO
}

func logoutHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil, false) //TODO
}

func qrPngHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	location := strings.ToUpper(q.Get("location"))

	if location == "" {
		log.Printf("there was no given location: %v \n", location)
		if _, err := w.Write([]byte("no given location")); err != nil {
			log.Printf("failed to write qrcode to Response: %#v \n %v \n", err, r)
			return
		}
		return
	}

	if len(location) != 3 {
		log.Printf("abbreviation has to be 3 characters, not %v \n", len(location))
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
	executeTemplate(w, "qr.html", data, false)
}

// executeTemplate creates a Template from a given file and writes the text filled with data into the http.Response
func executeTemplate(w http.ResponseWriter, file string, data interface{}, testData bool) {
	if !testData {
		file = GetFilePath() + "template/" + file
	}
	temp, err := template.ParseFiles(file)
	if err != nil {
		log.Printf("failed to parse template: %#v \n", err)
		return
	}

	if err := temp.Execute(w, data); err != nil {
		log.Printf("failed to execute template: %#v \n", err)
		return
	}
}

/*	CreateWebserver creates a webserver at the given port
 *	the handlers of the webserver are given by the handler map
 */
func CreateWebserver(port int, handlers map[string]http.HandlerFunc) (*http.Server, func()) {
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	for key, handler := range handlers {
		mux.HandleFunc(key, handler)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	destroy := func() {
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("SSL server Shutdown: %v", err)
		}
	}
	return &server, destroy
}

func RunWebserver(server *http.Server) error {
	workdir := GetFilePath()
	err := server.ListenAndServeTLS(workdir+"certification/cert.pem", workdir+"certification/key.pem")
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

func GetFilePath() string {
	_, filename, _, _ := runtime.Caller(0)
	index := strings.LastIndex(filename, "cmd")
	workdir := filename[0:index]
	return workdir
}
