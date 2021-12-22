package main

import (
	"context"
	"fmt"
	"html/template"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/internal/journal"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

//setting default values for global variables
//values get overwritten in main class by flags
var logIOUrl = "https://localhost:4443/"
var dataJournal = (*journal.Writer)(nil)
var cookieSecret = ""
var certFile = "certification/cert.pem"
var keyFile = "certification/key.pem"

// RunWebservers opening login/out and qrCode webservers at the given ports
func RunWebservers(portLogin uint, portQr uint) error {
	if portLogin == portQr {
		return fmt.Errorf("can't use the same port for two webservers")
	}

	//waitGroup to keep the method open until both servers were shut down
	wait := new(sync.WaitGroup)
	wait.Add(2)

	//creating webserver for QrCode
	handlerQR := map[string]http.HandlerFunc{
		"/":       homeHandler,
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
		"/":       cookieHandler,
		"/login":  loginHandler,
		"/logout": logoutHandler,
	}
	server, destroy = CreateWebserver(portLogin, handlerLogIO)

	//starting webserver for QrCode
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

// homeHandler creates a default response
func homeHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil, false)
}

// executeTemplate creates a Template from a given file and writes the text filled with data into the http.Response
func executeTemplate(w http.ResponseWriter, file string, data interface{}, directFilePath bool) {
	files := []string{file}
	if !directFilePath {
		base := GetPathToWd() + "/"
		files = []string{base + "template/" + file, base + "template/head.html", base + "template/footer.html"}
	}
	//Create template from file
	temp, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("failed to parse template: %v \n", err)
		return
	}

	//Execute template with given data
	if err := temp.Execute(w, data); err != nil {
		log.Printf("failed to execute template: %v \n", err)
		return
	}
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	if _, err := fmt.Fprintf(w, "<h1>Error %v</h1><p>%v</p>", code, message); err != nil {
		println("Failed to write error to client")
	}
}

/*	CreateWebserver creates a webserver at the given port
 *	the handlers of the webserver are given by the handler map
 */
func CreateWebserver(port uint, handlers map[string]http.HandlerFunc) (*http.Server, func()) {
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

// RunWebserver starts the given server
func RunWebserver(server *http.Server) error {
	_ = GetPathToWd()
	err := server.ListenAndServeTLS(certFile, keyFile)
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

// GetPathToWd returns the path to the Working directory
func GetPathToWd() string {
	_, filename, _, _ := runtime.Caller(0)
	index := strings.LastIndex(filename, "cmd")
	if index == -1 {
		return filename
	} else {
		return filename[0:index]
	}
}

func redirectToHome(w http.ResponseWriter, statusCode int) {
	w.Header().Add("Location", logIOUrl)
	w.WriteHeader(statusCode)
}
