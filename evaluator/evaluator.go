package evaluator

import (
	"fmt"
	"github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/expression"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/parser"
	"io"
	"maps"
	"reflect"
	"slices"
)

type Evaluator struct {
	programs []parser.Program

	output     *helpers.Queue[parser.Program]
	writer     io.Writer
	staticMode bool
	context    map[string]any
	sanitizer  func(string) string

	i int
}

func New(programs []parser.Program, sanitizer func(string) string) *Evaluator {
	return &Evaluator{programs: programs, sanitizer: sanitizer}
}

func NewStatic(output *helpers.Queue[parser.Program], programs []parser.Program, sanitizer func(string) string) *Evaluator {
	return &Evaluator{output: output, programs: programs, staticMode: true, sanitizer: sanitizer}
}

func (e *Evaluator) Evaluate(writer io.Writer, context map[string]any) error {
	e.i = 0
	e.writer = writer
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
		if e.staticMode {
			e.output.Push(program)
		} else {
			if _, err := e.writer.Write([]byte(program.Content)); err != nil {
				return err
			}
		}

		e.i++
		return nil
	case *parser.EndStatement:
		if e.staticMode && slices.Contains(*e.output, program.ClosedStatement) {
			e.output.Push(program)
		}
		e.i++
		return nil
	}

	prog, ok := program.(parser.WithDependencies)
	if !ok {
		if e.staticMode {
			e.output.Push(program)
			e.i++
			return nil
		} else {
			return errors.New(fmt.Sprintf("unexpected program type %T", program), program.Location())
		}
	}

	if program.Kind() == "for" {
		fmt.Println(prog.Dependencies(), availableInContext(context))
	}

	if e.staticMode && !helpers.Subset(prog.Dependencies(), availableInContext(context)) {
		e.output.Push(program)
		e.i++
		return nil
	}

	switch program := program.(type) {
	case *parser.Expression:
		return e.evaluatePrintStatement(program, context)
	case *parser.ForStatement:
		return e.evaluateForStatement(program, context)
	case *parser.IfStatement:
		return e.evaluateIfStatement(program, context)
	}

	panic("unreachable")
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
			if err := e.evaluate(e.program(), context); err != nil {
				return err
			}
		} else {
			if e.staticMode {
				e.output.Push(e.program())
			}
			e.i++
		}
	}

	if e.program() != ifStatement.EndStatement {
		panic("unreachable")
	}

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

	channel := make(chan helpers.KeyValuePair)
	go func() {
		helpers.ExtractValues(channel, obj)
		close(channel)
	}()

	e.i++
	programCount := 0
	before := e.i
	firstIteration := true

	localContext := make(map[string]any)
	maps.Copy(localContext, context)

	for v := range channel {
		for e.program() != forStatement.EndStatement {
			applyLoopVariables(localContext, forStatement.KeyName, forStatement.ValueName, v.Key, v.Value)

			if err := e.evaluate(e.program(), localContext); err != nil {
				return err
			}
		}
		if firstIteration {
			programCount = e.i - before
			firstIteration = false
		}
		e.i = before
	}

	e.i = before + programCount

	if e.program() != forStatement.EndStatement {
		panic("unreachable")
	}

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

	if e.staticMode {
		e.output.Push(&parser.Text{Content: stringified})
	} else {
		if _, err := e.writer.Write([]byte(stringified)); err != nil {
			return err
		}
	}

	e.i++

	return nil
}

func (e *Evaluator) program() parser.Program {
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

func applyLoopVariables(context map[string]any, keyName, valueName string, key, value any) {
	if keyName != "" {
		context[keyName] = key
	}
	context[valueName] = value
}
