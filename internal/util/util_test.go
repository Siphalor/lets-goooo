// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package util

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHash(t *testing.T) {
	type typeA struct {
		Val string
	}
	type typeB struct {
		Val string
	}
	type typeC struct {
		val string
	}

	hashA1 := Hash(typeA{Val: "Hello"})
	assert.Equal(t, 20, len(hashA1), "hash does not have the correct size")
	hashA2 := Hash(typeA{Val: "World"})
	assert.False(t, string(hashA1) == string(hashA2), "hashes of different values are reportedly equal")
	hashA3 := Hash(typeA{Val: "Hello"})
	assert.True(t, string(hashA1) == string(hashA3), "hashes of identical values are reportedly not equal")
	hashB1 := Hash(typeB{Val: "Hello"})
	assert.False(t, string(hashA1) == string(hashB1), "hashes of types with identical structures and values are reportedly equal")
	assert.Panics(t, func() { Hash(typeC{val: "hi"}) }, "struct with no visible fields should not be hashable")
}

func TestBase64Encode(t *testing.T) {
	assert.Equal(t, "SGVsbG8gV29ybGQ=", Base64Encode([]byte("Hello World")))
}

func TestBase64Decode(t *testing.T) {
	decoded, err := Base64Decode("SGVsbG8gV29ybGQ=")
	if assert.NoError(t, err, "failed to decode valid base64 string") {
		assert.Equal(t, "Hello World", string(decoded))
	}
}

func TestWriteString(t *testing.T) {
	buffer := bytes.Buffer{}
	if assert.NoError(t, WriteString(&buffer, "test"), "buffer write should not fail") {
		assert.Equal(t, "test", buffer.String())
	}

	mockWriter_ := newMockWriter()
	if assert.NoError(t, WriteString(&mockWriter_, "test"), "sequential buffer write should not fail") {
		assert.Equal(t, "test", buffer.String())
	}

	errorWriter_ := newErrorWriter()
	assert.Error(t, WriteString(&errorWriter_, "test"), "error in buffer write should fail WriteString")
}

type mockWriter struct {
	buffer []byte
}

func newMockWriter() mockWriter {
	return mockWriter{buffer: make([]byte, 0)}
}

func (mw *mockWriter) Write(text []byte) (int, error) {
	mw.buffer = append(mw.buffer, text[0])
	return 1, nil
}

type errorWriter bool

func newErrorWriter() errorWriter {
	return false
}

func (ew *errorWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("test error")
}
