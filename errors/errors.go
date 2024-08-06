package errors

import (
	"github.com/terawatthour/socks/internal/helpers"
)

type Error struct {
	Filename    string
	Source      string
	Message     string
	Location    helpers.Location
	EndLocation helpers.Location
}

func New(message string, location, endLocation helpers.Location) *Error {
	return &Error{
		Message:     message,
		Location:    location,
		EndLocation: endLocation,
	}
}

func (e *Error) Error() string {
	return e.Message
	//lines := strings.Split(e.Source, "\n")
	//line := lines[e.Location.Line-1]
	//
	//re := regexp.MustCompile(`[^\t]`)
	//arrowLine := re.ReplaceAllString(line[:e.Location.Column-1], " ")
	//
	//eot := "␄"
	//if len(lines) != e.Location.Line {
	//	eot = ""
	//}
	//
	//arrowCount := e.EndLocation.Cursor - e.Location.Cursor
	//if arrowCount <= 0 {
	//	arrowCount = 1
	//}
	//
	//return fmt.Sprintf("  ┌─ %s:%d:%d:\n%d | %s%s\n  | %s%s\n%s", e.Filename, e.Location.Line, e.Location.Column, e.Location.Line, line, eot, arrowLine, strings.Repeat("^", arrowCount), e.Message)
}
