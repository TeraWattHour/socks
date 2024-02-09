package evaluator

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/parser"
)

type Evaluator struct {
	programs []parser.Program
	result   string

	context   map[string]any
	sanitizer func(string) string

	i int
}

func NewEvaluator(programs []parser.Program, sanitizer func(string) string) *Evaluator {
	return &Evaluator{programs: programs, sanitizer: sanitizer}
}

func (e *Evaluator) Evaluate(context map[string]any) (string, error) {
	e.result = ""
	e.i = 0

	e.context = context

	for e.i < len(e.programs) {
		if err := e.evaluateProgram(e.programs[e.i], e.context); err != nil {
			return "", err
		}
	}

	return e.result, nil
}

func (e *Evaluator) evaluateProgram(program parser.Program, context map[string]any) error {
	switch program.Kind() {
	case "text":
		e.result += string(program.(parser.Text))
		e.i += 1
		return nil
	default:
		return e.evaluateStatement(program.(parser.Statement), context)
	}
}

func (e *Evaluator) evaluateStatement(statement parser.Statement, context map[string]any) error {
	switch statement.(type) {
	case *parser.PrintStatement:
		return e.evaluatePrintStatement(statement, context)
	case *parser.ForStatement:
		return e.evaluateForStatement(statement, context)
	case *parser.IfStatement:
		return e.evaluateIfStatement(statement, context)
	}

	tag := statement.Tag()
	if tag == nil {
		return errors.NewError("unexpected statement: " + statement.Kind())
	}

	return errors.NewError("unexpected statement kind: " + statement.Kind())
}

func (e *Evaluator) evaluateIfStatement(statement parser.Statement, context map[string]any) error {
	ifStatement := statement.(*parser.IfStatement)
	result, err := ifStatement.Program.Run(context)
	if err != nil {
		return errors.NewError("unable to evaluate expression: " + err.Error())
	}

	resultBool, ok := result.(bool)
	if !ok {
		return errors.NewError("expression is not a boolean")
	}

	// Discard the first tag program (if statement)
	e.i += 1

	before := e.i
	if resultBool {
		for e.i < before+ifStatement.Programs && e.i < len(e.programs) {
			err := e.evaluateProgram(e.programs[e.i], context)
			if err != nil {
				return err
			}
		}
		return nil
	}
	e.i += ifStatement.Programs
	return nil
}

func (e *Evaluator) evaluateForStatement(statement parser.Statement, context map[string]any) error {
	forStatement := statement.(*parser.ForStatement)
	obj, err := forStatement.Iterable.Run(context)
	if err != nil {
		return err
	}

	// Discard the first program (for statement)
	e.i += 1

	values := helpers.ConvertInterfaceToSlice(obj)
	if values == nil {
		return errors.NewError("for loop iterable must be either a slice, array or map")
	}

	before := e.i
	for i, v := range values {
		for e.i < before+forStatement.Programs {
			program := e.programs[e.i]
			if forStatement.KeyName != "" {
				context[forStatement.KeyName] = i
			}
			context[forStatement.ValueName] = v

			if err := e.evaluateProgram(program, context); err != nil {
				return err
			}
		}
		e.i = before
	}

	e.i = before + forStatement.Programs
	return nil
}

func (e *Evaluator) evaluatePrintStatement(statement parser.Statement, context map[string]any) error {
	printStatement := statement.(*parser.PrintStatement)

	result, err := printStatement.Program.Run(context)
	if err != nil {
		return errors.NewError("unable to evaluate expression: \n" + err.Error())
	}

	stringified := fmt.Sprintf("%v", result)
	if e.sanitizer != nil && printStatement.Tag().Sanitize {
		stringified = e.sanitizer(stringified)
	}

	e.result += stringified
	e.i += 1

	return err
}
