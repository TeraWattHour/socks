package html

import (
	"fmt"
	"github.com/terawatthour/socks/expression"
	"github.com/terawatthour/socks/runtime"
	"io"
	"regexp"
	"slices"
	"strings"
)

func Parse(file io.Reader) ([]runtime.Statement, error) {
	elements, err := Tokenize(file)
	if err != nil {
		return nil, err
	}

	return parseBlock(elements)
}

func parseBlock(block []Node) ([]runtime.Statement, error) {
	var output []runtime.Statement
	for _, e := range block {
		switch t := e.(type) {
		case *Text:
			if t.IsComment {
				output = append(output, &runtime.Text{t.Content})
				continue
			}

			parsed, err := parseText(t)
			if err != nil {
				return nil, err
			}

			output = append(output, parsed...)
		case *Tag:
			outlet := &output

			// conditions are always evaluated first
			if value, ok := t.Attributes[":if"]; ok {
				vm, err := expression.Create(value)
				if err != nil {
					return nil, err
				}

				_if := &runtime.IfStatement{Program: vm}
				*outlet = append(*outlet, _if)
				outlet = &_if.Consequence
			} else if value, ok := t.Attributes[":elif"]; ok {
				vm, err := expression.Create(value)
				if err != nil {
					return nil, err
				}
				_if := placeElse(&output)
				if _if == nil {
					return nil, fmt.Errorf("unexpected `:elif` outside if statement")
				}
				_elif := &runtime.ElifBranch{Condition: vm}
				_if.Alternatives = append(_if.Alternatives, _elif)
				outlet = &_elif.Consequence
			} else if _, ok := t.Attributes[":else"]; ok {
				_if := placeElse(&output)
				if _if == nil {
					return nil, fmt.Errorf("unexpected `:else` outside if statement")
				}
				outlet = &_if.Divergent
			}
			if value, ok := t.Attributes[":for"]; ok {
				pattern := `(?P<value>\w+)(,\s*(?P<key>\w+))?\s+in\s+(?P<iterable>.+)$`

				re := regexp.MustCompile(pattern)
				match := re.FindStringSubmatch(value)
				groupNames := re.SubexpNames()
				groupMap := make(map[string]string)
				for i, name := range groupNames {
					if i > 0 && name != "" {
						groupMap[name] = match[i]
					}
				}

				if groupMap["value"] == "" || groupMap["iterable"] == "" {
					return nil, fmt.Errorf("invalid `:for` syntax")
				}

				vm, err := expression.Create(groupMap["iterable"])
				if err != nil {
					return nil, err
				}

				_for := &runtime.ForStatement{Iterable: vm, ValueName: groupMap["value"], KeyName: groupMap["key"]}
				*outlet = append(*outlet, _for)
				outlet = &_for.Body
			}

			if value, ok := t.Attributes[":slot"]; ok {
				slot := &runtime.Slot{Name: value}
				*outlet = append(*outlet, slot)
				outlet = &slot.Children
			}

			// void and self-closing (for interoperability with svg) elements can't have children
			if slices.Contains(voidElements, t.Name) || t.IsSelfClosing && t.Name != "v-slot" && t.Name != "v-component" {
				if err := renderStartTag(t, outlet); err != nil {
					return nil, err
				}

				continue
			}

			block, err := parseBlock(t.Children)
			if err != nil {
				return nil, err
			}

			if t.Name == "v-slot" {
				slot := &runtime.Slot{
					Name:     t.Attributes["name"],
					Children: block,
				}

				if slot.Name == "" {
					return nil, fmt.Errorf("slot name is required")
				}

				*outlet = append(*outlet, slot)
				continue
			}

			if t.Name == "v-component" {
				component := &runtime.Component{
					Name:    t.Attributes["name"],
					Defines: make(map[string][]runtime.Statement),
				}

				if component.Name == "" {
					return nil, fmt.Errorf("component name is required")
				}

				for _, c := range block {
					switch c := c.(type) {
					case *runtime.Slot:
						if _, ok := component.Defines[c.Name]; ok {
							return nil, fmt.Errorf("slot `%s` is already defined", c.Name)
						}
						component.Defines[c.Name] = c.Children
					default:
						if t, ok := c.(*runtime.Text); ok && (strings.TrimSpace(t.Content) == "" || (strings.HasPrefix(t.Content, "<!--") && strings.HasSuffix(t.Content, "-->"))) {
							continue
						}
						return nil, fmt.Errorf("unexpected element in component, only slots are allowed")
					}
				}

				*outlet = append(*outlet, component)
				continue
			}

			if err := renderStartTag(t, outlet); err != nil {
				return nil, err
			}

			*outlet = append(*outlet, block...)
			*outlet = append(*outlet, &runtime.Text{Content: fmt.Sprintf("</%s>", t.Name)})
		}
	}

	return output, nil
}

func parseText(text *Text) (output []runtime.Statement, err error) {
	lastClosed := 0

outer:
	for i := 0; i < len(text.Content)-1; i++ {
		if text.Content[i] == '{' && text.Content[i+1] == '{' && (i == 0 || text.Content[i-1] != '\\') {
			if i > 0 {
				content := text.Content[:i]
				if !text.IsRaw {
					content = escape(content)
				}
				output = append(output, &runtime.Text{Content: content})
			}

			start := i + 2
			for i = start; i < len(text.Content)-1; i++ {
				var stringCharacter uint8 = 0
				if text.Content[i] == '"' && (i == 0 || text.Content[i-1] != '\\') {
					stringCharacter = '"'
				} else if text.Content[i] == '\'' && (i == 0 || text.Content[i-1] != '\\') {
					stringCharacter = '\''
				}

				if stringCharacter != 0 {
					i++
					for i < len(text.Content) && text.Content[i] != stringCharacter {
						i++
					}

					if i == len(text.Content) {
						return nil, fmt.Errorf("unclosed string literal")
					}
				}

				if i < len(text.Content)-1 && text.Content[i] == '}' && text.Content[i+1] == '}' {
					vm, err := expression.Create(text.Content[start:i])
					if err != nil {
						return nil, err
					}

					output = append(output, &runtime.Expression{Program: vm})
					lastClosed = i + 2
					continue outer
				}
			}

			return nil, fmt.Errorf("unclosed expression")
		}
	}

	output = append(output, &runtime.Text{Content: text.Content[lastClosed:]})
	return
}

func isEmptyText(node runtime.Statement) bool {
	if text, ok := node.(*runtime.Text); ok {
		return strings.TrimSpace(text.Content) == ""
	}
	return false
}

func placeElse(s *[]runtime.Statement) *runtime.IfStatement {
	if s == nil || len(*s) == 0 {
		return nil
	}

	previous := (*s)[len(*s)-1]
	if isEmptyText(previous) {
		if len(*s) > 1 {
			*s = (*s)[:len(*s)-1]
			previous = (*s)[len(*s)-1]
		} else {
			return nil
		}
	}

	if el, ok := previous.(*runtime.IfStatement); ok {
		return el
	}

	return nil
}

// voidAttributes are attributes that aren't outputted when rendered
var voidAttributes = []string{
	":slot",
	":if",
	":elif",
	":else",
	":for",
}

func renderStartTag(tag *Tag, output *[]runtime.Statement) (err error) {
	*output = append(*output, &runtime.Text{fmt.Sprintf("<%s ", tag.Name)})
	for key, value := range tag.Attributes {
		if strings.HasPrefix(key, ":") && !strings.HasPrefix(key, "::") {
			if slices.Contains(voidAttributes, key) {
				continue
			}

			vm, err := expression.Create(value)
			if err != nil {
				return err
			}

			*output = append(*output, &runtime.Attribute{key[1:], vm})
		} else {
			if strings.HasPrefix(key, "::") {
				key = key[1:]
			}
			*output = append(*output, &runtime.Text{fmt.Sprintf(`%s="%s" `, key, value)})
		}
	}

	closingBracket := ">"
	if tag.IsSelfClosing {
		closingBracket = "/>"
	}

	*output = append(*output, &runtime.Text{closingBracket})

	return nil
}
