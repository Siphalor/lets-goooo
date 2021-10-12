package main

import "testing"

func TestAdd(t *testing.T) {
	a := 12
	b := 34
	if Add(a, b) != 46 {
		t.Errorf("%v + %v != 46", a, b)
	}
}
