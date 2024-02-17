package tokenizer

import (
	"fmt"
	"testing"
)

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
