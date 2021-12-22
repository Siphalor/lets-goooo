// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package cmd

import "fmt"

type Error struct {
	message string
	code    int
	other   error
}

func NewError(code int, message string, other error) *Error {
	return &Error{
		message: message,
		code:    code,
		other:   other,
	}
}

func (error *Error) Code() int {
	return error.code
}

func (error *Error) Error() string {
	if error.other != nil {
		return fmt.Sprintf("error %d: %s: %v", error.code, error.message, error.other)
	}
	return fmt.Sprintf("error %d: %s", error.code, error.message)
}
