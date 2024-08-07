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
	options     *Options
	templates   map[string]*runtime.Evaluator
	files       map[string]io.Reader
	fileHandles map[string]*os.File
}

func newFileSystem(options *Options) *fileSystem {
	return &fileSystem{
		options:     options,
		templates:   make(map[string]*runtime.Evaluator),
		files:       make(map[string]io.Reader),
		fileHandles: make(map[string]*os.File),
	}
}

func (fs *fileSystem) preprocessTemplates(ctx runtime.Context) error {
	preprocessed, err := Preprocess(fs.files, ctx, fs.options.Sanitizer)
	if err != nil {
		return err
	}

	for path, programs := range preprocessed {
		fs.templates[path] = runtime.NewEvaluator(helpers.File{Name: path}, programs, fs.options.Sanitizer)
	}

	for _, file := range fs.fileHandles {
		_ = file.Close()
	}

	fs.files = make(map[string]io.Reader)
	fs.fileHandles = make(map[string]*os.File)

	return nil
}

// loadTemplates opens all files matching the provided globs.
func (fs *fileSystem) loadTemplates(globs ...string) error {
	for _, glob := range globs {
		matchedFiles, err := filepath.Glob(glob)
		if err != nil {
			return err
		}

		if len(matchedFiles) == 0 {
			return fmt.Errorf("no files found")
		}

		for _, path := range matchedFiles {
			if _, ok := fs.files[path]; ok {
				continue
			}

			st, err := os.Stat(path)
			if err != nil {
				return err
			}
			if st.IsDir() {
				continue
			}

			file, err := os.OpenFile(path, os.O_RDONLY, 0)
			if err != nil {
				return err
			}

			fs.fileHandles[path] = file
			fs.files[path] = file
		}
	}

	return nil
}

func (fs *fileSystem) loadTemplate(filename string, content io.ReadCloser) {
	fs.files[filename] = content
}
