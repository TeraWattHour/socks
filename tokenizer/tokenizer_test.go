package tokenizer

import (
	"fmt"
	"testing"
)

func TestNumbers(t *testing.T) {
	template := `{{ 2+4.123+0b11+0x123ABC+0o1234567+.2+0o62*076 }}`
	elements, err := Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(elements))
		return
	}
	mustache, ok := elements[0].(*Mustache)
	if !ok {
		t.Errorf("expected MustacheKind")
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
	if len(mustache.Tokens) != len(expected) {
		t.Errorf("expected %d elements, got %d", len(expected), len(mustache.Tokens))
		return
	}
	for i, e := range expected {
		el := mustache.Tokens[i]
		if el.Kind != e.Kind || el.Literal != e.Literal {
			t.Errorf("(%d) expected %v, got %v", i, e, el)
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
		_, err := Tokenize(fmt.Sprintf("{{ %s }}", c))
		if err == nil {
			t.Errorf("(%d) expected error, got nil", i)
			return
		}
	}
}

func TestStatements(t *testing.T) {
	template := `xd ą@if(1==1)@endif{{ 123 }}@slot("content")@if(predicate()) true @endif default \{{ slot }} {# comm{{ ent #} content @endslot xd {# @enddefine #}`
	elements, err := Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	expected := []string{"xd ą", "@if(1==1)", "@endif", " 123 ", "@slot(\"content\")", "@if(predicate())", " true ", "@endif", " default \\{{ slot }} ", " content ", "@endslot", " xd "}
	if len(elements) != len(expected) {
		t.Errorf("expected %d element, got %d", len(expected), len(elements))
		return
	}

	for i, e := range expected {
		el := elements[i]
		var lit string
		switch el := el.(type) {
		case *Statement:
			lit = el.Literal
		case *Mustache:
			lit = el.Literal
		case Text:
			lit = string(el)
		}
		if e != lit {
			t.Errorf("(%d) expected %s, got %s", i, e, lit)
			return
		}
	}
}

func TestExpressionTokenizing(t *testing.T) {

}
