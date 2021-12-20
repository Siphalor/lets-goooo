package main

import (
	"context"
	"fmt"
	"html/template"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/token"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

//setting default values for global variables
//will be overwritten in main class by flags
var logIOUrl = "https://localhost:4443/"
var dataJournal = (*journal.Writer)(nil)
var cookieSecret = ""

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

func defaultHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "default.html", nil, false)
}

func cookieHandler(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("token") {
		defaultHandler(w, r)
		return
	}

	userdataCookie := (*http.Cookie)(nil)
	userdataCookie, err := r.Cookie("Userdata")
	if err != nil {
		executeLogin(w, (*journal.User)(nil), r.URL.Query().Get("token"))
		return
	}
	userdata, err := Validate(userdataCookie.Value)
	if err != nil {
		executeLogin(w, (*journal.User)(nil), r.URL.Query().Get("token"))
		return
	}
	location, err := dataJournal.GetCurrentUserLocation(util.Base64Encode(userdata.Hash()))
	if err != nil || location == nil {
		executeLogin(w, &userdata, r.URL.Query().Get("token"))
	} else {
		executeLogout(w, &userdata, r.URL.Query().Get("token"))
	}
}

func executeLogin(w http.ResponseWriter, user *journal.User, toke string) {
	location, err := token.Validate(toke)
	if err != nil {
		log.Printf("invalid token: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	executeTemplate(w, "login.html", struct {
		User     *journal.User
		Location *journal.Location
		Token    string
	}{
		User:     user,
		Location: location,
		Token:    toke,
	}, false)
}

func executeLogout(w http.ResponseWriter, user *journal.User, toke string) {
	location, err := token.Validate(toke)
	if err != nil {
		log.Printf("invalid token: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	executeTemplate(w, "logout.html", struct {
		User     *journal.User
		Location *journal.Location
		Token    string
	}{
		User:     user,
		Location: location,
		Token:    toke,
	}, false)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	//check if token is valid
	tokenString := r.URL.Query().Get("token")
	tokenLocation, err := token.Validate(tokenString)
	if err != nil {
		log.Printf("invalid token: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	//create userdata
	err = r.ParseForm()
	if err != nil {
		log.Printf("couldn't parse form: %v \n", err)
		redirectToHome(w, 400)
		return
	}
	user := r.Form.Get("name")
	address := r.Form.Get("address")

	userdata := journal.User{
		Name:    user,
		Address: address,
	}

	//search for Cookie holding Userdata
	userdataCookie, err := r.Cookie("Userdata")
	if userdataCookie != nil {
		userdata, err = Validate(userdataCookie.Value)
	}

	if err != nil {

		data := util.Base64Encode(([]byte)(userdata.ToJournalLine()))
		hash := util.Base64Encode(util.HashString(data + "\t" + cookieSecret))

		//create cookie
		userdataCookie := &http.Cookie{
			Name:  "Userdata",
			Value: data + ":" + hash,
		}
		http.SetCookie(w, userdataCookie)
	}

	location, _ := dataJournal.GetCurrentUserLocation(util.Base64Encode(userdata.Hash()))
	if location != (*journal.Location)(nil) {
		//no Location to be logged in
		log.Printf("user is already elsewhere: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	//create entry in journal
	err = dataJournal.WriteEventUser(&userdata, tokenLocation, journal.LOGIN)
	if err != nil {
		log.Printf("couldn't write into journal: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	//return to Home
	redirectToHome(w, 302)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	//check if token is valid
	tokenString := r.URL.Query().Get("token")
	tokenLocation, err := token.Validate(tokenString)
	if err != nil {
		log.Printf("invalid token: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	//search for Cookie holding Userdata
	userdataCookie, err := r.Cookie("Userdata")
	if err != nil {
		//no User to be logged out
		log.Printf("failed to read user cookie: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	userdata, err := Validate(userdataCookie.Value)

	if err != nil {
		//no User to be logged out
		log.Printf("failed to read user cookie: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	//check if user is at a location
	location, err := dataJournal.GetCurrentUserLocation(util.Base64Encode(userdata.Hash()))
	if err != nil {
		log.Printf("user is at no location: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	//check if token is valid for user location
	if tokenLocation != location {
		log.Printf("user is not at the token's location: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	//log out user
	err = dataJournal.WriteEventUser(&userdata, location, journal.LOGOUT)
	if err != nil {
		log.Printf("couldn't write into journal: %v \n", err)
		redirectToHome(w, 400)
		return
	}

	//return to Home
	redirectToHome(w, 302)
}

func qrPngHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	location := strings.ToUpper(q.Get("location"))

	if location == "" {
		log.Printf("there was no given location: %s \n", location)
		if _, err := w.Write([]byte("no given location")); err != nil {
			log.Printf("failed to write qrcode to Response: %v \n %v \n", err, r)
			return
		}
		return
	}

	if len(location) != 3 {
		log.Printf("abbreviation has to be 3 characters, not %d \n", len(location))
		if _, err := w.Write([]byte("couldn't resolve location-abbreviation. Need 3 characters")); err != nil {
			log.Printf("failed to write qrcode to Response: %v \n", err)
			return
		}
		return
	}

	qrcode, err := token.GetQrCode(logIOUrl, location)
	if err != nil {
		log.Printf("failed to get qrcode: %v \n", err)
		return
	}

	if _, err := w.Write(qrcode); err != nil {
		log.Printf("failed to write qrcode to Response: %v \n", err)
		return
	}
}

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

// executeTemplate creates a Template from a given file and writes the text filled with data into the http.Response
func executeTemplate(w http.ResponseWriter, file string, data interface{}, directFilePath bool) {
	files := []string{file}
	if !directFilePath {
		base := GetFilePath() + "/"
		files = []string{base + "template/" + file, base + "template/head.html", base + "template/footer.html"}
	}
	temp, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("failed to parse template: %v \n", err)
		return
	}

	if err := temp.Execute(w, data); err != nil {
		log.Printf("failed to execute template: %v \n", err)
		return
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

func RunWebserver(server *http.Server) error {
	_ = GetFilePath()
	err := server.ListenAndServeTLS("certification/cert.pem", "certification/key.pem")
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

func GetFilePath() string {
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
