package socks

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/filesystem"
)

type Socks interface {
	Run(template string, context map[string]interface{}) (string, error)

	LoadTemplates(patterns ...string) error
	PreprocessTemplates(staticContext map[string]interface{}) error

	AddGlobal(key string, value interface{})
	AddGlobals(value map[string]interface{})
	ClearGlobals()
	ListGlobals() map[string]interface{}
}

type sock struct {
	fs      *filesystem.FileSystem
	globals map[string]interface{}
}

func NewSocks() Socks {
	fs := filesystem.NewFileSystem()

	return &sock{
		fs:      fs,
		globals: make(map[string]interface{}),
	}
}

func (s *sock) LoadTemplates(patterns ...string) error {
	return s.fs.LoadTemplates(patterns...)
}

func (s *sock) PreprocessTemplates(staticContext map[string]interface{}) error {
	return s.fs.PreprocessFiles(staticContext)
}

func (s *sock) Run(template string, context map[string]interface{}) (string, error) {
	eval, ok := s.fs.Templates[template]
	if !ok {
		return "", fmt.Errorf("template %s not found", template)
	}

	return eval.Evaluate(helpers.CombineMaps(s.globals, context))
}

func (s *sock) AddGlobal(key string, value interface{}) {
	s.globals[key] = value
}

func (s *sock) AddGlobals(value map[string]interface{}) {
	s.globals = helpers.CombineMaps(s.globals, value)
}

func (s *sock) ListGlobals() map[string]interface{} {
	return s.globals
}

func (s *sock) ClearGlobals() {
	clear(s.globals)
}
