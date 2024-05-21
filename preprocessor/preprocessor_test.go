package preprocessor

import (
	"github.com/terawatthour/socks/internal/debug"
	"testing"
)

func TestPreprocessor(t *testing.T) {
	embedded := `<h1>@slot("title")Fallback@endslot card</h1>`
	parent := `@extend("superparent.html") @define("content") <head><title>@slot("title")Fallback title@endslot </title></head> @slot("content") fallback content @endslot @enddefine`
	child := `@extend("parent.html") @define("content") @if(get == "good") @for(val in values) @template("embedded.html") @define("title") Home page @enddefine @endtemplate @endfor @if(something>10) @for(i in range(1, 4))<p>{{ i }}</p>@endfor @endif @endif @enddefine`
	superparent := `<html> aaaaa @slot("content") xd @endslot </html>`
	result, err := New(map[string]string{"parent.html": parent, "child.html": child, "embedded.html": embedded, "superparent.html": superparent}, map[string]string{
		"parent.html":      "parent.html",
		"child.html":       "child.html",
		"embedded.html":    "embedded.html",
		"superparent.html": "superparent.html",
	}, map[string]any{
		"values": []string{"one", "two", "three"},
		"Menus":  []string{"home", "about", "contact"},
	}, nil).Preprocess("child.html", false)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	debug.PrintPrograms("child.html", result)
}
