package socks

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/filesystem"
)

type soc struct {
	fs *filesystem.FileSystem
}

type Socks interface {
	Run(template string, context map[string]interface{}) (string, error)
}

func (s *soc) Run(template string, context map[string]interface{}) (string, error) {
	eval, ok := s.fs.Templates[template]
	if !ok {
		return "", fmt.Errorf("template %s not found", template)
	}

	return eval.Evaluate(context)
}

func NewSocks(templatesDirectory string) (Socks, error) {
	fs, err := filesystem.NewFileSystem(templatesDirectory)
	if err != nil {
		return nil, err
	}

	s := &soc{
		fs: fs,
	}

	return s, nil
}
