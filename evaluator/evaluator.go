package evaluator

import (
	"fmt"
	"github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/expression"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/parser"
	"io"
	"reflect"
)

type Evaluator struct {
	programs []parser.Program

	context   map[string]any
	sanitizer func(string) string
	w         io.Writer

	i int
}

func New(programs []parser.Program, sanitizer func(string) string) *Evaluator {
	return &Evaluator{programs: programs, sanitizer: sanitizer}
}

func (e *Evaluator) Evaluate(w io.Writer, context map[string]any) error {
	e.i = 0
	e.w = w
	e.context = context

	for e.i < len(e.programs) {
		if err := e.evaluate(e.programs[e.i], e.context); err != nil {
			return err
		}
	}

	return nil
}

func (e *Evaluator) evaluate(program parser.Program, context map[string]any) error {
	switch program := program.(type) {
	case *parser.Text:
		if _, err := e.w.Write([]byte(program.Content)); err != nil {
			return err
		}
		e.i += 1
		return nil
	case *parser.EndStatement:
		e.i += 1
		return nil
	case *parser.Expression:
		return e.evaluatePrintStatement(program, context)
	case *parser.ForStatement:
		return e.evaluateForStatement(program, context)
	case *parser.IfStatement:
		return e.evaluateIfStatement(program, context)
	default:
		return fmt.Errorf("unexpected program")
	}
}

func (e *Evaluator) evaluateIfStatement(ifStatement *parser.IfStatement, context map[string]any) error {
	result, err := ifStatement.Program.Run(context)
	if err != nil {
		return err
	}

	resultBool := expression.CastToBool(result)

	e.i++
	for e.program() != ifStatement.EndStatement {
		if resultBool {
			err := e.evaluate(e.program(), context)
			if err != nil {
				return err
			}
		} else {
			e.i++
		}
	}
	e.i++
	return nil
}

func (e *Evaluator) evaluateForStatement(forStatement *parser.ForStatement, context map[string]any) error {
	obj, err := forStatement.Iterable.Run(context)
	if err != nil {
		return err
	}

	if !helpers.IsIterable(obj) {
		return errors.New(fmt.Sprintf("expected <slice | array | map>, got <%s>", reflect.ValueOf(obj).Kind()), forStatement.Location())
	}

	channel := make(chan any)
	go func() {
		helpers.ExtractValues(channel, obj)
		close(channel)
	}()

	e.i++
	programCount := 0
	before := e.i
	i := 0

	previousKey := context[forStatement.KeyName]
	previousValue := context[forStatement.ValueName]

	for v := range channel {
		for e.program() != forStatement.EndStatement {
			if forStatement.KeyName != "" {
				context[forStatement.KeyName] = i
			}
			context[forStatement.ValueName] = v

			if err := e.evaluate(e.program(), context); err != nil {
				return err
			}
		}
		if i == 0 {
			programCount = e.i - before
		}
		i++
		e.i = before
	}

	context[forStatement.KeyName] = previousKey
	context[forStatement.ValueName] = previousValue

	e.i = before + programCount + 1
	return nil
}

func (e *Evaluator) evaluatePrintStatement(printStatement *parser.Expression, context map[string]any) error {
	result, err := printStatement.Program.Run(context)
	if err != nil {
		return err
	}

	stringified := fmt.Sprintf("%v", result)
	if e.sanitizer != nil && printStatement.Tag().Sanitize {
		stringified = e.sanitizer(stringified)
	}

	if _, err := e.w.Write([]byte(stringified)); err != nil {
		return err
	}

	e.i++

	return nil
}

func (e *Evaluator) program() parser.Program {
	return e.programs[e.i]
}
