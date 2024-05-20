package evaluator

import (
	"fmt"
	"io"
	"reflect"

	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/expression"
	"github.com/terawatthour/socks/pkg/parser"
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
		if err := e.evaluateProgram(e.programs[e.i], e.context); err != nil {
			return err
		}
	}

	return nil
}

func (e *Evaluator) evaluateProgram(program parser.Program, context map[string]any) error {
	switch program.Kind() {
	case "text":
		if _, err := e.w.Write([]byte(program.(*parser.Text).Content)); err != nil {
			return err
		}
		e.i += 1
		return nil
	case "end":
		e.i += 1
		return nil
	default:
		return e.evaluateStatement(program.(parser.Statement), context)
	}
}

func (e *Evaluator) evaluateStatement(statement parser.Statement, context map[string]any) error {
	switch statement.(type) {
	case *parser.Expression:
		return e.evaluatePrintStatement(statement, context)
	case *parser.ForStatement:
		return e.evaluateForStatement(statement, context)
	case *parser.IfStatement:
		return e.evaluateIfStatement(statement, context)
	}

	return fmt.Errorf("unexpected statement")
}

func (e *Evaluator) evaluateIfStatement(statement parser.Statement, context map[string]any) error {
	ifStatement := statement.(*parser.IfStatement)
	result, err := ifStatement.Program.Run(context)
	if err != nil {
		return err
	}

	resultBool := expression.CastToBool(result)

	e.i++
	for e.program() != ifStatement.EndStatement {
		if resultBool {
			err := e.evaluateProgram(e.program(), context)
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

func (e *Evaluator) evaluateForStatement(statement parser.Statement, context map[string]any) error {
	forStatement := statement.(*parser.ForStatement)
	obj, err := forStatement.Iterable.Run(context)
	if err != nil {
		return err
	}

	values := helpers.ConvertInterfaceToSlice(obj)
	if values == nil {
		return errors.New(fmt.Sprintf("for loop iterable must be either a slice, array or map, received %s", reflect.TypeOf(obj)), forStatement.Location())
	}

	e.i++
	programCount := 0
	before := e.i

	for i, v := range values {
		for e.program() != forStatement.EndStatement {
			if forStatement.KeyName != "" {
				context[forStatement.KeyName] = i
			}
			context[forStatement.ValueName] = v

			if err := e.evaluateProgram(e.program(), context); err != nil {
				return err
			}
		}
		if i == 0 {
			programCount = e.i - before
		}
		e.i = before
	}

	e.i = before + programCount + 1
	return nil
}

func (e *Evaluator) evaluatePrintStatement(statement parser.Statement, context map[string]any) error {
	printStatement := statement.(*parser.Expression)

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
