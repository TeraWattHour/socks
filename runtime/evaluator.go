package runtime

import (
	"fmt"
	"github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"io"
	"reflect"
)

type Evaluator struct {
	programs []Statement

	staticOutput *helpers.Queue[Statement]
	writer       io.Writer
	staticMode   bool
	context      map[string]any
	sanitizer    func(string) string
}

func NewEvaluator(programs []Statement, sanitizer func(string) string) *Evaluator {
	return &Evaluator{programs: programs, sanitizer: sanitizer}
}

func NewStaticEvaluator(output *helpers.Queue[Statement], programs []Statement, sanitizer func(string) string) *Evaluator {
	return &Evaluator{staticOutput: output, programs: programs, staticMode: true, sanitizer: sanitizer}
}

func (e *Evaluator) Evaluate(writer io.Writer, context Context) error {
	e.writer = writer
	e.context = context

	for _, program := range e.programs {
		if err := e.evaluateProgram(program, e.context); err != nil {
			return err
		}
	}

	return nil
}

func (e *Evaluator) evaluateProgram(program Statement, context Context) error {
	if text, ok := program.(*Text); ok {
		return e.write(text.Content)
	}

	// Evaluable programs (If, For, Expression) can be evaluated both at compile time and at runtime.
	// Any other program kind is left for the runtime evaluation but is expected to be evaluated
	// in its block, e.g. elif statement may not be evaluated on its own but only together with the if statement.
	prog, ok := program.(Evaluable)
	if !ok && e.staticMode || (e.staticMode && !helpers.IsSubset(prog.Dependencies(), availableInContext(context))) {
		e.staticOutput.Push(program)
		return nil
	} else if !ok {
		return e.error(fmt.Sprintf("unexpected %s statement encountered at runtime", program.Kind()), program.Location())
	}

	return prog.Evaluate(e, context)
}

func (e *Evaluator) write(data any) error {
	if !e.staticMode {
		_, err := fmt.Fprint(e.writer, data)
		return err
	}

	if text, ok := data.(string); ok {
		e.staticOutput.Push(&Text{Content: text})
	} else {
		e.staticOutput.Push(data.(Statement))
	}

	return nil
}

func (e *Evaluator) error(message string, location helpers.Location) error {
	return errors.New(message, location)
}

func availableInContext(context map[string]any) []string {
	keys := reflect.ValueOf(context).MapKeys()
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.String()
	}

	return result
}
