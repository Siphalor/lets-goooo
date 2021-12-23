package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/journal"
	"lehre.mosbach.dhbw.de/lets-goooo/v2/pkg/token"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

type testResponseWriter struct {
	content string
}

func (rw *testResponseWriter) Header() http.Header {
	return http.Header{}
}

func (rw *testResponseWriter) Write(bytes []byte) (int, error) {
	rw.content += string(bytes)
	return len(bytes), nil
}

func (rw *testResponseWriter) WriteHeader(_ int) {

}

func (rw *testResponseWriter) clear() {
	rw.content = ""
}

func testHandler(w http.ResponseWriter, _ *http.Request) {
	executeTemplate(w, "test.html", nil, true)
}

func TestExecuteTemplate(t *testing.T) {
	trw := testResponseWriter{content: ""}
	buf := bytes.Buffer{}
	defTem := "<html>\n  <head>\n    <title>Let's Goooo</title>\n  </head>\n  <body style=\"text-align:center;\">\n    <img src=\"assets/logoooo.svg\" style=\"max-width:500px;\" alt=\"Logo\" />\n  </body>\n</html>"
	dynTem := "<html>\n  <head>\n    <title>Let's Goooo</title>\n  </head>\n  <body style=\"text-align:center;\">\n    <p> {{.Text}} </p>\n  </body>\n</html>"

	tempDir := t.TempDir()
	fileName := tempDir + "/test.html"

	//test the correct output by executing a basic template (without dynamic data)
	if err := ioutil.WriteFile(fileName, []byte(defTem), 0755); err != nil {
		fmt.Printf("Unable to write file: %v", err)
	}
	reset := LogToBuffer(&buf)
	executeTemplate(&trw, fileName, nil, true)
	reset()
	assert.Equal(t, "", buf.String())
	assert.Equal(t, defTem, trw.content)
	buf.Reset()
	trw.clear()

	//test the correct output by executing a dynamic template (without dynamic data)
	if err := ioutil.WriteFile(fileName, []byte(dynTem), 0755); err != nil {
		fmt.Printf("Unable to write file: %v", err)
	}
	reset = LogToBuffer(&buf)
	executeTemplate(&trw, fileName, nil, true)
	reset()
	assert.Equal(t, "", buf.String())
	dynTest := strings.ReplaceAll(dynTem, "{{.Text}}", "")
	assert.Equal(t, dynTest, trw.content)
	buf.Reset()
	trw.clear()

	//test the correct output by executing a dynamic template (with dynamic data)
	if err := ioutil.WriteFile(fileName, []byte(dynTem), 0755); err != nil {
		fmt.Printf("Unable to write file: %v", err)
	}
	reset = LogToBuffer(&buf)
	executeTemplate(&trw, fileName, struct{ Text string }{Text: "text"}, true)
	reset()
	assert.Equal(t, "", buf.String())
	dynTest = strings.ReplaceAll(dynTem, "{{.Text}}", "text")
	assert.Equal(t, dynTest, trw.content)
	buf.Reset()
	trw.clear()

	//test to use not existing Template -> shouldn't parse template
	reset = LogToBuffer(&buf)
	executeTemplate(&trw, "notExisting.html", nil, false)
	reset()
	assert.NotEqual(t, "", buf.String())
	buf.Reset()
}

func TestCreateWebserver(t *testing.T) {
	if os.Getenv("webitesti") == "" {
		return
	}
	//Turn of ssl check, to avoid self-signed certificates error
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	//CreateWebserver with no handlers -> go default handler
	server, destroy := CreateWebserver(4443, nil)
	go func() {
		err := RunWebserver(server)
		assert.Error(t, http.ErrServerClosed, err)
	}()
	time.Sleep(time.Second)
	req, err := http.NewRequest("GET", "https://localhost:4443/", nil)
	assert.NoError(t, err)
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, res.StatusCode)
	destroy()

	//CreateWebserver with a handler (but no template)
	handler := map[string]http.HandlerFunc{
		"/": testHandler,
	}
	server, destroy = CreateWebserver(4444, handler)
	go func() {
		err := RunWebserver(server)
		assert.Error(t, http.ErrServerClosed, err)
	}()
	time.Sleep(time.Second)
	req, err = http.NewRequest("GET", "https://localhost:4444/", nil)
	assert.NoError(t, err)
	res, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	//CreateWebserver on a used port
	server2, destroy2 := CreateWebserver(4444, handler)

	go func() {
		err := RunWebserver(server2) // error because already running on port
		assert.NotEqual(t, http.ErrServerClosed, err)
	}()

	destroy()
	destroy2()
}

func TestHandlers(t *testing.T) {
	cookieSecret = "thisis32bitlongpassphrasetooyay"
	token.ValidTime = 120
	token.EncryptionKey = "thisis32bitlongpassphraseimusing"
	journal.Locations = map[string]*journal.Location{
		"MOS": {Name: "Mosbach", Code: "MOS"},
		"TST": {Name: "Test", Code: "TST"},
	}
	tempDir := t.TempDir()
	err := error(nil)
	dataJournal, err = journal.NewWriter(tempDir)
	defer func() {
		err := dataJournal.Close()
		assert.NoError(t, err)
	}()
	journal.FileCreationPermissions = 0777

	//different get parameters
	invalToken := url.Values{}
	invalToken.Set("token", "12345")

	validToken := url.Values{}
	toke, err := token.CreateToken("MOS")
	assert.NoError(t, err)
	validToken.Set("token", toke)

	invalLocat := url.Values{}
	invalLocat.Set("location", "this location does not exist")

	validLocat := url.Values{}
	validLocat.Set("location", "MOS")

	//homeHandler
	assert.HTTPStatusCode(t, homeHandler, "GET", "https://localhost", nil, 200) //reachable

	//cookieHandler
	assert.HTTPStatusCode(t, cookieHandler, "GET", "https://localhost", nil, 200)        //reachable
	assert.HTTPStatusCode(t, cookieHandler, "GET", "https://localhost", invalToken, 400) // redirecting with wrong token
	assert.HTTPStatusCode(t, cookieHandler, "GET", "https://localhost", validToken, 200) // redirecting with wrong token

	//loginHandler
	assert.HTTPStatusCode(t, loginHandler, "GET", "https://localhost", nil, 400)        //no token -> 400
	assert.HTTPStatusCode(t, loginHandler, "GET", "https://localhost", invalToken, 400) //wrong token -> 400
	assert.HTTPStatusCode(t, loginHandler, "GET", "https://localhost", validToken, 302) //correct token + not logged in -> log in + redirect to home
	assert.HTTPStatusCode(t, loginHandler, "GET", "https://localhost", validToken, 400) //correct token + already logged in -> already at location -> cant log in -> 400

	//logoutHandler
	assert.HTTPStatusCode(t, logoutHandler, "GET", "https://localhost", nil, 400)        //no token -> 400
	assert.HTTPStatusCode(t, logoutHandler, "GET", "https://localhost", invalToken, 400) //wrong token -> 400
	assert.HTTPStatusCode(t, logoutHandler, "GET", "https://localhost", validToken, 400) //correct token + no cookie -> 400

	//qrHandler
	assert.HTTPStatusCode(t, qrHandler, "GET", "https://localhost", nil, 200) //reachable

	//qrPngHandler
	assert.HTTPStatusCode(t, qrPngHandler, "GET", "https://localhost", nil, 400)        //no location -> 400
	assert.HTTPStatusCode(t, qrPngHandler, "GET", "https://localhost", invalLocat, 400) //no existing location -> 400
	assert.HTTPStatusCode(t, qrPngHandler, "GET", "https://localhost", validLocat, 200) // existing location -> 200
	//breaking token generation
	token.EncryptionKey = "thisisno32bitlongpassphrase"
	assert.HTTPStatusCode(t, qrPngHandler, "GET", "https://localhost", validLocat, 400) // cant generate QRCode -> 400
	token.EncryptionKey = "thisis32bitlongpassphraseimusing"
}

func TestRunWebservers(t *testing.T) {
	if os.Getenv("webitesti") == "" {
		return
	}
	//Turn of ssl check, to avoid self-signed certificates error
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	go func() {
		err := RunWebservers(442, 442)
		assert.Error(t, err)
	}()

	go func() {
		err := RunWebservers(443, 4443)
		assert.NoError(t, err)
	}()
	time.Sleep(time.Second)

	req, err := http.NewRequest("GET", "https://localhost", nil)
	assert.NoError(t, err)
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	req, err = http.NewRequest("GET", "https://localhost:4443", nil)
	assert.NoError(t, err)
	res, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

}

func LogToBuffer(buffer *bytes.Buffer) func() {
	log.SetOutput(buffer)
	flags := log.Flags()
	log.SetFlags(0)

	return func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(flags)
	}
}
