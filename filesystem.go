package socks

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/runtime"
	"io"
	"os"
	"path/filepath"
)

type fileSystem struct {
	options       *Options
	files         map[string]io.Reader
	templates     map[string]*runtime.Evaluator
	staticContext map[string]any
}

func newFileSystem(options *Options) *fileSystem {
	return &fileSystem{
		options:   options,
		files:     make(map[string]io.Reader),
		templates: make(map[string]*runtime.Evaluator),
	}
}

func (fs *fileSystem) loadTemplates(pattern string) error {
	entryNames, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(entryNames) == 0 {
		return fmt.Errorf("no files found")
	}

	for _, entryName := range entryNames {
		st, err := os.Stat(entryName)
		if err != nil {
			return err
		}
		if st.IsDir() {
			continue
		}

		file, err := os.OpenFile(entryName, os.O_RDONLY, 0)
		if err != nil {
			return err
		}

		fs.files[entryName] = file
	}

	return nil
}

func (fs *fileSystem) loadTemplate(filename string, content io.Reader) {
	fs.files[filename] = content
}

func (fs *fileSystem) preprocessTemplates(staticContext map[string]interface{}) error {
	preprocessed, err := Preprocess(fs.files, fs.staticContext, fs.options.Sanitizer)
	if err != nil {
		return err
	}

	for fileName, programs := range preprocessed {
		fs.templates[fileName] = runtime.NewEvaluator(helpers.File{Name: fileName}, programs, fs.options.Sanitizer)
	}

	return nil
}
