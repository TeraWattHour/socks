package tokenizer

import (
	"fmt"
	"testing"
)

// func TestTokenizer(t *testing.T) {
// 	template := `{% extend "some_template.html"%}<html><head><title>{{ Title }}</title></head><body><h1>{{ Title.Format(.Datum, "dddd", 1).ToUTC()}} {{ nice_ident }}</h1>{! for i, v in .Table !} {{ v }} {! end !} </body></html>{{ xdd }}`
// 	tok := NewTokenizer(template))
// 	if err := tok.Tokenize(); err != nil {
// 		t.Errorf("unexpected error: %s", err)
// 		return
// 	}

// 	expected := []Mustache{
// 		{
// 			tokens: []Token{{
// 				Type:    "extend",
// 				Literal: "extend",
// 			}, {
// 				Type:    "string",
// 				Literal: "some_template.html",
// 			}},
// 		},
// 		{
// 			tokens: []Token{{
// 				Type:    "ident",
// 				Literal: "Title",
// 			}},
// 		},
// 		{
// 			tokens: []Token{{
// 				Type:    "ident",
// 				Literal: "Title",
// 			}, {
// 				Type:    "unknown",
// 				Literal: ".",
// 			}, {
// 				Type:    "ident",
// 				Literal: "Format",
// 			}, {
// 				Type:    "unknown",
// 				Literal: "(",
// 			}, {
// 				Type:    "unknown",
// 				Literal: ".",
// 			}, {
// 				Type:    "ident",
// 				Literal: "Datum",
// 			}, {
// 				Type:    "comma",
// 				Literal: ",",
// 			}, {
// 				Type:    "string",
// 				Literal: "dddd",
// 			}, {
// 				Type:    "comma",
// 				Literal: ",",
// 			}, {
// 				Type:    "unknown",
// 				Literal: "1",
// 			}, {
// 				Type:    "unknown",
// 				Literal: ")",
// 			}, {
// 				Type:    "unknown",
// 				Literal: ".",
// 			}, {
// 				Type:    "ident",
// 				Literal: "ToUTC",
// 			}, {
// 				Type:    "unknown",
// 				Literal: "(",
// 			}, {
// 				Type:    "unknown",
// 				Literal: ")",
// 			}},
// 		},
// 		{
// 			tokens: []Token{{
// 				Type:    "ident",
// 				Literal: "nice_ident",
// 			}},
// 		},
// 		{
// 			tokens: []Token{{
// 				Type:    "for",
// 				Literal: "for",
// 			}, {
// 				Type:    "ident",
// 				Literal: "i",
// 			}, {
// 				Type:    "comma",
// 				Literal: ",",
// 			}, {
// 				Type:    "ident",
// 				Literal: "v",
// 			}, {
// 				Type:    "in",
// 				Literal: "in",
// 			}, {
// 				Type:    "unknown",
// 				Literal: ".",
// 			}, {
// 				Type:    "ident",
// 				Literal: "Table",
// 			}},
// 		},
// 		{
// 			tokens: []Token{{
// 				Type:    "ident",
// 				Literal: "v",
// 			}},
// 		},
// 		{
// 			tokens: []Token{{
// 				Type:    "end",
// 				Literal: "end",
// 			}},
// 		},
// 		{
// 			tokens: []Token{{
// 				Type:    "ident",
// 				Literal: "xdd",
// 			}},
// 		},
// 	}

// 	for i, block := range expected {
// 		for j, token := range block.tokens {
// 			failed := false
// 			if token.Type != tok.Tags[i].tokens[j].Type {
// 				t.Errorf("expected Type %s, got %s", tok.Tags[i].tokens[j].Type, token.Type)
// 				failed = true
// 			}
// 			if token.Literal != tok.Tags[i].tokens[j].Literal {
// 				t.Errorf("expected token %s, got %s", tok.Tags[i].tokens[j].Literal, token.Literal)
// 				failed = true
// 			}
// 			if failed {
// 				return
// 			}
// 		}
// 	}
// }

// func TestTokenizerWhitespace(t *testing.T) {
// 	template := `{{ v }}`
// 	tok := NewTokenizer([]rune(template))
// 	if err := tok.Tokenize(); err != nil {
// 		t.Errorf("unexpected error: %s", err)
// 		return
// 	}
// 	if tok.Tags[0].tokens[0].Literal != "v" {
// 		t.Errorf("expected token %s, got %s", "v", tok.Tags[0].tokens[0].Literal)
// 	}
// }

// func TestTokenizerUnexpectedTerminator(t *testing.T) {
// 	template := `{{ end %}`
// 	tok := NewTokenizer([]rune(template))
// 	if err := tok.Tokenize(); err == nil {
// 		t.Errorf("expected error, got nil")
// 		return
// 	}
// }

// func TestTokenizerNiceLetters(t *testing.T) {
// 	template := `{{ żółć }} {{ ęśąłó }}`
// 	tok := NewTokenizer([]rune(template))
// 	if err := tok.Tokenize(); err != nil {
// 		t.Errorf("unexpected error: %s", err)
// 		return
// 	}

// 	expected := []Mustache{{
// 		tokens: []Token{{
// 			Type:    "ident",
// 			Literal: "żółć",
// 		}},
// 	}, {
// 		tokens: []Token{{
// 			Type:    "ident",
// 			Literal: "ęśąłó",
// 		}},
// 	}}
// 	fmt.Println(tok.Tags)
// 	for i, block := range expected {
// 		for j, token := range block.tokens {
// 			failed := false
// 			if token.Type != tok.Tags[i].tokens[j].Type {
// 				t.Errorf("expected kind %s, got %s", tok.Tags[i].tokens[j].Type, token.Type)
// 				failed = true
// 			}
// 			if token.Literal != tok.Tags[i].tokens[j].Literal {
// 				t.Errorf("expected token %s, got %s", tok.Tags[i].tokens[j].Literal, token.Literal)
// 				failed = true
// 			}
// 			if failed {
// 				return
// 			}
// 		}
// 	}
// }

// func TestEofError(t *testing.T) {
// 	template := `nice {{ `
// 	tok := NewTokenizer([]rune(template))
// 	err := tok.Tokenize()
// 	if err == nil {
// 		t.Errorf("expected error, got nil")
// 		return
// 	}
// 	lit := err.Error()
// 	fmt.Println(lit)
// }

func TestTokenization(t *testing.T) {
	template := `
@extend("base.html")

@define("content")
    @template("templates/header.html")
    
    @endtemplate

	<style>
		@import url(some font)
	</style>


    <p>Hello from the {{ Server }} server</p>

    @for[nostatic](phrase, i in Phrases)
    <div>
        <p>
            @template("templates/number.html") @define("number"){{ i + 1 }}@enddefine@endtemplate
            : {{ phrase.Content }}</p>
        @if(i > 0)
        <p>Previous {{ i }}: {{ Phrases[i-1].Content }}</p>
        @endif
    </div>
    @endfor
@enddefine`
	elements, err := Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	for _, element := range elements {
		fmt.Println(element)
	}
}
