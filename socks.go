package socks

import (
	"bytes"
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/runtime"
	"io"
	"maps"
	"strings"
)

type Socks struct {
	fs       *fileSystem
	globals  map[string]any
	options  *Options
	compiled bool
}

type Options struct {
	Sanitizer func(string) string
}

func New(options ...*Options) *Socks {
	if len(options) > 1 {
		panic("expected one or no options, got more than one")
	}
	opts := &Options{}
	if len(options) == 1 {
		opts = options[0]
	}

	return &Socks{
		fs:      newFileSystem(opts),
		globals: make(map[string]any),
		options: opts,
	}
}

func (s *Socks) LoadTemplates(glob ...string) error {
	s.compiled = false
	if err := s.fs.loadTemplates(glob...); err != nil {
		return err
	}

	return nil
}

func (s *Socks) LoadTemplate(filename string, reader io.ReadCloser) {
	s.compiled = false
	s.fs.loadTemplate(filename, reader)
}

func (s *Socks) Compile(staticContext map[string]any) error {
	if err := s.fs.preprocessTemplates(helpers.Combine(s.globals, staticContext)); err != nil {
		return err
	}

	s.compiled = true
	return nil
}

func (s *Socks) ExecuteToString(template string, context map[string]any) (string, error) {
	eval, err := s.resolveTemplate(template)
	if err != nil {
		return "", err
	}

	result := bytes.NewBufferString("")
	if err := eval.Evaluate(result, helpers.Combine(s.globals, context)); err != nil {
		return "", err
	}
	return result.String(), nil
}

func (s *Socks) Execute(w io.Writer, template string, context map[string]any) error {
	eval, err := s.resolveTemplate(template)
	if err != nil {
		return err
	}

	return eval.Evaluate(w, helpers.Combine(s.globals, context))
}

func (s *Socks) resolveTemplate(template string) (*runtime.Evaluator, error) {
	if !s.compiled {
		return nil, fmt.Errorf("templates not compiled")
	}

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

func (s *Socks) AddGlobal(key string, value any) {
	s.globals[key] = value
}

func (s *Socks) AddGlobals(value map[string]any) {
	maps.Copy(s.globals, value)
}

func (s *Socks) GetGlobals() map[string]any {
	return s.globals
}

func (s *Socks) ClearGlobals() {
	clear(s.globals)
}
