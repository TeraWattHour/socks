package expression

import (
	"fmt"
	"github.com/terawatthour/socks/tokenizer"
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
		`not "str" in [false]`,
		true,
	}, {
		`base.structure.Method(123.4) + base.structure.ReceiverMethod() + someInt.Method()`,
		`the ratio is 123.4 non-pointer method value of SomeInt is 123`,
	}, {
		`sprintf("%.2f", 12 ? 12. + 123. ** 2.5 : 0.123)`,
		"167800.73",
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
			"sprintf": fmt.Sprintf,
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

func TestVM_Errors(t *testing.T) {
	sets := []struct {
		expr string
		err  string
	}{
		{
			`10 + someNilThing`,
			"undefined variable: someNilThing",
		},
	}

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

		_, err = NewVM(chunk).Run(map[string]any{})
		if err == nil {
			t.Errorf("expected error for set %d, got nil", i)
			return
		}
		fmt.Println(err)
		//if result != set.expect {
		//	t.Errorf("expected %v, got %v", set.expect, result)
		//}
	}
}
