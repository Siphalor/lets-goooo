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
