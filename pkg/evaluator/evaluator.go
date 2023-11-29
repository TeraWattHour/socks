package evaluator

import (
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/parser"
	"reflect"
)

type Evaluator struct {
	initialContent string
	currentContent string
	initialRunes   []rune
	runes          []rune
	programs       []parser.TagProgram
	context        map[string]interface{}
	state          map[string]interface{}
	i              int
	offset         int
}

func NewEvaluator(p *parser.Parser) *Evaluator {
	return &Evaluator{
		programs:       p.Programs,
		initialContent: p.Tokenizer.Template,
		runes:          p.Tokenizer.Runes,
	}
}

func (e *Evaluator) Evaluate(context map[string]interface{}) (result string, err error) {
	e.currentContent = e.initialContent
	e.initialRunes = []rune(e.initialContent)
	e.i = 0
	e.offset = 0

	e.state = make(map[string]interface{})
	e.context = context

	for e.i < len(e.programs) {

		program := e.programs[e.i]
		var evaluated interface{}
		var evaluatedString string
		if program.Tag.Kind != "preprocessor" {
			evaluated, err = e.evaluateStatement(program.Statement, e.context, e.state)
			if err != nil {
				return "", err
			}
			evaluatedString = fmt.Sprintf("%v", evaluated)
		} else {
			evaluatedString = ""
			e.i += 1
		}

		previousLength := len(e.runes)

		e.currentContent = program.Statement.Replace([]rune(evaluatedString), e.offset, e.runes)

		e.runes = []rune(e.currentContent)
		e.offset += len(e.runes) - previousLength
	}

	return e.currentContent, nil
}

func (e *Evaluator) evaluateStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) (interface{}, error) {
	switch statement.Kind() {
	case "variable":
		e.i += 1
		return e.evaluateVariableStatement(statement, context, state)
	case "for":
		return e.evaluateForStatement(statement, context, state)
	case "if":
		return e.evaluateIfStatement(statement, context, state)
	case "end":
		e.i += 1
		return "", nil
	}

	tag := statement.Tag()
	if tag == nil {
		return nil, errors.NewEvaluatorError("unexpected statement: "+statement.Kind(), -1, -1)
	}

	return nil, errors.NewEvaluatorError("unexpected statement kind: "+statement.Kind(), statement.Tag().Start, statement.Tag().End)
}

func convertToInterfaceSlice(obj interface{}) ([]interface{}, error) {
	sliceValue := reflect.ValueOf(obj)

	if sliceValue.Kind() != reflect.Slice {
		return nil, errors.NewEvaluatorError("object is not iterable", -1, -1)
	}

	resultSlice := make([]interface{}, sliceValue.Len())

	for i := 0; i < sliceValue.Len(); i++ {
		value := reflect.ValueOf(sliceValue.Index(i)).Interface()
		resultSlice[i] = value
	}

	return resultSlice, nil
}

func (e *Evaluator) evaluateIfStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) (interface{}, error) {
	ifStatement := statement.(*parser.IfStatement)
	result, err := expr.Run(ifStatement.Program, helpers.CombineMaps(context, state))
	if err != nil {
		return nil, errors.NewEvaluatorError("unable to evaluate expression: "+err.Error(), ifStatement.StartTag.Start, ifStatement.EndTag.End)
	}
	resultBool, ok := result.(bool)
	if !ok {
		return nil, errors.NewEvaluatorError("expression is not a boolean", ifStatement.StartTag.Start, ifStatement.EndTag.End)
	}

	ifBody := ifStatement.Body

	// Discard the first program (if statement)
	e.i += 1

	j := 0
	for e.programs[e.i+j].Tag.Start != ifStatement.EndTag.Start {
		j += 1
	}

	if resultBool {
		offset := 0
		before := e.i
		for e.i < before+j && e.i < len(e.programs) {
			program := e.programs[e.i]
			previousLength := len(ifBody)
			evaluated, err := e.evaluateStatement(program.Statement, context, state)
			if err != nil {
				return nil, err
			}
			evaluatedString := fmt.Sprintf("%v", evaluated)

			ifBody = []rune(program.Statement.Replace([]rune(evaluatedString), offset-ifStatement.StartTag.End-1, ifBody))

			newLength := len(ifBody)
			offset += newLength - previousLength
		}
		return string(ifBody), nil
	}
	e.i += j
	return "", nil
}

func (e *Evaluator) evaluateForStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) (interface{}, error) {
	forStatement := statement.(*parser.ForStatement)
	obj, err := e.evaluateVariableStatement(forStatement.Iterable, context, state)
	if err != nil {
		return nil, err
	}

	// Discard the first program (for statement)
	e.i += 1

	j := 0
	for e.programs[e.i+j].Tag.Start != forStatement.EndTag.Start {
		j += 1
	}

	loopBody := forStatement.Body
	result := ""

	values, err := convertToInterfaceSlice(obj)
	if err != nil {
		return nil, err
	}

	before := e.i
	for i, v := range values {
		currentLoopBody := loopBody
		offset := 0
		for e.i < before+j && e.i < len(e.programs) {
			program := e.programs[e.i]
			previousLength := len(currentLoopBody)
			if program.Statement.Kind() == "end" {
				e.i += 1
				continue
			}
			evaluated, err := e.evaluateStatement(program.Statement, context, helpers.CombineMaps(state, map[string]interface{}{
				forStatement.IteratorName: i,
				forStatement.ValueName:    v,
			}))
			if err != nil {
				return nil, err
			}
			evaluatedString := fmt.Sprintf("%v", evaluated)

			currentLoopBody = []rune(program.Statement.Replace([]rune(evaluatedString), offset-forStatement.StartTag.End-1, currentLoopBody))

			newLength := len(currentLoopBody)
			offset += newLength - previousLength
		}
		e.i = before
		result += string(currentLoopBody)
	}

	e.i += j + 1
	return result, nil
}

// Evaluator.evaluateVariableStatement evaluates variable statement and returns the result.
// This method doesn't increment the evaluator's index!
func (e *Evaluator) evaluateVariableStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) (interface{}, error) {
	vs, ok := statement.(*parser.VariableStatement)
	if !ok {
		return nil, errors.NewEvaluatorError("statement is not a variable statement", -1, -1)
	}

	result, err := expr.Run(vs.Program, helpers.CombineMaps(context, state))
	if err != nil {
		return nil, errors.NewEvaluatorError("unable to evaluate expression: "+err.Error(), vs.Tag().Start, vs.Tag().End)
	}

	return result, err
}
