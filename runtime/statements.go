package runtime

import (
	"fmt"
	"github.com/terawatthour/socks/expression"
	"github.com/terawatthour/socks/internal/helpers"
	"maps"
)

type Context = map[string]any

type Statement interface {
	Kind() string
	Location() helpers.Location
}

type Evaluable interface {
	Statement
	Dependencies() []string
	Evaluate(evaluator *Evaluator, context Context) error
}

type Text struct {
	Content string
}

func (t *Text) Kind() string {
	return "text"
}

func (t *Text) Location() helpers.Location {
	panic("unreachable")
}

type Attribute struct {
	Name  string
	Value *expression.VM
}

func (a *Attribute) Kind() string {
	return "attribute"
}

func (a *Attribute) Dependencies() []string {
	return nil
}

func (a *Attribute) Evaluate(e *Evaluator, context Context) error {
	res, err := a.Value.Run(context)
	if err != nil {
		return err
	}

	return e.write(fmt.Sprintf(`%s="%s" `, a.Name, res))
}

func (a *Attribute) Location() helpers.Location {
	panic("unreachable")
}

// ---------------------- Expression Statement ----------------------

type Expression struct {
	Program *expression.VM
	//tag          *tokenizer.Mustache
	dependencies []string
}

func (expr *Expression) Evaluate(e *Evaluator, context Context) (err error) {
	result, err := expr.Program.Run(context)
	if err != nil {
		return err
	}

	stringified := fmt.Sprintf("%v", result)
	if e.sanitizer != nil {
		if _, ok := result.(expression.Raw); !ok {
			stringified = e.sanitizer(fmt.Sprintf("%v", result))
		}
	}

	return e.write(stringified)
}

func (expr *Expression) Dependencies() []string {
	return expr.dependencies
}

func (expr *Expression) Location() helpers.Location {
	panic("unreacahble")
	//return expr.tag.Location
}

func (expr *Expression) Kind() string {
	return "expression"
}

// ---------------------- If Statement ----------------------

type IfStatement struct {
	Program      *expression.VM
	location     helpers.Location
	dependencies []string

	Consequence  []Statement
	Alternatives []*ElifBranch
	Divergent    []Statement
}

func (st *IfStatement) Dependencies() []string {
	return st.dependencies
}

func (st *IfStatement) Location() helpers.Location {
	return st.location
}

func (st *IfStatement) Kind() string {
	return "if"
}

func (st *IfStatement) Evaluate(e *Evaluator, context Context) error {
	executeBlock := func(block []Statement) error {
		for _, p := range block {
			if err := e.evaluateProgram(p, context); err != nil {
				return err
			}
		}

		return nil
	}

	result, err := st.Program.Run(context)
	if err != nil {
		return err
	}

	if expression.CastToBool(result) {
		return executeBlock(st.Consequence)
	}

	for _, branch := range st.Alternatives {
		result, err := branch.Condition.Run(context)
		if err != nil {
			return err
		}

		if expression.CastToBool(result) {
			return executeBlock(branch.Consequence)
		}
	}

	return executeBlock(st.Divergent)
}

type ElifBranch struct {
	Condition   *expression.VM
	Consequence []Statement
}

// ---------------------- For Statement ----------------------

type ForStatement struct {
	Iterable     *expression.VM
	KeyName      string
	ValueName    string
	location     helpers.Location
	Body         []Statement
	dependencies []string
}

func (st *ForStatement) Dependencies() []string {
	return st.dependencies
}

func (st *ForStatement) Location() helpers.Location {
	return st.location
}

func (st *ForStatement) Kind() string {
	return "for"
}

func (st *ForStatement) Evaluate(e *Evaluator, context Context) error {
	obj, err := st.Iterable.Run(context)
	if err != nil {
		return err
	}

	if !helpers.IsIterable(obj) {
		return e.error(fmt.Sprintf("expected <slice | array | map>, got <%T>", obj), st.location)
	}

	channel := make(chan helpers.KeyValuePair)
	go func() {
		helpers.ExtractValues(channel, obj)
		close(channel)
	}()

	ctx := make(Context)
	maps.Copy(ctx, context)

	for v := range channel {
		helpers.ApplyVariable(ctx, st.ValueName, v.Value)
		helpers.ApplyVariable(ctx, st.KeyName, v.Key)

		for _, p := range st.Body {
			if err := e.evaluateProgram(p, ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

type Slot struct {
	Name     string
	Children []Statement
}

func (s *Slot) Kind() string {
	return "slot"
}

func (s *Slot) Location() helpers.Location {
	panic("unreachable")
}

type Component struct {
	Name    string
	Defines map[string][]Statement
}

func (t *Component) Kind() string {
	return "component"
}

func (t *Component) Location() helpers.Location {
	panic("unreachable")
}
