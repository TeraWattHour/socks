package socks

import (
	"fmt"
	"github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/expression"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/tokenizer"
	"maps"
	"reflect"
)

type Context = map[string]any

type Statement interface {
	Kind() string
	Location() helpers.Location

	// IsClosable returns true if the statement is supposed to be closed with an `@end...` statement
	IsClosable() bool
}

type Evaluable interface {
	Statement
	Dependencies() []string
	Evaluate(evaluator *evaluator, context Context) error
}

type Text struct {
	Content string
}

func (t *Text) Kind() string {
	return "text"
}

func (t *Text) IsClosable() bool {
	return false
}

func (t *Text) Location() helpers.Location {
	panic("unreachable")
}

// ---------------------- Expression Statement ----------------------

type Expression struct {
	Program      *expression.VM
	tag          *tokenizer.Mustache
	dependencies []string
}

func (expr *Expression) Evaluate(e *evaluator, context Context) (err error) {
	result, err := expr.Program.Run(context)
	if err != nil {
		return err
	}

	stringified := fmt.Sprintf("%v", result)
	if e.sanitizer != nil && expr.tag.Sanitize {
		stringified = e.sanitizer(stringified)
	}

	if e.staticMode {
		e.output.Push(&Text{Content: stringified})
	} else {
		_, err = e.writer.Write([]byte(stringified))
	}

	e.i++

	return
}

func (expr *Expression) IsClosable() bool {
	return false
}

func (expr *Expression) Dependencies() []string {
	return expr.dependencies
}

func (expr *Expression) Location() helpers.Location {
	return expr.tag.Location
}

func (expr *Expression) Kind() string {
	return "expression"
}

// ---------------------- If Statement ----------------------

type IfStatement struct {
	Program        *expression.VM
	location       helpers.Location
	dependencies   []string
	ElifStatements []Statement
	ElseStatement  Statement
	EndStatement   Statement
}

func (st *IfStatement) IsClosable() bool {
	return true
}

func (st *IfStatement) Dependencies() []string {
	return st.dependencies
}

func (st *IfStatement) Location() helpers.Location {
	return st.location
}

func (st *IfStatement) String() string {
	return fmt.Sprintf("%-8s", "IF")
}

func (st *IfStatement) Kind() string {
	return "if"
}

func (st *IfStatement) Evaluate(e *evaluator, context Context) error {
	result, err := st.Program.Run(context)
	if err != nil {
		return err
	}

	resultBool := expression.CastToBool(result)

	e.i++

	if resultBool {
		for (e.program() != st.EndStatement && e.program() != st.ElseStatement) &&
			(len(st.ElifStatements) == 0 || e.program() != st.ElifStatements[0]) {
			if err := e.evaluateProgram(context); err != nil {
				return err
			}
		}

		for e.program() != st.EndStatement {
			e.i++
		}
		e.i++

		return nil
	}

	for (e.program() != st.EndStatement && e.program() != st.ElseStatement) &&
		(len(st.ElifStatements) == 0 || e.program() != st.ElifStatements[0]) {
		e.i++
	}

	// at this point there are 3 possibilities:
	// 1. the program is an else statement
	// 2. the program is an elif statement
	// 3. the program is an end statement

	// end statement
	if e.program() == st.EndStatement {
		return nil
	}

	// else statement
	if e.program() == st.ElseStatement {
		e.i++
		for e.program() != st.EndStatement {
			if err := e.evaluateProgram(context); err != nil {
				return err
			}
		}

		if e.program() != st.EndStatement {
			panic("unreachable")
		}

		return nil
	}

	matchedElif := false

	// elif statement
	for i, elifStatement := range st.ElifStatements {
		result, err := elifStatement.(*ElifStatement).Program.Run(context)
		if err != nil {
			return err
		}

		resultBool := expression.CastToBool(result)

		e.i++
		for (e.program() != st.EndStatement && e.program() != st.ElseStatement) &&
			(i+1 >= len(st.ElifStatements) || e.program() != st.ElifStatements[i+1]) {
			if resultBool {
				if err := e.evaluateProgram(context); err != nil {
					return err
				}
			} else {
				e.i++
			}
		}

		if resultBool {
			matchedElif = true
			break
		}
	}

	if e.program() == st.ElseStatement {
		e.i++
	}

	for e.program() != st.EndStatement {
		if !matchedElif {
			if err := e.evaluateProgram(context); err != nil {
				return err
			}
		} else {
			e.i++
		}
	}

	return nil
}

type ElseStatement struct {
	location helpers.Location
}

func (vs *ElseStatement) IsClosable() bool {
	return false
}

func (vs *ElseStatement) Location() helpers.Location {
	return vs.location
}

func (vs *ElseStatement) Kind() string {
	return "else"
}

type ElifStatement struct {
	location helpers.Location
	Program  *expression.VM
}

func (vs *ElifStatement) IsClosable() bool {
	return false
}

func (vs *ElifStatement) Location() helpers.Location {
	return vs.location
}

func (vs *ElifStatement) Kind() string {
	return "elif"
}

// ---------------------- For Statement ----------------------

type ForStatement struct {
	Iterable     *expression.VM
	KeyName      string
	ValueName    string
	location     helpers.Location
	dependencies []string
	EndStatement *EndStatement
}

func (st *ForStatement) IsClosable() bool {
	return true
}

func (st *ForStatement) Dependencies() []string {
	return st.dependencies
}

func (st *ForStatement) Location() helpers.Location {
	return st.location
}

func (st *ForStatement) String() string {
	if st.KeyName != "" {
		return fmt.Sprintf("%-8s: %s, %s in [%p]", "FOR", st.KeyName, st.ValueName, st)
	}
	return fmt.Sprintf("%-8s: %s in [%p]", "FOR", st.ValueName, st)
}

func (st *ForStatement) Kind() string {
	return "for"
}

func (st *ForStatement) Evaluate(e *evaluator, context Context) error {
	obj, err := st.Iterable.Run(context)
	if err != nil {
		return err
	}

	if !helpers.IsIterable(obj) {
		return errors.New(fmt.Sprintf("expected <slice | array | map>, got <%st>", reflect.ValueOf(obj).Kind()), st.Location())
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
		for e.program() != st.EndStatement {
			applyLoopVariables(localContext, st.KeyName, st.ValueName, v.Key, v.Value)

			if err := e.evaluateProgram(localContext); err != nil {
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

	if e.program() != st.EndStatement {
		panic("expected end statement after for evaluation, the template is malformed")
	}

	e.i++

	return nil
}

type EndStatement struct {
	location        helpers.Location
	ClosedStatement Statement
}

func (es *EndStatement) Kind() string {
	return "end"
}

func (es *EndStatement) Location() helpers.Location {
	return es.ClosedStatement.Location()
}

func (es *EndStatement) IsClosable() bool {
	return false
}
