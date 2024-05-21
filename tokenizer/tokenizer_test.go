package tokenizer

import (
	"testing"
)

func TestLookup(t *testing.T) {
	template := `Ł{{ 2+4 }} @if(1==1){{ "hello" }} @endif \{{ žœ{# comment #}`
	found := lookupElements(template)
	expected := [][]int{{2, 4}, {12, 15}, {21, 23}, {35, 41}, {50, 52}}
	if len(found) != len(expected) {
		t.Errorf("expected %d elements, got %d", len(expected), len(found))
		return
	}
	for i, f := range found {
		if f[0] != expected[i][0] || f[1] != expected[i][1] {
			t.Errorf("expected %v, got %v", expected[i], f)
			return
		}
	}
}

func TestTokenize(t *testing.T) {
	template := `ł{{ 'hello' + 'he\'s' }} 你能肯定吗？ Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras imperdiet imperdiet orci, non cursus metus iaculis ac. Duis fermentum, diam ac luctus mollis, turpis felis suscipit nulla, vitae volutpat nisi diam sagittis dui. Fusce quis nibh eget tortor pulvinar rutrum. Nunc eget ante congue tortor iaculis congue sed quis mi. Sed vitae varius justo. Suspendisse ac elementum mi. Integer nec malesuada leo. Phasellus fermentum varius laoreet. Cras diam dui, congue eu auctor gravida, euismod sit amet magna. Nullam fringilla dolor non enim lacinia fermentum. Mauris dapibus faucibus rhoncus. Nulla sapien dui, volutpat eget tellus in, bibendum feugiat ante. Quisque non venenatis felis.
{{ 10 }} @if(1==1) @endif aaa {! 2+2 !}{# hello from the comment tag #}`
	elements, err := Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	for _, e := range elements {
		t.Log(e)
	}
}

func BenchmarkTokenize(b *testing.B) {
	template := `ł{{ 'hello' + 'he\'s' }} 你能肯定吗？ Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras imperdiet imperdiet orci, non cursus metus iaculis ac. Duis fermentum, diam ac luctus mollis, turpis felis suscipit nulla, vitae volutpat nisi diam sagittis dui. Fusce quis nibh eget tortor pulvinar rutrum. Nunc eget ante congue tortor iaculis congue sed quis mi. Sed vitae varius justo. Suspendisse ac elementum mi. Integer nec malesuada leo. Phasellus fermentum varius laoreet. Cras diam dui, congue eu auctor gravida, euismod sit amet magna. Nullam fringilla dolor non enim lacinia fermentum. Mauris dapibus faucibus rhoncus. Nulla sapien dui, volutpat eget tellus in, bibendum feugiat ante. Quisque non venenatis felis.
{{ 10 }} @if(1==1) @endif aaa {! 2+2 !}{# hello from the comment tag #}`
	for i := 0; i < b.N; i++ {
		_, _ = Tokenize(template)
	}
}
