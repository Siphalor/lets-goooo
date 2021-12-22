package main

import (
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/token"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/util"
	"log"
	"net/http"
)

// cookieHandler decides where to redirect
func cookieHandler(w http.ResponseWriter, r *http.Request) {
	// without token no login or logout possible -> home
	if !r.URL.Query().Has("token") {
		homeHandler(w, r)
		return
	}

	userdataCookie := (*http.Cookie)(nil)
	userdataCookie, err := r.Cookie("Userdata")
	// no cookie with userdata -> login
	if err != nil {
		redirectIO(w, "login.html", (*journal.User)(nil), r.URL.Query().Get("token"))
		return
	}
	userdata, err := Validate(userdataCookie.Value)
	// no valid userdata -> login
	if err != nil {
		redirectIO(w, "login.html", (*journal.User)(nil), r.URL.Query().Get("token"))
		return
	}
	location, err := dataJournal.GetCurrentUserLocation(util.Base64Encode(userdata.Hash()))
	if err != nil || location == nil {
		// in no location -> login
		redirectIO(w, "login.html", &userdata, r.URL.Query().Get("token"))
	} else {
		// in a location -> logout
		redirectIO(w, "logout.html", &userdata, r.URL.Query().Get("token"))
	}
}

// redirectIO creates the data struct for login/logout templates
func redirectIO(w http.ResponseWriter, templateFile string, user *journal.User, toke string) {
	//validating token
	location, err := token.Validate(toke)
	if err != nil {
		log.Printf("invalid token: %v\n", err)
		writeError(w, 400, "invalid token")
		return
	}

	executeTemplate(w, templateFile, struct {
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
		log.Printf("invalid token: %v\n", err)
		writeError(w, 400, "invalid token")
		return
	}

	//create userdata
	err = r.ParseForm()
	if err != nil {
		log.Printf("couldn't parse form: %v\n", err)
		writeError(w, 400, "invalid form")
		return
	}
	user := r.Form.Get("name")
	address := r.Form.Get("address")

	userdata := journal.User{
		Name:    user,
		Address: address,
	}

	data := util.Base64Encode(([]byte)(userdata.ToJournalLine()))
	hash := util.Base64Encode(util.HashString(data + "\t" + cookieSecret))

	//create cookie
	userdataCookie := &http.Cookie{
		Name:  "Userdata",
		Value: data + ":" + hash,
	}

	//search for Cookie holding Userdata
	oldCookie, _ := r.Cookie("Userdata")
	if oldCookie == nil || (oldCookie.Value == userdataCookie.Value) {
		http.SetCookie(w, userdataCookie)
	}

	location, _ := dataJournal.GetCurrentUserLocation(util.Base64Encode(userdata.Hash()))
	if location != (*journal.Location)(nil) {
		//no Location to be logged in
		log.Printf("user is already elsewhere: %v\n", err)
		writeError(w, 400, "already logged in")
		return
	}

	//create entry in journal
	err = dataJournal.WriteEventUser(&userdata, tokenLocation, journal.LOGIN)
	if err != nil {
		log.Printf("couldn't write into journal: %v\n", err)
		writeError(w, 500, "failed to log in")
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
		log.Printf("invalid token: %v\n", err)
		writeError(w, 400, "invalid token")
		return
	}

	//search for Cookie holding Userdata
	userdataCookie, err := r.Cookie("Userdata")
	if err != nil {
		//no User to be logged out
		log.Printf("failed to read user cookie: %v\n", err)
		writeError(w, 400, "invalid session")
		return
	}

	userdata, err := Validate(userdataCookie.Value)

	if err != nil {
		//no User to be logged out
		log.Printf("failed to read user cookie: %v\n", err)
		writeError(w, 400, "invalid session")
		return
	}

	//check if user is at a location
	location, err := dataJournal.GetCurrentUserLocation(util.Base64Encode(userdata.Hash()))
	if err != nil {
		log.Printf("user is at no location: %v\n", err)
		writeError(w, 400, "you're not logged in anywhere")
		return
	}

	//check if token is valid for user location
	if tokenLocation != location {
		log.Printf("user is not at the token's location: %v\n", err)
		writeError(w, 400, "trying to log out from wrong location")
		return
	}

	//log out user
	err = dataJournal.WriteEventUser(&userdata, location, journal.LOGOUT)
	if err != nil {
		log.Printf("couldn't write into journal: %v\n", err)
		writeError(w, 500, "failed to log out")
		return
	}

	//return to Home
	redirectToHome(w, 302)
}
