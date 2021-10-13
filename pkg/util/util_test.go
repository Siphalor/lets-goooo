package util

import (
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
