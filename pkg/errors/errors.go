package errors

import "fmt"

type ErrorStage string

const (
	ERROR_PREPROCESSOR = "preprocessor"
	ERROR_TOKENIZER    = "tokenizer"
	ERROR_PARSER       = "parser"
	ERROR_EVALUATOR    = "evaluator"
)

type Error struct {
	Stage   ErrorStage
	Message string
	Start   int
	End     int
}

func NewError(stage ErrorStage, message string, start int, end int) *Error {
	if end < start {
		end = start
	}
	return &Error{
		Stage:   stage,
		Message: message,
		Start:   start,
		End:     end,
	}
}

func NewPreprocessorError(message string, start int, end int) *Error {
	return NewError(ERROR_PREPROCESSOR, message, start, end)
}

func NewEvaluatorError(message string, start int, end int) *Error {
	return NewError(ERROR_EVALUATOR, message, start, end)
}

func NewTokenizerError(message string, start int, end int) *Error {
	return NewError(ERROR_TOKENIZER, message, start, end)
}

func NewParserError(message string, start int, end int) *Error {
	return NewError(ERROR_PARSER, message, start, end)
}

func (pe *Error) Error() string {
	return fmt.Sprintf("%s error: %s", pe.Stage, pe.Message)
}
