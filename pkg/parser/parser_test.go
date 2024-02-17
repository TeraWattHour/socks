package parser

import (
	"github.com/terawatthour/socks/internal/debug"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"testing"
)

func TestParserSimple(t *testing.T) {
	template := `
    @template("templates/header.html")
        @define("page")
            nested page
        @enddefine

        @define("message")
			@if(some != nil)
				@for(idx in some)
					{{ idx }}
				@endfor	
			@endif

            Hello from the nested page
        @enddefine
    @endtemplate
`

	elements, err := tokenizer.Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	programs, err := Parse(elements)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	debug.PrintPrograms("TestParserSimple", programs)
}

func TestDependencies(t *testing.T) {
	template := `
@for(idx in someSlice)
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
		{"someSlice", "otherSlice"},
		{"jdx"},
		{"argument1", "argument2", "argument3", "argument4"},
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
			if !helpers.SlicesEqual(program.(*ForStatement).Dependencies, expect[i]) {
				t.Errorf("unexpected result: %v, expected: %v", program.(*ForStatement).Dependencies, expect[i])
			}
		case "if":
			if !helpers.SlicesEqual(program.(*IfStatement).Dependencies, expect[i]) {
				t.Errorf("unexpected result: %v, expected: %v", program.(*IfStatement).Dependencies, expect[i])
			}
		case "expression":
			if !helpers.SlicesEqual(program.(*Expression).Dependencies, expect[i]) {
				t.Errorf("unexpected result: %v, expected: %v", program.(*Expression).Dependencies, expect[i])
			}
		}
		i++
	}
}
