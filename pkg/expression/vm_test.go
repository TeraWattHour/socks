package expression

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"testing"
)

type Structure struct {
}

func (s *Structure) Method(ratio float64) string {
	return fmt.Sprintf("the ratio is %v", ratio)
}

type SomeInt int

func (s *SomeInt) Method() string {
	return " value of SomeInt is " + fmt.Sprintf("%d", *s)
}

func (s Structure) ReceiverMethod() string {
	return " non-pointer method"
}

func TestVM_Run(t *testing.T) {
	sets := []struct {
		expr   string
		expect any
	}{{
		"ordinals[nil ?: 1]",
		"2nd",
	}, {
		"voidMember?.method() ?: 1 != 1 ?: 420",
		false,
	}, {
		"voidMember?.property",
		nil,
	}, {
		"2 ** 3 / 4",
		2,
	}, {
		`not "str" in [true]`,
		false,
	}, {
		`base.structure.Method(123.4) + base.structure.ReceiverMethod() + someInt.Method()`,
		`the ratio is 123.4 non-pointer method value of SomeInt is 123`,
	}, {
		`12 ? 12 + 123 ** 2 : false`,
		12 + 123*123,
	}}

	for i, set := range sets {
		elements, err := tokenizer.Tokenize(fmt.Sprintf("{{ %s }}", set.expr))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		expr, err := Parse(elements[0].(*tokenizer.Mustache).Tokens)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		compiler := NewCompiler(expr.Expr)
		chunk, err := compiler.Compile()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		vm := NewVM(chunk)
		result, err := vm.Run(map[string]any{
			"ordinals": []string{"1st", "2nd"},
			"base": map[string]any{
				"structure": &Structure{},
			},
			"someInt": SomeInt(123),
		})
		if err != nil {
			t.Errorf("unexpected error for set %d: %v", i, err)
			return
		}
		if result != set.expect {
			t.Errorf("expected %v, got %v", set.expect, result)
		}
	}
}
