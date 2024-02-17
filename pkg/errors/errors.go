package errors

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
)

type Error struct {
	Message  string
	File     string
	Location helpers.Location
}

func New(message string, location helpers.Location) *Error {
	return &Error{
		Location: location,
		Message:  message,
	}
}

func (pe *Error) Error() string {
	if pe.File == "" {
		return pe.Message
	}
	return fmt.Sprintf("%s: %s", fmt.Sprintf("%s:%d:%d", pe.File, pe.Location.Line, pe.Location.Column), pe.Message)
}
