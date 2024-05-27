package tokenizer

import (
	"fmt"
	"testing"
)

func TestNumbers(t *testing.T) {
	template := `{{ 2+4.123+0b11+0x123ABC+0o1234567+.2+0o62 }}`
	elements, err := Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(elements))
		return
	}
	if elements[0].Kind() != MustacheKind {
		t.Errorf("expected MustacheKind, got %s", elements[0].Kind())
		return
	}
	fmt.Println(elements[0].(*Mustache).Tokens)
}

func TestStatements(t *testing.T) {
	template := `xd ą@if(1==1)@endif{{ 123 }}@slot("content") default \{{ slot }} {# comm{{ ent #} content @endslot xd {{ {# @enddefine #}`
	elements, err := Tokenize(template)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	expected := []string{"xd ą", "@if(1==1)", "@endif", " 123 ", "@slot(\"content\")", " default \\{{ slot }} ", " content ", "@endslot", " xd "}
	if len(elements) != len(expected) {
		t.Errorf("expected %d element, got %d", len(expected), len(elements))
		return
	}

	for i, e := range expected {
		el := elements[i]
		var lit string
		switch el := el.(type) {
		case *Statement:
			lit = el.Literal
		case *Mustache:
			lit = el.Literal
		case Text:
			lit = string(el)
		}
		if e != lit {
			t.Errorf("(%d) expected %s, got %s", i, e, lit)
			return
		}
	}
}

//func TestAdjacent(t *testing.T) {
//	template := `{{ 2+4 }}{{ 1+2 }}@if(1==1)@endif{{ "hello" }}\{{ žœ{# comment #}{{ 1+2 }}`
//	found := lookupElements(template)
//	var previous []int
//	fmt.Println(template)
//	for _, f := range found {
//		if previous == nil {
//			for i := 0; i < f[0]; i++ {
//				fmt.Print(" ")
//			}
//		} else {
//			for i := previous[1]; i < f[0]; i++ {
//				fmt.Print(" ")
//			}
//		}
//		for i := f[0]; i < f[1]; i++ {
//			fmt.Print("^")
//		}
//		previous = f
//	}
//	fmt.Println()
//	expected := [][]int{{0, 2}, {9, 11}, {18, 21}, {27, 33}, {42, 44}, {54, 56}}
//	if len(found) != len(expected) {
//		t.Errorf("expected %d elements, got %d", len(expected), len(found))
//		return
//	}
//	for i, f := range found {
//		if f[0] != expected[i][0] || f[1] != expected[i][1] {
//			t.Errorf("expected %v, got %v", expected[i], f)
//			return
//		}
//	}
//}
