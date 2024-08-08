package expression

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
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
	}{
		{
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
		}, {
			"float64(2) / float64(4)",
			0.5,
		}, {
			"range(1, 2, 1)[0]",
			1,
		}}

	for i, set := range sets {
		elements, err := Tokenize(set.expr, helpers.Location{Line: 1, Column: 1})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		expr, err := Parse(elements)
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
			t.Errorf("test %d: expected %v, got %v", i, set.expect, result)
			return
		}
		fmt.Println(result)
	}
}

//func TestVM_Errors(t *testing.T) {
//	sets := []struct {
//		expr string
//		err  string
//	}{
//		{
//			`someNilValue.accessed`,
//			"  ┌─ debug.txt:1:17:\n1 | {{ someNilValue.accessed }}␄\n  |                 ^^^^^^^^\ncan't access properties of <nil>",
//		}, {
//			"arrayValue[\"wrong_index\"]",
//			"  ┌─ debug.txt:1:15:\n1 | {{ arrayValue[\"wrong_index\"] }}␄\n  |               ^^^^^^^^^^^^^\nforbidden array index access, can't cast string to int",
//		}, {
//			"structValue[123]",
//			"  ┌─ debug.txt:1:16:\n1 | {{ structValue[123] }}␄\n  |                ^^^\nstruct field accessor must be of type string, got int64",
//		}, {
//			"(1 + 2)[3]",
//			"  ┌─ debug.txt:1:11:\n1 | {{ (1 + 2)[3] }}␄\n  |           ^^^\nforbidden access of properties of int64",
//		}, {
//			"someNilValue . method()",
//			"  ┌─ debug.txt:1:19:\n1 | {{ someNilValue . method() }}␄\n  |                   ^^^^^^\ncan't access properties of <nil>",
//		}, {
//			"someNilValue()",
//			"  ┌─ debug.txt:1:16:\n1 | {{ someNilValue() }}␄\n  |                ^^\ncan't call <nil>",
//		}, {
//			"float64(\"string\")",
//			"  ┌─ debug.txt:1:11:\n1 | {{ float64(\"string\") }}␄\n  |           ^^^^^^^^^^\ncan't cast string to float64",
//		},
//	}
//
//	for i, set := range sets {
//		file := helpers.File{"debug.txt", fmt.Sprintf("{{ %s }}", set.expr)}
//		elements, err := Tokenize("debug.txt", fmt.Sprintf("{{ %s }}", set.expr))
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//		expr, err := Parse(file, elements[0].(*tokenizer.Mustache).Tokens)
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		compiler := NewCompiler(file, expr.Expr)
//		program, err := compiler.Compile()
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		_, err = NewVM(file, program).Run(map[string]any{
//			"arrayValue":  []int{1, 2, 3, 4},
//			"structValue": Structure{},
//		})
//		if err == nil {
//			t.Errorf("expected error for set %d, got nil", i)
//			return
//		}
//		if err.Error() != set.err {
//			t.Errorf("expected\n%v\ngot\n%v", set.err, err)
//		}
//	}
//}
//
//type context map[string]any
//
//func TestChains(t *testing.T) {
//	sets := []struct {
//		string
//		context
//		any
//	}{
//		{"a.b.c", map[string]any{"a": map[string]any{"b": map[string]any{"c": 123}}}, 123},
//		{"some?.value", map[string]any{"some": nil}, nil},
//		{"func?.()", map[string]any{"func": func() int { return 444 }}, 444},
//		{"func()", map[string]any{"func": func() int { return 444 }}, 444},
//		{"func?.()", map[string]any{}, nil},
//		{"arr[1+1]", map[string]any{"arr": []string{"one", "two", "three"}}, "three"},
//	}
//	for i, set := range sets {
//		file := helpers.File{"debug.txt", fmt.Sprintf("{{ %s }}", set.string)}
//		elements, err := Tokenize("debug.txt", fmt.Sprintf("{{ %s }}", set.string))
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//		expr, err := Parse(file, elements[0].(*tokenizer.Mustache).Tokens)
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		compiler := NewCompiler(file, expr.Expr)
//		program, err := compiler.Compile()
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		res, err := NewVM(file, program).Run(set.context)
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			continue
//		}
//
//		if res != set.any {
//			t.Errorf("expected %v, got %v", set.any, res)
//		}
//
//		t.Logf("set %d: pass", i)
//	}
//}
