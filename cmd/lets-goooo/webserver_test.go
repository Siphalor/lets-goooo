package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
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

func (rw *testResponseWriter) WriteHeader(statusCode int) {

}

func (rw *testResponseWriter) clear() {
	rw.content = ""
}

func TestExecuteTemplate(t *testing.T) {
	trw := testResponseWriter{content: ""}
	buf := bytes.Buffer{}
	defTem := "<html>\n  <head>\n    <title>Let's Goooo</title>\n  </head>\n  <body style=\"text-align:center;\">\n    <img src=\"assets/logoooo.svg\" style=\"max-width:500px;\" alt=\"Logo\" />\n  </body>\n</html>"
	dynTem := "<html>\n  <head>\n    <title>Let's Goooo</title>\n  </head>\n  <body style=\"text-align:center;\">\n    <p> {{.Text}} </p>\n  </body>\n</html>"

	tempDir, remover := CreateTempDir(t)
	fileName := tempDir + "/test.html"
	defer remover()

	//test the correct output by executing a basic template (without dynamic data)
	if err := ioutil.WriteFile(fileName, []byte(defTem), 0755); err != nil {
		fmt.Printf("Unable to write file: %v", err)
	}
	reset := LogToBuffer(&buf)
	executeTemplate(&trw, fileName, nil)
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
	executeTemplate(&trw, fileName, nil)
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
	executeTemplate(&trw, fileName, struct{ Text string }{Text: "text"})
	reset()
	assert.Equal(t, "", buf.String())
	dynTest = strings.ReplaceAll(dynTem, "{{.Text}}", "text")
	assert.Equal(t, dynTest, trw.content)
	buf.Reset()
	trw.clear()

	//test to use not existing Template -> shouldn't parse template
	reset = LogToBuffer(&buf)
	executeTemplate(&trw, "notExisting.html", nil)
	reset()
	assert.NotEqual(t, "", buf.String())
	buf.Reset()
}

func TestCreateRunWebserver(t *testing.T) {
	//Turn of ssl check, to avoid self-signed certificates error
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	server := CreateWebserver(443, nil)
	go RunWebserver(server)
	req, err := http.NewRequest("GET", "https://localhost", nil)
	assert.NoError(t, err)
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, res.StatusCode)
	assert.NoError(t, server.Close())

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

func CreateTempDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err, "internal error: failed to create temp dir")
	return tempDir, func() {
		_ = os.Remove(tempDir)
	}
}
