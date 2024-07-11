package expression

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
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

//func TestVM_Run(t *testing.T) {
//	sets := []struct {
//		expr   string
//		expect any
//	}{{
//		"ordinals[nil ?: 1]",
//		"2nd",
//	}, {
//		"voidMember?.method() ?: 1 != 1 ?: 420",
//		false,
//	}, {
//		"voidMember?.property",
//		nil,
//	}, {
//		"2 ** 3 / 4",
//		2,
//	}, {
//		`not "str" in [false]`,
//		true,
//	}, {
//		`base.structure.Method(123.4) + base.structure.ReceiverMethod() + someInt.Method()`,
//		`the ratio is 123.4 non-pointer method value of SomeInt is 123`,
//	}, {
//		`sprintf("%.2f", 12 ? 12. + 123. ** 2.5 : 0.123)`,
//		"167800.73",
//	}}
//
//	for i, set := range sets {
//		elements, err := tokenizer.Tokenize(fmt.Sprintf("{{ %s }}", set.expr))
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//		expr, err := Parse(elements[0].(*tokenizer.Mustache).Tokens)
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		compiler := NewCompiler(expr.Expr)
//		chunk, err := compiler.Compile()
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//
//		vm := NewVM(chunk)
//		result, err := vm.Run(map[string]any{
//			"ordinals": []string{"1st", "2nd"},
//			"base": map[string]any{
//				"structure": &Structure{},
//			},
//			"someInt": SomeInt(123),
//			"sprintf": fmt.Sprintf,
//		})
//		if err != nil {
//			t.Errorf("unexpected error for set %d: %v", i, err)
//			return
//		}
//		if result != set.expect {
//			t.Errorf("expected %v, got %v", set.expect, result)
//		}
//	}
//}

func TestVM_Errors(t *testing.T) {
	sets := []struct {
		expr string
		err  string
	}{
		{
			`someNilValue.accessed`,
			"  ┌─ debug.txt:1:4:\n1 | {{ someNilValue.accessed }}␄\n  |    ^^^^^^^^^^^^^^^^^^^^^\ncannot access properties of <nil>",
		}, {
			"arrayValue[\"wrong_index\"]",
			"  ┌─ debug.txt:1:15:\n1 | {{ arrayValue[\"wrong_index\"] }}␄\n  |               ^^^^^^^^^^^^^\nforbidden array index access, cannot cast string to int",
		}, {
			"structValue[123]",
			"  ┌─ debug.txt:1:16:\n1 | {{ structValue[123] }}␄\n  |                ^^^\nstruct field accessor must be of type string, got int",
		}, {
			"(1 + 2)[3]",
			"  ┌─ debug.txt:1:5:\n1 | {{ (1 + 2)[3] }}␄\n  |     ^^^^^\nexpected array, struct or map, got int",
		}, {
			"someNilValue . method()",
			"  ┌─ debug.txt:1:4:\n1 | {{ someNilValue . method() }}␄\n  |    ^^^^^^^^^^^^^^^^^^^^^\ncannot access properties of <nil>",
		}, {
			"someNilValue()",
			"  ┌─ debug.txt:1:4:\n1 | {{ someNilValue() }}␄\n  |    ^^^^^^^^^^^^\nexpected function, got <nil>",
		}, {
			"float64(\"string\")",
			"  ┌─ debug.txt:1:12:\n1 | {{ float64(\"string\") }}␄\n  |            ^^^^^^^^\ncannot cast string to float64",
		}, {
			"range(2, 1, 1)",
			"  ┌─ debug.txt:1:10:\n1 | {{ range(2, 1, 1) }}␄\n  |          ^^^^^^^\nstep cannot be positive while start > end",
		},
	}

	for i, set := range sets {
		file := helpers.File{"debug.txt", fmt.Sprintf("{{ %s }}", set.expr)}
		elements, err := tokenizer.Tokenize("debug.txt", fmt.Sprintf("{{ %s }}", set.expr))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		expr, err := Parse(file, elements[0].(*tokenizer.Mustache).Tokens)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		compiler := NewCompiler(file, expr.Expr)
		chunk, err := compiler.Compile()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		_, err = NewVM(file, chunk).Run(map[string]any{
			"arrayValue":  []int{1, 2, 3, 4},
			"structValue": Structure{},
		})
		if err == nil {
			t.Errorf("expected error for set %d, got nil", i)
			return
		}
		if err.Error() != set.err {
			fmt.Println(err)
			t.Errorf("expected %v, got %v", set.err, err)
		}
	}
}

type context map[string]any

func TestChains(t *testing.T) {
	sets := []struct {
		string
		context
	}{
		{"some?.value", map[string]any{"some": nil}},
		{"some.value", map[string]any{"some": 123}},
	}
	for _, set := range sets {
		file := helpers.File{"debug.txt", fmt.Sprintf("{{ %s }}", set.string)}
		elements, err := tokenizer.Tokenize("debug.txt", fmt.Sprintf("{{ %s }}", set.string))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		expr, err := Parse(file, elements[0].(*tokenizer.Mustache).Tokens)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		compiler := NewCompiler(file, expr.Expr)
		chunk, err := compiler.Compile()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		dumpChunk(chunk)

		res, err := NewVM(file, chunk).Run(set.context)
		fmt.Println(res, err)
	}
}
