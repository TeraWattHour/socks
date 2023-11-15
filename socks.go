package socks

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/filesystem"
)

type Socks interface {
	Run(template string, context map[string]interface{}) (string, error)
	SetGlobals(value map[string]interface{})
}

type sock struct {
	fs      *filesystem.FileSystem
	globals map[string]interface{}
}

func NewSocks(templatesDirectory string) (Socks, error) {
	fs, err := filesystem.NewFileSystem(templatesDirectory)
	if err != nil {
		return nil, err
	}

	s := &sock{
		fs: fs,
	}

	return s, nil
}

func (s *sock) Run(template string, context map[string]interface{}) (string, error) {
	eval, ok := s.fs.Templates[template]
	if !ok {
		return "", fmt.Errorf("template %s not found", template)
	}

	return eval.Evaluate(helpers.CombineMaps(s.globals, context))
}

func (s *sock) SetGlobals(value map[string]interface{}) {
	s.globals = value
}
