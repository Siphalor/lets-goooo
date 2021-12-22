// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileExists(t *testing.T) {
	exists, err := FileExists("file_test.go")
	assert.NoError(t, err, "an error occurred while checking file existence: %#v", err)
	assert.True(t, exists, "found existent file to be non-existent")

	exists, err = FileExists("a-file-that-definitely-doesnt-exist")
	assert.NoError(t, err, "an error occurred while checking file existence: %#v", err)
	assert.False(t, exists, "a non-existent file reportedly exists")

	exists, err = FileExists("..")
	assert.NoError(t, err, "an error occurred while checking file existence: %#v", err)
	assert.False(t, exists, "directory has been reported as file")
}
