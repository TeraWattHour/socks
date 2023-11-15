package evaluator

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"reflect"
)

type Evaluator struct {
	initialContent string
	currentContent string
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
	e.i = 0
	e.offset = 0

	e.state = make(map[string]interface{})
	e.context = context

	for e.i < len(e.programs) {
		previousLength := len(e.currentContent)

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
		}
		e.i += 1

		ru := []rune(e.currentContent)

		// for statement is a special case, we don't want to replace the whole currentContent since it evaluateForStatement
		// handles it itself
		if program.Statement.Kind() != "for" {
			e.currentContent = string(ru[:program.Tag.Start+e.offset]) + evaluatedString + string(ru[program.Tag.End+1+e.offset:])
		} else {
			e.currentContent = evaluatedString
		}

		e.offset += len(e.currentContent) - previousLength
	}

	return e.currentContent, nil
}

func (e *Evaluator) evaluateStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) (interface{}, error) {
	switch statement.Kind() {
	case "variable":
		return e.evaluateVariableStatement(statement, context, state)
	case "string":
		return statement.(*parser.StringStatement).Value, nil
	case "integer":
		return statement.(*parser.IntegerStatement).Value, nil
	case "for":
		return e.evaluateForStatement(statement, context, state)
	case tokenizer.TOK_END:
		return "", nil
	}

	tag := statement.Tag()
	if tag == nil {
		return nil, errors.NewEvaluatorError("unexpected statement: "+statement.Kind(), -1, -1)
	}

	return nil, errors.NewEvaluatorError("unexpected statement kind: "+statement.Kind(), statement.Tag().Start, statement.Tag().End)
}

func (e *Evaluator) accessVariable(name string, obj interface{}) (interface{}, error) {
	t := reflect.TypeOf(obj)
	if t == nil {
		return nil, errors.NewEvaluatorError("variable not found", -1, -1)
	}
	if t.Kind() == reflect.Map {
		return obj.(map[string]interface{})[name], nil
	} else if t.Kind() == reflect.Struct {
		return reflect.ValueOf(obj).FieldByName(name).Interface(), nil
	}
	return nil, errors.NewEvaluatorError("variable cannot be accessed", -1, -1)
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

func (e *Evaluator) evaluateForStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) (interface{}, error) {
	forStatement := statement.(*parser.ForStatement)
	obj, err := e.evaluateVariableStatement(forStatement.Iterable, context, state)
	if err != nil {
		return nil, err
	}

	j := 0
	for e.programs[e.i+1+j].Tag.Start != forStatement.EndTag.Start {
		j += 1
	}
	loopPrograms := e.programs[e.i+1 : e.i+1+j]
	result := ""
	loopBody := e.runes[forStatement.StartTag.End+1 : forStatement.EndTag.Start]
	leading := e.runes[:forStatement.StartTag.Start]
	trailing := e.runes[forStatement.EndTag.End+1:]
	interSlice, err := convertToInterfaceSlice(obj)
	if err != nil {
		return nil, err
	}
	for i, v := range interSlice {
		currentLoopBody := loopBody
		offset := 0
		for _, program := range loopPrograms {
			previousLength := program.Tag.End - program.Tag.Start
			evaluated, err := e.evaluateStatement(program.Statement, context, helpers.CombineMaps(state, map[string]interface{}{
				forStatement.IteratorName: i,
				forStatement.ValueName:    v,
			}))
			if err != nil {
				return nil, err
			}
			evaluatedString := fmt.Sprintf("%v", evaluated)
			newLength := len(evaluatedString)
			currentLoopBody = []rune(fmt.Sprintf("%s%s%s", string(currentLoopBody[:program.Tag.Start-forStatement.StartTag.End-1+offset]), evaluatedString, string(currentLoopBody[program.Tag.End-forStatement.StartTag.End+offset:])))
			offset += newLength - previousLength - 1
		}
		result += string(currentLoopBody)
	}
	e.i += len(loopPrograms) + 1
	return string(leading) + result + string(trailing), nil
}

func (e *Evaluator) callFunction(args []parser.Statement, obj interface{}, context map[string]interface{}, state map[string]interface{}) (interface{}, error) {
	if reflect.TypeOf(obj).Kind() != reflect.Func {
		return nil, errors.NewEvaluatorError("tried to call a variable that is not a function", -1, -1)
	}

	functionArgs := []reflect.Value{}
	for _, arg := range args {
		evaluated, err := e.evaluateStatement(arg, context, state)
		if err != nil {
			return nil, err
		}
		functionArgs = append(functionArgs, reflect.ValueOf(evaluated))
	}

	results := reflect.ValueOf(obj).Call(functionArgs)
	if len(results) == 0 {
		return nil, errors.NewEvaluatorError("function call returned no results", -1, -1)
	}

	return results[0].Interface(), nil
}

func (e *Evaluator) evaluateVariableStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) (interface{}, error) {
	vs, ok := statement.(*parser.VariableStatement)
	if !ok {
		return nil, errors.NewEvaluatorError("statement is not a variable statement", -1, -1)
	}

	var obj interface{}
	var err error
	if vs.IsLocal {
		obj = context
	} else {
		obj = state
	}

	for _, part := range vs.Parts {
		if part.Kind() == "variable_part" {
			obj, err = e.accessVariable(part.(*parser.VariablePartStatement).Name, obj)
		} else if part.Kind() == "function_call" {
			obj, err = e.callFunction(part.(*parser.FunctionCallStatement).Args, obj, context, state)
		}
		if err != nil {
			return nil, err
		}
	}

	return obj, nil
}
