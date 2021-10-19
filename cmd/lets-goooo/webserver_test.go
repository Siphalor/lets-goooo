package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"os"
	"testing"
)

type testResponseWriter struct {
	content string
}

func (rw testResponseWriter) Header() http.Header {
	return http.Header{}
}

func (rw testResponseWriter) Write(bytes []byte) (int, error) {
	rw.content += string(bytes)
	return len(bytes), nil
}

func (rw testResponseWriter) WriteHeader(statusCode int) {

}

func TestExecuteTemplate(t *testing.T) {
	trw := testResponseWriter{content: ""}
	buf := bytes.Buffer{}

	//test on not existing Template
	reset := LogToBuffer(&buf)
	executeTemplate(trw, "notExisting.html", nil)
	reset()
	assert.NotEqual(t, "", buf.String())
	buf.Reset()
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
