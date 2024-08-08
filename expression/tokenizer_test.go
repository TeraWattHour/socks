package expression

import (
	"github.com/terawatthour/socks/internal/helpers"
	"testing"
)

func TestNumbers(t *testing.T) {
	template := `2 + 4.123 + 0b11 + 0x123ABC + 0o1234567 + .2 + 0o62 * 076`
	tokens, err := Tokenize(template, helpers.Location{Line: 1, Column: 1})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	expected := []Token{
		{Kind: TokNumeric, Literal: "2"},
		{Kind: TokPlus, Literal: "+"},
		{Kind: TokNumeric, Literal: "4.123"},
		{Kind: TokPlus, Literal: "+"},
		{Kind: TokNumeric, Literal: "0b11"},
		{Kind: TokPlus, Literal: "+"},
		{Kind: TokNumeric, Literal: "0x123ABC"},
		{Kind: TokPlus, Literal: "+"},
		{Kind: TokNumeric, Literal: "0o1234567"},
		{Kind: TokPlus, Literal: "+"},
		{Kind: TokNumeric, Literal: ".2"},
		{Kind: TokPlus, Literal: "+"},
		{Kind: TokNumeric, Literal: "0o62"},
		{Kind: TokAsterisk, Literal: "*"},
		{Kind: TokNumeric, Literal: "076"},
	}
	if len(tokens) != len(expected) {
		t.Errorf("expected %d tokens, got %d", len(expected), len(tokens))
		return
	}
	for i, e := range expected {
		el := tokens[i]
		if el.Kind != e.Kind || el.Literal != e.Literal {
			t.Errorf("token %d: expected %v, got %v", i, e, el)
			return
		}
	}
}

func TestMalformedNumbers(t *testing.T) {
	cases := []string{
		"2.3.4",
		"0b112",
		"0x2G_",
		"0x_2",
		"0o2_",
		"0._1123",
		"0923456789",
		"0x99.23",
	}
	for i, c := range cases {
		_, err := Tokenize(c, helpers.Location{Line: 1, Column: 1})
		if err == nil {
			t.Errorf("(%d) expected error, got nil", i)
			return
		}
	}
}

//
//func TestErrorReporting(t *testing.T) {
//	sets := []struct {
//		template string
//		expect   string
//	}{
//		{
//			`	{#      a`,
//			"  ┌─ debug.txt:1:11:\n1 | \t{#      a␄\n  | \t         ^\nunexpected EOF, unclosed comment",
//		}, {
//			"\na\n|\n @if ",
//			"  ┌─ debug.txt:4:2:\n4 |  @if ␄\n  |  ^^^^\nunexpected EOF, expected `(` after statement",
//		}, {
//			"{{ ",
//			"  ┌─ debug.txt:1:4:\n1 | {{ ␄\n  |    ^\nunexpected EOF",
//		}, {
//			"@if ([))",
//			"  ┌─ debug.txt:1:7:\n1 | @if ([))␄\n  |       ^\nunexpected `)`, as it closes `[`",
//		}, {
//			"\n{{) }}",
//			"  ┌─ debug.txt:2:3:\n2 | {{) }}␄\n  |   ^\nunexpected `)`",
//		}, {
//			"\n{{(  ]}}",
//			"  ┌─ debug.txt:2:6:\n2 | {{(  ]}}␄\n  |      ^\nunexpected `]`, as it closes `(`",
//		}, {
//			"\n@if(]",
//			"  ┌─ debug.txt:2:5:\n2 | @if(]␄\n  |     ^\nunexpected `]`",
//		}, {
//			"{{ \"unclosed ",
//			"  ┌─ debug.txt:1:14:\n1 | {{ \"unclosed ␄\n  |              ^\nunexpected EOF, unclosed string",
//		}, {
//			"{{ $ }}",
//			"  ┌─ debug.txt:1:4:\n1 | {{ $ }}␄\n  |    ^\nunexpected token: `$`",
//		}, {
//			"{{ good }}zb {{ aaa !}zb",
//			"  ┌─ debug.txt:1:21:\n1 | {{ good }}zb {{ aaa !}zb␄\n  |                     ^^\nexpected `}}` to close mustache",
//		}, {
//			"{{ }}",
//			"  ┌─ debug.txt:1:1:\n1 | {{ }}␄\n  | ^^^^^\nempty statement",
//		}, {
//			"@if()",
//			"  ┌─ debug.txt:1:1:\n1 | @if()␄\n  | ^^^^^\nempty statement",
//		},
//	}
//	for i, set := range sets {
//		_, err := Tokenize("debug.txt", set.template)
//		if err == nil {
//			t.Errorf("(%d) expected error, got nil", i)
//			return
//		}
//		if err.Error() != set.expect {
//			fmt.Println(err)
//			t.Errorf("(%d) expected %q, got %q", i, set.expect, err.Error())
//		}
//	}
//}
//
//func TestTokenLength(t *testing.T) {
//	sets := []struct {
//		token  string
//		expect int
//	}{
//		{
//			"\"string_length\"",
//			15,
//		}, {
//			".",
//			1,
//		}, {
//			"2.3",
//			3,
//		}, {
//			"0b101",
//			5,
//		}, {
//			"identifier",
//			10,
//		},
//	}
//	for i, set := range sets {
//		elements, err := Tokenize("debug.txt", fmt.Sprintf("{{ %s }}", set.token))
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//			return
//		}
//		if len(elements) != 1 {
//			t.Errorf("expected 1 element, got %d", len(elements))
//			return
//		}
//		mustache, ok := elements[0].(*Mustache)
//		if !ok {
//			t.Errorf("expected MustacheKind")
//			return
//		}
//		if mustache.Tokens[0].Location.Length != set.expect {
//			t.Errorf("(%d) expected %d, got %d", i, set.expect, mustache.Tokens[0].Location.Length)
//		}
//	}
//
//}
//
//func TestTokenLocation(t *testing.T) {
//
//	tokens, err := Tokenize("debug.txt", "{{ \"wrong_index\" }}")
//	if err != nil {
//		t.Errorf("unexpected error: %v", err)
//		return
//	}
//
//	fmt.Println(errors2.New("eee", "debug.txt", "{{ \"wrong_index\" }}", tokens[0].(*Mustache).Tokens[0].Location, tokens[0].(*Mustache).Tokens[0].Location.FromOther()).Error())
//}
