package preprocessor

import (
	"github.com/terawatthour/socks/pkg/parser"
	"testing"
)

func TestPreprocessor(t *testing.T) {
	embedded := `<h1>@slot("title")Fallback@endslot title</h1>`
	parent := `<html><head><title>@slot("title")Fallback title@endslot</title></head> @slot("content") fallback content @endslot </html>`
	child := `@extend("base.html") @define("content") @if(get == "good") @for(val in values) @template("embedded.html") @define("title") Home page @enddefine @define("footer") @for(menu in Menus) {{ menu }} @endfor @enddefine @endtemplate @endfor @if(something>10) @for(i in range(1, 4))<p>{{ i }}</p>@endfor @endif @endif @enddefine`
	result, err := New(map[string]string{"base.html": parent, "child.html": child, "embedded.html": embedded}, map[string]any{
		"values": []string{"one", "two", "three"},
		"Menus":  []string{"home", "about", "contact"},
	}, nil).Preprocess("child.html", false)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	parser.PrintPrograms("child.html", result)
}
