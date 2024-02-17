package preprocessor

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/parser"
	"reflect"
)

type staticEvaluator struct {
	programs []parser.Program
	result   []parser.Program

	context      map[string]any
	available    []string
	sanitizer    func(string) string
	balanceQueue []balance

	i int
}

type balance struct {
	element parser.Statement
	delta   int
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

	for _, balance := range e.balanceQueue {
		balance.element.ChangeProgramCount(balance.delta)
	}

	return e.result, nil
}

func (e *staticEvaluator) evaluateProgram(program parser.Program, context map[string]any) error {
	switch program.Kind() {
	case "text":
		e.result = append(e.result, program)
		e.i += 1
		return nil
	default:
		return e.evaluateStatement(program.(parser.Statement), context)
	}
}

func (e *staticEvaluator) evaluateStatement(statement parser.Statement, context map[string]any) error {
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

func (e *staticEvaluator) evaluateIfStatement(statement parser.Statement, context map[string]any) error {
	ifStatement := statement.(*parser.IfStatement)
	if ifStatement.NoStatic() || !helpers.Subset(ifStatement.Dependencies, availableInContext(context)) {
		e.result = append(e.result, ifStatement)
		e.i += 1
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

	// Discard the first tag program (if statement)
	e.i += 1
	beforeCount := len(e.result)
	programCount := ifStatement.Programs
	before := e.i
	if resultBool {
		for e.i < before+programCount {
			err := e.evaluateProgram(e.programs[e.i], context)
			if err != nil {
				return err
			}
		}
		e.balanceQueue = append(e.balanceQueue, balance{element: ifStatement, delta: len(e.result) - beforeCount - programCount - 1})
		return nil
	}

	e.balanceQueue = append(e.balanceQueue, balance{element: ifStatement, delta: -programCount - 1})
	e.i += ifStatement.Programs
	return nil
}

func (e *staticEvaluator) evaluateForStatement(statement parser.Statement, context map[string]any) error {
	forStatement := statement.(*parser.ForStatement)
	if forStatement.NoStatic() || !helpers.Subset(forStatement.Dependencies, availableInContext(context)) {
		e.result = append(e.result, forStatement)
		e.i += 1
		return nil
	}

	obj, err := forStatement.Iterable.Run(context)
	if err != nil {
		return err
	}

	// Discard the first program (for statement)
	e.i += 1

	values := helpers.ConvertInterfaceToSlice(obj)
	if values == nil {
		return errors.New("for loop iterable must be either a slice, array or map", forStatement.Location())
	}

	before := e.i
	beforeCount := len(e.result)
	for i, v := range values {
		for e.i-before < forStatement.Programs {
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
	e.balanceQueue = append(e.balanceQueue, balance{element: forStatement, delta: len(e.result) - beforeCount - forStatement.Programs - 1})

	return nil
}

func (e *staticEvaluator) evaluatePrintStatement(statement parser.Statement, context map[string]any) error {
	printStatement := statement.(*parser.Expression)
	if printStatement.NoStatic() || !helpers.Subset(printStatement.Dependencies, availableInContext(context)) {
		e.result = append(e.result, printStatement)
		e.i += 1
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

	e.result = append(e.result, &parser.Text{Content: stringified, Parent: printStatement.Parent})
	e.i += 1

	return err
}

func availableInContext(context map[string]any) []string {
	keys := reflect.ValueOf(context).MapKeys()
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.String()
	}
	return result
}
