package socks

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/tokenizer"
	"testing"
)

func TestMalformed(t *testing.T) {
	sets := []struct {
		template string
		expect   string
	}{
		{
			`@if(1==1) @endfor`,
			"  ┌─ debug.txt:1:11:\n1 | @if(1==1) @endfor␄\n  |           ^^^^^^^\nunexpected `@endfor`, expected `@endif`",
		}, {
			"@for(a bin A) @endfor",
			"  ┌─ debug.txt:1:8:\n1 | @for(a bin A) @endfor␄\n  |        ^^^\nunexpected identifier, expected \"in\"",
		}, {
			"@for(in A)",
			"  ┌─ debug.txt:1:6:\n1 | @for(in A)␄\n  |      ^^\nunexpected \"in\", expected identifier",
		}, {
			"@for(a)",
			"  ┌─ debug.txt:1:7:\n1 | @for(a)␄\n  |       ^\nunexpected end of statement, expected \"in\"",
		}, {
			"@for(a in)",
			"  ┌─ debug.txt:1:10:\n1 | @for(a in)␄\n  |          ^\nunexpected end of statement, expected expression",
		}, {
			"@for(a in A",
			"  ┌─ debug.txt:1:1:\n1 | @for(a in A␄\n  | ^^^^^^^^^^^^\nunclosed tag",
		}, {
			"@for(a in with)",
			"  ┌─ debug.txt:1:10:\n1 | @for(a in with)␄\n  |          ^\nexpected expression",
		}, {
			"@for(a in A with)",
			"  ┌─ debug.txt:1:17:\n1 | @for(a in A with)␄\n  |                 ^\nunexpected end of statement, expected identifier",
		}, {
			"@for(a in A with b c)",
			"  ┌─ debug.txt:1:20:\n1 | @for(a in A with b c)␄\n  |                    ^\nunexpected identifier, expected end of statement",
		}, {
			"@if(1==1) abc @else @elif(1==2) @endif",
			"  ┌─ debug.txt:1:21:\n1 | @if(1==1) abc @else @elif(1==2) @endif␄\n  |                     ^^^^^^^^^^^\nunexpected `@elif` after `@else`",
		}, {
			"@for(a in A) @else @endfor",
			"  ┌─ debug.txt:1:14:\n1 | @for(a in A) @else @endfor␄\n  |              ^^^^^\nunexpected `@else` outside if statement",
		}, {
			"@elif(1==1)",
			"  ┌─ debug.txt:1:1:\n1 | @elif(1==1)␄\n  | ^^^^^^^^^^^\nunexpected `@elif` outside if statement",
		}, {
			"@endif",
			"  ┌─ debug.txt:1:1:\n1 | @endif␄\n  | ^^^^^^\nunexpected end tag",
		}, {
			"@if(1==1) @enddefine",
			"  ┌─ debug.txt:1:11:\n1 | @if(1==1) @enddefine␄\n  |           ^^^^^^^^^^\nunexpected `@enddefine`, expected `@endif`",
		},
	}
	for i, set := range sets {
		elements, err := tokenizer.Tokenize("debug.txt", set.template)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		_, err = Parse(helpers.File{"debug.txt", set.template}, elements)
		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}
		if err.Error() != set.expect {
			fmt.Println(err)
			t.Errorf("(%d) expected `%s`, got `%s`", i, set.expect, err)
		}
	}
}
