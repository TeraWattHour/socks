package tokenizer

import (
	"testing"
)

func TestTokenizePrintTags(t *testing.T) {
	template := `<html><head><title>{{ .Title }}</title></head><body><h1>{{ .Title.Format(.Datum, "dddd", 1.2).ToUTC() }}{{ najs }}</h1></body></html>`
	tok := NewTokenizer(template)
	tok.Tokenize()

	expected := []Tag{
		{
			Tokens: []Token{{
				Kind:    "dot",
				Literal: ".",
			}, {
				Kind:    "ident",
				Literal: "Title",
			}},
		},
		{
			Tokens: []Token{{
				Kind:    "dot",
				Literal: ".",
			}, {
				Kind:    "ident",
				Literal: "Title",
			}, {
				Kind:    "dot",
				Literal: ".",
			}, {
				Kind:    "ident",
				Literal: "Format",
			}, {
				Kind:    "lparen",
				Literal: "(",
			}, {
				Kind:    "dot",
				Literal: ".",
			}, {
				Kind:    "ident",
				Literal: "Datum",
			}, {
				Kind:    "comma",
				Literal: ",",
			}, {
				Kind:    "string",
				Literal: "dddd",
			}, {
				Kind:    "comma",
				Literal: ",",
			}, {
				Kind:    "float",
				Literal: "1.2",
			}, {
				Kind:    "rparen",
				Literal: ")",
			}, {
				Kind:    "dot",
				Literal: ".",
			}, {
				Kind:    "ident",
				Literal: "ToUTC",
			}, {
				Kind:    "lparen",
				Literal: "(",
			}, {
				Kind:    "rparen",
				Literal: ")",
			}},
		},
		{
			Tokens: []Token{{
				Kind:    "ident",
				Literal: "najs",
			}},
		},
	}

outer:
	for i, block := range tok.Tags {
		for j, token := range block.Tokens {
			failed := false
			if token.Kind != expected[i].Tokens[j].Kind {
				t.Errorf("expected Kind %s, got %s", expected[i].Tokens[j].Kind, token.Kind)
				failed = true
			}
			if token.Literal != expected[i].Tokens[j].Literal {
				t.Errorf("expected token %s, got %s", expected[i].Tokens[j].Literal, token.Literal)
				failed = true
			}
			if failed {
				break outer
			}
		}
	}
}
