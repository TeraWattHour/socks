package parser

import (
	"github.com/terawatthour/socks/pkg/tokenizer"
	"testing"
)

func TestParserSimple(t *testing.T) {
	template := `<html><head><title>{% .Title %}</title></head><body><h1>{% .Format("najs kok", .Title()) %}</h1></body></html>`
	tok := tokenizer.NewTokenizer(template)
	tok.Tokenize()

	p := NewParser(tok)
	p.Parse()
}
