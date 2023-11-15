package tokenizer

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	template := `{% extend "some_template.html" %}<html><head><title>{{ .Title }}</title></head><body><h1>{{ .Title.Format(.Datum, "dddd", 1.2).ToUTC() }} {{ nice_ident }}</h1>{! for i, v in .Table !} {{ v }} {! end !} </body></html>`
	tok := NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	expected := []Tag{
		{
			Tokens: []Token{{
				Kind:    "extend",
				Literal: "extend",
			}, {
				Kind:    "string",
				Literal: "some_template.html",
			}},
		},
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
				Literal: "nice_ident",
			}},
		},
		{
			Tokens: []Token{{
				Kind:    "for",
				Literal: "for",
			}, {
				Kind:    "ident",
				Literal: "i",
			}, {
				Kind:    "comma",
				Literal: ",",
			}, {
				Kind:    "ident",
				Literal: "v",
			}, {
				Kind:    "in",
				Literal: "in",
			}, {
				Kind:    "dot",
				Literal: ".",
			}, {
				Kind:    "ident",
				Literal: "Table",
			}},
		},
		{
			Tokens: []Token{{
				Kind:    "ident",
				Literal: "v",
			}},
		},
		{
			Tokens: []Token{{
				Kind:    "end",
				Literal: "end",
			}},
		},
	}

	for i, block := range expected {
		for j, token := range block.Tokens {
			failed := false
			if token.Kind != tok.Tags[i].Tokens[j].Kind {
				t.Errorf("expected Kind %s, got %s", tok.Tags[i].Tokens[j].Kind, token.Kind)
				failed = true
			}
			if token.Literal != tok.Tags[i].Tokens[j].Literal {
				t.Errorf("expected token %s, got %s", tok.Tags[i].Tokens[j].Literal, token.Literal)
				failed = true
			}
			if failed {
				return
			}
		}
	}
}

func TestTokenizerUnexpectedTerminator(t *testing.T) {
	template := `{{ end %}`
	tok := NewTokenizer(template)
	if err := tok.Tokenize(); err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestTokenizerNiceLetters(t *testing.T) {
	template := "{{ żółć }} {{ ęśąłó }}"
	tok := NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	expected := []Tag{{
		Tokens: []Token{{
			Kind:    "ident",
			Literal: "żółć",
		}},
	}, {
		Tokens: []Token{{
			Kind:    "ident",
			Literal: "ęśąłó",
		}},
	}}

	for i, block := range expected {
		for j, token := range block.Tokens {
			failed := false
			if token.Kind != tok.Tags[i].Tokens[j].Kind {
				t.Errorf("expected Kind %s, got %s", tok.Tags[i].Tokens[j].Kind, token.Kind)
				failed = true
			}
			if token.Literal != tok.Tags[i].Tokens[j].Literal {
				t.Errorf("expected token %s, got %s", tok.Tags[i].Tokens[j].Literal, token.Literal)
				failed = true
			}
			if failed {
				return
			}
		}
	}
}
