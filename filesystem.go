package socks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type fileSystem struct {
	options       *Options
	files         map[string]string
	nativeMap     map[string]string
	templates     map[string]*evaluator
	staticContext map[string]interface{}
}

func newFileSystem(options *Options) *fileSystem {
	return &fileSystem{
		options:   options,
		files:     make(map[string]string),
		templates: make(map[string]*evaluator),
		nativeMap: make(map[string]string),
	}
}

func (fs *fileSystem) loadTemplates(pattern string, removePrefix string) error {
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

		by, err := os.ReadFile(entryName)
		if err != nil {
			return err
		}

		trimmed := strings.TrimLeft(strings.TrimPrefix(entryName, removePrefix), "/")
		fs.nativeMap[trimmed] = entryName
		fs.files[trimmed] = string(by)
	}

	return nil
}

func (fs *fileSystem) loadTemplateFromString(filename string, content string) {
	fs.files[filename] = content
	fs.nativeMap[filename] = filename
}

func (fs *fileSystem) preprocessTemplates(staticContext map[string]interface{}) error {
	proc := New(fs.files, fs.nativeMap, staticContext, fs.options.Sanitizer)
	for filename := range fs.files {
		if content, err := proc.Preprocess(filename, false); err != nil {
			return err
		} else {
			fs.templates[filename] = newEvaluator(content, fs.options.Sanitizer)
		}
	}

	return nil
}
