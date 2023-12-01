package filesystem

import (
	"github.com/terawatthour/socks/pkg/evaluator"
	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/preprocessor"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"os"
	"path/filepath"
)

type FileSystem struct {
	Root          string
	Files         map[string]string
	Processed     map[string]string
	Templates     map[string]*evaluator.Evaluator
	staticContext map[string]interface{}
}

func NewFileSystem() *FileSystem {
	return &FileSystem{
		Files:     make(map[string]string),
		Processed: make(map[string]string),
		Templates: make(map[string]*evaluator.Evaluator),
	}
}

func (fs *FileSystem) LoadTemplates(patterns ...string) error {
	for _, pattern := range patterns {
		entryNames, err := filepath.Glob(pattern)
		if err != nil {
			return err
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

			fs.Files[entryName] = string(by)
		}
	}

	return nil
}

func (fs *FileSystem) parseProcessedFiles() error {
	for name, content := range fs.Processed {
		tok := tokenizer.NewTokenizer(content)
		if err := tok.Tokenize(); err != nil {
			return err
		}
		par := parser.NewParser(tok)
		if err := par.Parse(); err != nil {
			return err
		}

		fs.Templates[name] = evaluator.NewEvaluator(par, evaluator.RuntimeMode)
	}

	return nil
}

func (fs *FileSystem) PreprocessFiles(staticContext map[string]interface{}) error {
	proc := preprocessor.NewPreprocessor(fs.Files, staticContext)

	for filename := range fs.Files {
		if result, err := proc.Preprocess(filename); err == nil {
			fs.Processed[filename] = result
		} else {
			return err
		}
	}

	return fs.parseProcessedFiles()
}
