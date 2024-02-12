package socks

import (
	"github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/evaluator"
	"github.com/terawatthour/socks/pkg/preprocessor"
	"os"
	"path/filepath"
)

type fileSystem struct {
	options       *Options
	files         map[string]string
	templates     map[string]*evaluator.Evaluator
	staticContext map[string]interface{}
}

func newFileSystem(options *Options) *fileSystem {
	return &fileSystem{
		options:   options,
		files:     make(map[string]string),
		templates: make(map[string]*evaluator.Evaluator),
	}
}

func (fs *fileSystem) loadTemplates(patterns ...string) error {
	for _, pattern := range patterns {
		entryNames, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		if len(entryNames) == 0 {
			return errors.NewError("no files found")
		}

		for _, entryName := range entryNames {
			st, err := os.Stat(entryName)
			if err != nil {
				return err
			}
			if st.IsDir() {
				continue
			}

			by, err := os.ReadFile(entryName)
			if err != nil {
				return err
			}

			fs.files[entryName] = string(by)
		}
	}

	return nil
}

func (fs *fileSystem) loadTemplateFromString(filename string, content string) {
	fs.files[filename] = content
}

func (fs *fileSystem) preprocessTemplates(staticContext map[string]interface{}) error {
	proc := preprocessor.New(fs.files, staticContext, fs.options.Sanitizer)
	for filename := range fs.files {
		if content, err := proc.Preprocess(filename, false); err != nil {
			return err
		} else {
			fs.templates[filename] = evaluator.New(content, fs.options.Sanitizer)
		}
	}

	return nil
}
