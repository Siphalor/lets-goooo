// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetDateFilename(t *testing.T) {
	values := [...]struct {
		time     time.Time
		expected string
	}{
		{time.Date(2021, time.December, 25, 0, 0, 0, 0, time.Local), "20211225"},
		{time.Date(2021, time.December, 25, 23, 59, 59, 999, time.Local), "20211225"},
		{time.Date(1, time.April, 5, 0, 0, 0, 0, time.Local), "00010405"},
	}

	for _, value := range values {
		assert.Equal(t, value.expected, GetDateFilename(value.time))
	}
}
