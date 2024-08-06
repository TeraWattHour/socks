package socks

import (
	"bytes"
	"fmt"
	"github.com/terawatthour/socks/runtime"
	"io"
	"maps"
	"strings"
)

type Socks interface {
	ExecuteToString(template string, context map[string]interface{}) (string, error)
	Execute(w io.Writer, template string, context map[string]interface{}) error

	LoadTemplates(glob string) error
	LoadTemplateFromString(filename string, reader io.Reader)
	Compile(staticContext map[string]interface{}) error

	AddGlobal(key string, value interface{})
	AddGlobals(value map[string]interface{})
	GetGlobals() map[string]interface{}
	ClearGlobals()
}

type socks struct {
	fs      *fileSystem
	globals map[string]interface{}
	options *Options
}

type Options struct {
	Sanitizer func(string) string
}

func NewSocks(options ...*Options) Socks {
	if len(options) > 1 {
		panic("expected one or no options, got more than one")
	}
	opts := &Options{}
	if len(options) == 1 {
		opts = options[0]
	}
	fs := newFileSystem(opts)

	return &socks{
		fs:      fs,
		globals: make(map[string]interface{}),
		options: opts,
	}
}

func (s *socks) LoadTemplates(glob string) error {
	return s.fs.loadTemplates(glob)
}

func (s *socks) LoadTemplateFromString(filename string, reader io.Reader) {
	s.fs.loadTemplate(filename, reader)
}

func (s *socks) Compile(staticContext runtime.Context) error {
	maps.Copy(s.globals, staticContext)
	return s.fs.preprocessTemplates(staticContext)
}

func (s *socks) ExecuteToString(template string, context runtime.Context) (string, error) {
	eval, err := s.resolveTemplate(template)
	if err != nil {
		return "", err
	}

	result := bytes.NewBufferString("")
	maps.Copy(s.globals, context)
	if err := eval.Evaluate(result, s.globals); err != nil {
		return "", err
	}
	return result.String(), nil
}

func (s *socks) Execute(w io.Writer, template string, context runtime.Context) error {
	eval, err := s.resolveTemplate(template)
	if err != nil {
		return err
	}

	maps.Copy(s.globals, context)
	return eval.Evaluate(w, s.globals)
}

func (s *socks) resolveTemplate(template string) (*runtime.Evaluator, error) {
	if eval, ok := s.fs.templates[template]; ok {
		return eval, nil
	}

	var matching *runtime.Evaluator
	for key, eval := range s.fs.templates {
		if key == template || strings.HasSuffix(key, "/"+template) {
			if matching != nil {
				return nil, fmt.Errorf(`reference "%s" is ambiguous as it matches multiple templates`, template)
			}
			matching = eval
		}
	}

	if matching != nil {
		return matching, nil
	}

	return nil, fmt.Errorf("template `%s` not found", template)
}

func (s *socks) AddGlobal(key string, value any) {
	s.globals[key] = value
}

func (s *socks) AddGlobals(value map[string]any) {
	maps.Copy(s.globals, value)
}

func (s *socks) GetGlobals() map[string]any {
	return s.globals
}

func (s *socks) ClearGlobals() {
	clear(s.globals)
}
