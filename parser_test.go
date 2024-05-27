package socks

import (
	"github.com/terawatthour/socks/tokenizer"
	"slices"
	"testing"
)

func TestDependencies(t *testing.T) {
	template := `
@for(idx in someSlice with iterator_index)
	@if(test == 1)
		{{ fun.method(argument1, argument2[argument4.method(123)]) + argument3 }}
	@endif

	@for(jdx in otherSlice)
		{{ jdx }}
	@endfor
	
	{{ idx }}
@endfor

{{ independent }}

{{ another }}
`
	elements, err := tokenizer.Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	expect := [][]string{
		{"argument1", "argument2", "argument3", "argument4", "fun", "someSlice", "test", "otherSlice"},
		{"argument1", "argument2", "argument3", "argument4", "fun", "test"},
		{"argument1", "argument2", "argument3", "argument4", "fun"},
		{"otherSlice"},
		{"jdx"},
		{"idx"},
		{"independent"},
		{"another"},
	}

	programs, err := Parse(elements)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	i := 0
	for _, program := range programs {
		if program.Kind() == "text" {
			continue
		}
		switch program.Kind() {
		case "for":
			if !slices.Equal(program.(*ForStatement).dependencies, expect[i]) {
				t.Errorf("unexpected result: %v, expected: %v", program.(*ForStatement).dependencies, expect[i])
			}
		case "if":
			if !slices.Equal(program.(*IfStatement).dependencies, expect[i]) {
				t.Errorf("unexpected result: %v, expected: %v", program.(*IfStatement).dependencies, expect[i])
			}
		case "expression":
			if !slices.Equal(program.(*Expression).dependencies, expect[i]) {
				t.Errorf("unexpected result: %v, expected: %v", program.(*Expression).dependencies, expect[i])
			}
		}
		i++
	}
}
