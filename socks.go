package socks

import (
	"bytes"
	"fmt"
	"io"
	"maps"
)

type Socks interface {
	ExecuteToString(template string, context map[string]interface{}) (string, error)
	Execute(w io.Writer, template string, context map[string]interface{}) error

	LoadTemplates(pattern string, removePrefix ...string) error
	LoadTemplateFromString(filename string, content string)
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

func (s *socks) LoadTemplates(pattern string, removePrefix ...string) error {
	if len(removePrefix) > 1 {
		panic("expected one or zero removePrefix, got more than one")
	}
	var toRemove string
	if len(removePrefix) == 1 {
		toRemove = removePrefix[0]
	}
	return s.fs.loadTemplates(pattern, toRemove)
}

func (s *socks) LoadTemplateFromString(filename string, content string) {
	s.fs.loadTemplateFromString(filename, content)
}

func (s *socks) Compile(staticContext map[string]interface{}) error {
	maps.Copy(s.globals, staticContext)
	return s.fs.preprocessTemplates(staticContext)
}

func (s *socks) ExecuteToString(template string, context map[string]interface{}) (string, error) {
	eval, ok := s.fs.templates[template]
	if !ok {
		return "", fmt.Errorf("template `%s` not found", template)
	}

	result := bytes.NewBufferString("")
	maps.Copy(s.globals, context)
	err := eval.evaluate(result, s.globals)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

func (s *socks) Execute(w io.Writer, template string, context map[string]any) error {
	eval, ok := s.fs.templates[template]
	if !ok {
		return fmt.Errorf("template `%s` not found", template)
	}

	maps.Copy(s.globals, context)
	err := eval.evaluate(w, s.globals)
	if err != nil {
		return err
	}
	return nil
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
