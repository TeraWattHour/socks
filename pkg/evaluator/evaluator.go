package evaluator

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"reflect"
)

type Evaluator struct {
	Content string
	Parser  *parser.Parser
	context map[string]interface{}
	state   map[string]interface{}
	i       int
	offset  int
}

func NewEvaluator(p *parser.Parser) *Evaluator {
	return &Evaluator{
		Content: p.Tokenizer.Template,
		Parser:  p,
		i:       0,
		offset:  0,
		state:   make(map[string]interface{}),
		context: make(map[string]interface{}),
	}
}

func (e *Evaluator) Evaluate(context map[string]interface{}) (string, error) {
	e.context = context

	for e.i < len(e.Parser.Programs) {
		previousLength := len(e.Content)

		program := e.Parser.Programs[e.i]
		var evaluated interface{}
		var evaluatedString string
		if program.Tag.Kind != "preprocessor" {
			evaluated = e.evaluateStatement(program.Statement, e.context, e.state)
			evaluatedString = fmt.Sprintf("%v", evaluated)
		} else {
			evaluatedString = ""
		}
		e.i += 1

		ru := []rune(e.Content)

		// for statement is a special case, we don't want to replace the whole Content since it evaluateForStatement
		// handles it itself
		if program.Statement.Kind() != "for" {
			e.Content = string(ru[:program.Tag.Start+e.offset]) + evaluatedString + string(ru[program.Tag.End+1+e.offset:])
		} else {
			e.Content = evaluatedString
		}

		e.offset += len(e.Content) - previousLength

	}

	return e.Content, nil
}

func (e *Evaluator) evaluateStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) interface{} {
	switch statement.Kind() {
	case "variable":
		return e.evaluateVariableStatement(statement, context, state)
	case "string":
		return statement.(*parser.StringStatement).Value
	case "integer":
		return statement.(*parser.IntegerStatement).Value
	case "for":
		return e.evaluateForStatement(statement, context, state)
	case tokenizer.TOK_END:
		return ""
	}

	panic("unreachable, unknown statement kind " + statement.Kind())
}

func (e *Evaluator) accessVariable(name string, obj interface{}) interface{} {
	t := reflect.TypeOf(obj)
	if t == nil {
		panic("variable is nil")
	}
	if t.Kind() == reflect.Map {
		return obj.(map[string]interface{})[name]
	} else if t.Kind() == reflect.Struct {
		return reflect.ValueOf(obj).FieldByName(name).Interface()
	}
	return nil
}

func convertToInterfaceSlice(obj interface{}) []interface{} {
	sliceValue := reflect.ValueOf(obj)
	if sliceValue.Kind() != reflect.Slice {
		return nil
	}

	resultSlice := make([]interface{}, sliceValue.Len())

	for i := 0; i < sliceValue.Len(); i++ {
		value := reflect.ValueOf(sliceValue.Index(i)).Interface()
		resultSlice[i] = value
	}

	return resultSlice
}

func combineMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func (e *Evaluator) evaluateForStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) interface{} {
	forStatement := statement.(*parser.ForStatement)
	obj := e.evaluateVariableStatement(forStatement.Iterable, context, state)

	tok := e.Parser.Tokenizer

	j := 0
	for e.Parser.Programs[e.i+1+j].Tag.Start != forStatement.EndTag.Start {
		j += 1
	}
	loopPrograms := e.Parser.Programs[e.i+1 : e.i+1+j]
	result := ""
	loopBody := tok.Runes[forStatement.StartTag.End+1 : forStatement.EndTag.Start]
	leading := tok.Runes[:forStatement.StartTag.Start]
	trailing := tok.Runes[forStatement.EndTag.End+1:]
	for i, v := range convertToInterfaceSlice(obj) {
		currentLoopBody := loopBody
		offset := 0
		for _, program := range loopPrograms {
			previousLength := program.Tag.End - program.Tag.Start
			evaluated := e.evaluateStatement(program.Statement, context, combineMaps(state, map[string]interface{}{
				forStatement.IteratorName: i,
				forStatement.ValueName:    v,
			}))
			evaluatedString := fmt.Sprintf("%v", evaluated)
			newLength := len(evaluatedString)
			currentLoopBody = []rune(fmt.Sprintf("%s%s%s", string(currentLoopBody[:program.Tag.Start-forStatement.StartTag.End-1+offset]), evaluatedString, string(currentLoopBody[program.Tag.End-forStatement.StartTag.End+offset:])))
			offset += newLength - previousLength - 1
		}
		result += string(currentLoopBody)
	}
	e.i += len(loopPrograms) + 1
	return string(leading) + result + string(trailing)
}

func (e *Evaluator) callFunction(args []parser.Statement, obj interface{}, context map[string]interface{}, state map[string]interface{}) interface{} {
	if reflect.TypeOf(obj).Kind() != reflect.Func {
		panic("tried to call a variable that is not a function")
	}

	functionArgs := []reflect.Value{}
	for _, arg := range args {
		evaluated := e.evaluateStatement(arg, context, state)
		functionArgs = append(functionArgs, reflect.ValueOf(evaluated))
	}

	results := reflect.ValueOf(obj).Call(functionArgs)
	if len(results) == 0 {
		panic("function call returned no results")
	}

	return results[0].Interface()
}

func (e *Evaluator) evaluateVariableStatement(statement parser.Statement, context map[string]interface{}, state map[string]interface{}) interface{} {
	vs, ok := statement.(*parser.VariableStatement)
	if !ok {
		panic("nice variable statement")
	}

	var obj interface{}
	if vs.IsLocal {
		obj = context
	} else {
		obj = state
	}

	for _, part := range vs.Parts {
		if part.Kind() == "variable_part" {
			obj = e.accessVariable(part.(*parser.VariablePartStatement).Name, obj)
		} else if part.Kind() == "function_call" {
			obj = e.callFunction(part.(*parser.FunctionCallStatement).Args, obj, context, state)
		}
	}

	return obj
}
