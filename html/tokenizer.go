package html

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"golang.org/x/net/html"
	"io"
	"slices"
	"strings"
)

type Node interface {
	Kind() string
}

type Tag struct {
	Name          string
	IsSelfClosing bool
	Attributes    map[string]string
	Children      []Node
	Location      helpers.Location
}

func (t *Tag) Kind() string {
	return "tag"
}

type Text struct {
	IsRaw     bool
	IsComment bool
	Content   string
	Location  helpers.Location
}

func (t *Text) Kind() string {
	return "text"
}

func (t *Text) String() string {
	if !t.IsRaw {
		return escape(t.Content)
	}
	return t.Content
}

type Tokenizer struct {
	*html.Tokenizer
	unclosedTags helpers.Stack[string]
	location     helpers.Location
	lastLocation helpers.Location
}

type Token struct {
	html.Token
	Location    helpers.Location
	EndLocation helpers.Location
}

func (t *Tokenizer) Next() html.TokenType {
	tokenType := t.Tokenizer.Next()
	raw := t.Raw()
	t.lastLocation = t.location
	for _, r := range string(raw) {
		if r == '\n' {
			t.location.Line++
			t.location.Column = 1
		} else {
			t.location.Column++
		}
	}
	return tokenType
}

func (t *Tokenizer) Token() Token {
	return Token{t.Tokenizer.Token(), t.lastLocation, t.location}
}

func (t *Tokenizer) CurrentLine() int {
	return t.lastLocation.Line
}

func (t *Tokenizer) CurrentColumn() int {
	return t.lastLocation.Column
}

func Tokenize(r io.Reader) ([]Node, error) {
	t := &Tokenizer{
		Tokenizer:    html.NewTokenizer(r),
		unclosedTags: make(helpers.Stack[string], 0),
		location:     helpers.Location{Line: 1, Column: 1},
	}

	elements, err := t.tokenizeBlock()
	if err != nil {
		return nil, err
	}

	if len(t.unclosedTags) > 0 {
		return nil, fmt.Errorf("unclosed tags: %s", strings.Join(t.unclosedTags, ", "))
	}

	return elements, nil
}

func (t *Tokenizer) tokenizeBlock() ([]Node, error) {
	depth := len(t.unclosedTags)

	var output []Node
	for depth <= len(t.unclosedTags) {
		if t.Next() == html.ErrorToken {
			break
		}

		partial, err := t.tokenize()
		if err != nil {
			return nil, err
		}

		if partial != nil {
			output = append(output, partial)
		}
	}

	return output, nil
}

func (t *Tokenizer) tokenize() (Node, error) {
	token := t.Token()

	switch token.Type {
	case html.ErrorToken:
		return nil, t.Err()
	case html.TextToken:
		if len(t.unclosedTags) > 0 && childTextNodesAreLiteral(t.unclosedTags[len(t.unclosedTags)-1]) {
			return &Text{Content: token.Data, IsRaw: true, Location: t.location}, nil
		} else {
			return &Text{Content: token.Data, IsRaw: false, Location: t.location}, nil
		}
	case html.StartTagToken:
		tag := &Tag{
			Name:       token.Data,
			Attributes: make(map[string]string),
			Location:   t.location,
		}

		for _, a := range token.Attr {
			if _, ok := tag.Attributes[a.Key]; ok {
				return nil, fmt.Errorf("duplicate attribute: %s", a.Key)
			}
			tag.Attributes[a.Key] = a.Val
		}

		if slices.Contains(voidElements, tag.Name) {
			return tag, nil
		}

		t.unclosedTags.Push(tag.Name)

		var err error
		tag.Children, err = t.tokenizeBlock()
		if err != nil {
			return nil, err
		}

		return tag, nil
	case html.EndTagToken:
		if len(t.unclosedTags) == 0 {
			return nil, fmt.Errorf("unexpected end tag: </%s> has nothing to close", token.Data)
		}
		closed := t.unclosedTags.Pop()
		if closed != token.Data {
			return nil, fmt.Errorf("unexpected end tag: <%s> is closed by </%s>", closed, token.Data)
		}

		return nil, nil
	case html.SelfClosingTagToken:
		tag := &Tag{
			Name:          token.Data,
			IsSelfClosing: true,
			Attributes:    make(map[string]string),
			Location:      t.location,
		}

		for _, a := range token.Attr {
			if _, ok := tag.Attributes[a.Key]; ok {
				return nil, fmt.Errorf("duplicate attribute: %s", a.Key)
			}
			tag.Attributes[a.Key] = a.Val
		}

		return tag, nil
	case html.CommentToken:
		return &Text{IsRaw: true, IsComment: true, Content: fmt.Sprintf("<!--%s-->", escapeComment(token.Data)), Location: t.location}, nil
	case html.DoctypeToken:
		content := fmt.Sprintf("<!DOCTYPE %s", escape(token.Data))

		if token.Attr != nil {
			var p, s string
			for _, a := range token.Attr {
				switch a.Key {
				case "public":
					p = a.Val
				case "system":
					s = a.Val
				}
			}

			if p != "" {
				content += " PUBLIC " + writeQuoted(p)

				if s != "" {
					content += " " + writeQuoted(s)
				}
			} else if s != "" {
				content += " SYSTEM " + writeQuoted(s)
			}
		}

		content += ">"

		return &Text{IsRaw: true, Content: content, Location: t.location}, nil
	}

	return nil, nil
}

func childTextNodesAreLiteral(tagName string) bool {
	switch tagName {
	case "iframe", "noembed", "noframes", "noscript", "plaintext", "script", "style", "xmp":
		return true
	default:
		return false
	}
}

func writeQuoted(s string) string {
	var q byte = '"'
	if strings.Contains(s, `"`) {
		q = '\''
	}

	return fmt.Sprintf("%c%s%c", q, s, q)
}

func escapeComment(s string) (res string) {
	if len(s) == 0 {
		return
	}
	i := 0
	for j := 0; j < len(s); j++ {
		escaped := ""
		switch s[j] {
		case '&':
			escaped = "&amp;"

		case '>':
			if j > 0 {
				if prev := s[j-1]; (prev != '!') && (prev != '-') {
					continue
				}
			}
			escaped = "&gt;"

		default:
			continue
		}

		if i < j {
			res += s[i:j]
		}
		res += escaped
		i = j + 1
	}

	if i < len(s) {
		res += s[i:]
	}

	return
}

func escape(s string) (res string) {
	const escapedChars = "&'<>\"\r"

	i := strings.IndexAny(s, escapedChars)
	for i != -1 {
		res += s[:i]

		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			esc = "&#39;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			esc = "&#34;"
		case '\r':
			esc = "&#13;"
		default:
			panic("unrecognized escape character")
		}
		s = s[i+1:]
		res += esc
		i = strings.IndexAny(s, escapedChars)
	}
	res += s

	return res
}

var voidElements = []string{
	"area",
	"base",
	"br",
	"col",
	"embed",
	"hr",
	"img",
	"input",
	"keygen",
	"link",
	"meta",
	"param",
	"source",
	"track",
	"wbr",
}
