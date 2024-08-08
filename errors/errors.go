package errors

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
)

type Error struct {
	Message  string
	Location helpers.Location
}

func New(message string, location helpers.Location) *Error {
	return &Error{
		Message:  message,
		Location: location,
	}
}

func (e *Error) Error() string {
	if e.Location.File == "" {
		return e.Message
	}
	return fmt.Sprintf("%s:%d:%d: %s", e.Location.File, e.Location.Line, e.Location.Column, e.Message)
}
