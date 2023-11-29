package parser

import (
	"github.com/terawatthour/socks/pkg/tokenizer"
	"testing"
)

func TestParserSimple(t *testing.T) {
	template := `<html><head><title>{{ .Title }}</title></head><body><h1>{{ .Format("najs kok", .Title()) }} {% template "xd" %} {% define "content" %} defined content {% end %} {% end %}  </h1></body></html>`
	tok := tokenizer.NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	p := NewParser(tok)
	if err := p.Parse(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestParseIfStatement(t *testing.T) {
	template := `{! if Title !} <h1>{{ Title }}</h1> {! end !}`
	tok := tokenizer.NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	p := NewParser(tok)
	if err := p.Parse(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
