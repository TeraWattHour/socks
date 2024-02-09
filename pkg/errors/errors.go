package errors

import (
	"fmt"
)

type Error struct {
	Message string
}

func NewError(message string) *Error {
	return &Error{
		Message: message,
	}
}

func (pe *Error) Error() string {
	return fmt.Sprintf("%s %s: %s", "test_data/templates/header.html:2", colorize("ERROR", RED), bold(pe.Message))
}

type Color int

const (
	RED Color = iota + 31
	GREEN
	YELLOW
	BLUE
)

func colorize(content string, color Color) string {
	return fmt.Sprintf("\033[1;%dm%s\033[00m", color, content)
}

func bold(content string) string {
	return fmt.Sprintf("\033[1m%s\033[00m", content)
}
