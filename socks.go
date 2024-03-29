package socks

import (
	"errors"
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	errors2 "github.com/terawatthour/socks/pkg/errors"
	"io"
)

type Socks interface {
	ExecuteToString(template string, context map[string]interface{}) (string, error)
	Execute(w io.Writer, template string, context map[string]interface{}) (int, error)

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
	s.globals = helpers.CombineMaps(s.globals, staticContext)
	return s.fs.preprocessTemplates(staticContext)
}

func (s *socks) ExecuteToString(template string, context map[string]interface{}) (string, error) {
	eval, ok := s.fs.templates[template]
	if !ok {
		return "", fmt.Errorf("template `%s` not found", template)
	}

	nativeName := s.fs.nativeMap[template]

	result, err := eval.Evaluate(helpers.CombineMaps(s.globals, context))
	if err != nil {
		var nativeError *errors2.Error
		if errors.As(err, &nativeError) {
			nativeError.File = nativeName
			return "", nativeError
		}
		return "", err
	}
	return result, nil
}

func (s *socks) Execute(w io.Writer, template string, context map[string]any) (int, error) {
	eval, ok := s.fs.templates[template]
	if !ok {
		return 0, fmt.Errorf("template `%s` not found", template)
	}

	result, err := eval.Evaluate(helpers.CombineMaps(s.globals, context))
	if err != nil {
		var nativeError *errors2.Error
		if errors.As(err, &nativeError) {
			nativeError.File = template
			return 0, nativeError
		}
		return 0, err
	}
	return w.Write([]byte(result))
}

func (s *socks) AddGlobal(key string, value any) {
	s.globals[key] = value
}

func (s *socks) AddGlobals(value map[string]any) {
	s.globals = helpers.CombineMaps(s.globals, value)
}

func (s *socks) GetGlobals() map[string]any {
	return s.globals
}

func (s *socks) ClearGlobals() {
	clear(s.globals)
}
