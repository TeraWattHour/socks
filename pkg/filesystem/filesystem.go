package filesystem

import (
	"github.com/terawatthour/socks/pkg/evaluator"
	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/preprocessor"
	"github.com/terawatthour/socks/pkg/tokenizer"
	"os"
	"path"
)

type FileSystem struct {
	Root      string
	Files     map[string]string
	Templates map[string]*evaluator.Evaluator
}

func NewFileSystem(root string) (*FileSystem, error) {
	fs := &FileSystem{
		Root:      root,
		Files:     make(map[string]string),
		Templates: make(map[string]*evaluator.Evaluator),
	}

	if err := fs.loadDirectory(root); err != nil {
		return nil, err
	}
	if err := fs.preprocessFiles(); err != nil {
		return nil, err
	}
	if err := fs.parseFiles(); err != nil {
		return nil, err
	}

	return fs, nil
}

func (fs *FileSystem) parseFiles() error {
	for name, content := range fs.Files {
		tok := tokenizer.NewTokenizer(content)
		if err := tok.Tokenize(); err != nil {
			return err
		}
		par := parser.NewParser(tok)
		if err := par.Parse(); err != nil {
			return err
		}

		fs.Templates[name] = evaluator.NewEvaluator(par)
	}

	return nil
}

func (fs *FileSystem) preprocessFiles() error {
	proc := preprocessor.NewPreprocessor(fs.Files)

	for filename := range fs.Files {
		if result, err := proc.Preprocess(filename); err != nil {
			return err
		} else {
			fs.Files[filename] = result
		}
	}

	return nil
}

func (fs *FileSystem) loadDirectory(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			by, err := os.ReadFile(path.Join(root, entry.Name()))
			if err != nil {
				return err
			}
			withoutBase := path.Join(root, entry.Name())[len(fs.Root)+1:]
			fs.Files[withoutBase] = string(by)
		} else {
			if err := fs.loadDirectory(path.Join(root, entry.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}
