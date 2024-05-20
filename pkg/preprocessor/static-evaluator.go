package preprocessor

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/parser"
	"reflect"
	"slices"
)

type staticEvaluator struct {
	programs []parser.Program
	result   []parser.Program

	context   map[string]any
	available []string
	sanitizer func(string) string

	i int
}

func evaluate(programs []parser.Program, context map[string]any, sanitizer func(string) string) ([]parser.Program, error) {
	staticEvaluator := &staticEvaluator{
		programs:  programs,
		sanitizer: sanitizer,
		i:         0,
		context:   context,
		result:    make([]parser.Program, 0),
	}
	return staticEvaluator.evaluate()
}

func (e *staticEvaluator) evaluate() ([]parser.Program, error) {
	for e.i < len(e.programs) {
		if err := e.evaluateProgram(e.programs[e.i], e.context); err != nil {
			return nil, err
		}
	}

	return e.result, nil
}

func (e *staticEvaluator) evaluateProgram(program parser.Program, context map[string]any) error {
	switch program.Kind() {
	case "text":
		e.result = append(e.result, program)
		e.i++
		return nil
	case "end":
		end := program.(*parser.EndStatement)
		e.i++
		if slices.Contains(e.result, parser.Program(end.ClosedStatement)) {
			e.result = append(e.result, program)
		}
		return nil
	default:
		return e.evaluateStatement(program.(parser.Statement), context)
	}
}

func (e *staticEvaluator) evaluateStatement(statement parser.Statement, context map[string]any) error {
	switch statement.(type) {
	case *parser.Expression:
		return e.evaluateExpression(statement, context)
	case *parser.ForStatement:
		return e.evaluateForStatement(statement, context)
	case *parser.IfStatement:
		return e.evaluateIfStatement(statement, context)
	}
	return fmt.Errorf("unexpected statement")
}

func (e *staticEvaluator) evaluateIfStatement(statement parser.Statement, context map[string]any) error {
	ifStatement := statement.(*parser.IfStatement)
	if !helpers.Subset(ifStatement.Dependencies, availableInContext(context)) {
		e.result = append(e.result, ifStatement)
		e.i++
		return nil
	}

	result, err := ifStatement.Program.Run(context)
	if err != nil {
		return errors.New("unable to evaluate: "+err.Error(), ifStatement.Location())
	}

	resultBool, ok := result.(bool)
	if !ok {
		return errors.New("expression doesn't return a boolean", ifStatement.Location())
	}

	e.i++
	for e.program() != ifStatement.EndStatement {
		if resultBool {
			if err := e.evaluateProgram(e.program(), context); err != nil {
				return err
			}
		} else {
			e.i++
		}
	}
	e.i++
	return nil
}

func (e *staticEvaluator) evaluateForStatement(statement parser.Statement, context map[string]any) error {
	forStatement := statement.(*parser.ForStatement)
	if !helpers.Subset(forStatement.Dependencies, availableInContext(context)) {
		e.result = append(e.result, forStatement)
		e.i++
		return nil
	}

	obj, err := forStatement.Iterable.Run(context)
	if err != nil {
		return err
	}

	values := helpers.ConvertInterfaceToSlice(obj)
	if values == nil {
		return errors.New("for loop iterable must be either a slice, array or map", forStatement.Location())
	}

	e.i++

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
		if i != len(values)-1 {
			e.i = before
		}
	}
	delete(context, forStatement.KeyName)
	delete(context, forStatement.ValueName)
	e.i++

	return nil
}

func (e *staticEvaluator) evaluateExpression(statement parser.Statement, context map[string]any) error {
	printStatement := statement.(*parser.Expression)
	if !helpers.Subset(printStatement.Dependencies, availableInContext(context)) {
		e.result = append(e.result, printStatement)
		e.i++
		return nil
	}

	result, err := printStatement.Program.Run(context)
	if err != nil {
		return errors.New("unable to evaluate expression: "+err.Error(), printStatement.Location())
	}

	stringified := fmt.Sprintf("%v", result)
	if e.sanitizer != nil && printStatement.Tag().Sanitize {
		stringified = e.sanitizer(stringified)
	}

	e.result = append(e.result, &parser.Text{stringified})
	e.i++

	return err
}

func (e *staticEvaluator) program() parser.Program {
	return e.programs[e.i]
}

func availableInContext(context map[string]any) []string {
	keys := reflect.ValueOf(context).MapKeys()
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.String()
	}
	return result
}
