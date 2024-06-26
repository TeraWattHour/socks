package socks

import (
	"fmt"
	"github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"io"
	"reflect"
	"slices"
)

type evaluator struct {
	programs []Statement

	file       helpers.File
	output     *helpers.Queue[Statement]
	writer     io.Writer
	staticMode bool
	context    map[string]any
	sanitizer  func(string) string

	i int
}

func newEvaluator(file helpers.File, programs []Statement, sanitizer func(string) string) *evaluator {
	return &evaluator{programs: programs, file: file, sanitizer: sanitizer}
}

func newStaticEvaluator(file helpers.File, output *helpers.Queue[Statement], programs []Statement, sanitizer func(string) string) *evaluator {
	return &evaluator{output: output, file: file, programs: programs, staticMode: true, sanitizer: sanitizer}
}

func (e *evaluator) evaluate(writer io.Writer, context Context) error {
	e.i = 0
	e.writer = writer
	e.context = context

	for e.i < len(e.programs) {
		if err := e.evaluateProgram(e.context); err != nil {
			return err
		}
	}

	return nil
}

func (e *evaluator) evaluateProgram(context Context) error {
	program := e.program()

	switch program := program.(type) {
	case *Text:
		if e.staticMode {
			e.output.Push(program)
		} else {
			if _, err := e.writer.Write([]byte(program.Content)); err != nil {
				return err
			}
		}

		e.i++
		return nil
	case *EndStatement:
		if e.staticMode && slices.Contains(*e.output, program.ClosedStatement) {
			e.output.Push(program)
		}
		e.i++
		return nil
	}

	// Evaluable programs (If, For, Expression) can be evaluated both at compile time and at runtime.
	// Any other program kind is left for the runtime evaluation but is expected to be evaluated
	// in its block, e.g. elif statement may not be evaluated on its own but only together with the if statement.
	prog, ok := program.(Evaluable)
	if !ok && e.staticMode || (e.staticMode && !helpers.Subset(prog.Dependencies(), availableInContext(context))) {
		e.output.Push(program)
		e.i++
		return nil
	} else if !ok {
		return e.error(fmt.Sprintf("unexpected %s statement encountered at runtime", program.Kind()), program.Location())
	}

	return prog.Evaluate(e, context)
}

func (e *evaluator) program() Statement {
	return e.programs[e.i]
}

func (e *evaluator) error(message string, location helpers.Location) error {
	return errors.New(message, e.file.Name, e.file.Content, location, location.FromOther())
}

func availableInContext(context map[string]any) []string {
	keys := reflect.ValueOf(context).MapKeys()
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.String()
	}
	return result
}

func applyLoopVariables(context map[string]any, keyName, valueName string, key, value any) {
	if keyName != "" {
		context[keyName] = key
	}
	context[valueName] = value
}
