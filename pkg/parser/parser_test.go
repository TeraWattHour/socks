package parser

import (
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
            Hello from the nested page
        @enddefine
    @endtemplate
`

	tok := tokenizer.NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	p := NewParser(tok)
	programs, err := p.Parse()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	PrintPrograms(programs)
}

func TestParserPrint(t *testing.T) {
	template := ` {{ Title }} `
	tok := tokenizer.NewTokenizer(template)
	if err := tok.Tokenize(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	p := NewParser(tok)
	if _, err := p.Parse(); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
