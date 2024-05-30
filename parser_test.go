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
			`{{ 1! }}`,
			"  ┌─ debug.txt:1:5:\n1 | {{ 1! }}␄\n  |     ^\nunexpected token `!`",
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
