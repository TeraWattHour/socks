package parser

type Error struct {
	Message  string
	Location int
}

func NewParserError(message string, location int) *Error {
	return &Error{
		Message:  message,
		Location: location,
	}
}

func (pe *Error) Error() string {
	return pe.Message
}
