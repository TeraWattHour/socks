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

	elements, err := tokenizer.Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	programs, err := Parse(elements)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	PrintPrograms("TestParserSimple", programs)
}
