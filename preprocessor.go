package socks

import (
	"fmt"
	"github.com/terawatthour/socks/html"
	"github.com/terawatthour/socks/runtime"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

type Preprocessor struct {
	files                 map[string][]runtime.Statement
	preprocessed          map[string][]runtime.Statement
	preprocessedWithSlots map[string][]runtime.Statement

	staticContext runtime.Context
	sanitizer     func(string) string
}

// Preprocess reads and preprocesses all files from the provided map. It takes ownership of the files and closes them.
func Preprocess(files map[string]io.Reader, staticContext runtime.Context, sanitizer func(string) string) (preprocessed map[string][]runtime.Statement, err error) {
	parsedFiles := make(map[string][]runtime.Statement)

	for filename, file := range files {
		if parsedFiles[filename], err = html.Parse(file); err != nil {
			return nil, err
		}
	}

	p := &Preprocessor{
		files:                 parsedFiles,
		preprocessed:          make(map[string][]runtime.Statement),
		preprocessedWithSlots: make(map[string][]runtime.Statement),
		staticContext:         staticContext,
		sanitizer:             sanitizer,
	}
	for filename := range files {
		if err = p.preprocess(filename, false); err != nil {
			return nil, err
		}
	}

	return p.preprocessed, nil
}

func (p *Preprocessor) preprocess(filename string, keepSlots bool, cycle ...string) error {
	if slices.Contains(cycle, filename) {
		return fmt.Errorf("cyclic import detected: %v", strings.Join(append(cycle, filename), "->"))
	}

	// if file was already preprocessed in selected mode, return it immediately
	if _, ok := p.preprocessedFile(filename, keepSlots); ok {
		return nil
	}

	output, err := p.preprocessBlock(filename, p.files[filename], keepSlots, cycle...)
	if err != nil {
		return err
	}

	output = foldTexts(output)

	if keepSlots {
		p.preprocessedWithSlots[filename] = output
	} else {
		p.preprocessed[filename] = output
	}

	return nil
}

func (p *Preprocessor) preprocessBlock(filename string, block []runtime.Statement, keepSlots bool, cycle ...string) (output []runtime.Statement, err error) {
	for _, program := range block {
		switch program := program.(type) {
		case *runtime.Component:
			componentPath := filepath.Join(filename, "..", program.Name)

			if _, ok := p.files[componentPath]; !ok {
				return nil, fmt.Errorf("component `%s` not found", program.Name)
			}

			if err := p.preprocess(componentPath, true, append(cycle, filename)...); err != nil {
				return nil, err
			}

			for k, pr := range program.Defines {
				program.Defines[k], err = p.preprocessBlock(filename, pr, true, cycle...)
				if err != nil {
					return nil, err
				}
			}

			output = append(output, replaceSlots(p.preprocessedWithSlots[componentPath], program.Defines)...)
		default:
			if slot, ok := program.(*runtime.Slot); ok && !keepSlots {
				block, err := p.preprocessBlock(filename, slot.Children, false, cycle...)
				if err != nil {
					return nil, err
				}
				output = append(output, block...)
				continue
			}
			output = append(output, program)
		}
	}

	return output, nil
}

func (p *Preprocessor) preprocessedFile(filename string, keepSlots bool) ([]runtime.Statement, bool) {
	if keepSlots {
		if preprocessed, ok := p.preprocessedWithSlots[filename]; ok {
			return preprocessed, true
		}
	} else {
		if preprocessed, ok := p.preprocessed[filename]; ok {
			return preprocessed, true
		}
	}

	return nil, false
}

func replaceSlots(component []runtime.Statement, defines map[string][]runtime.Statement) []runtime.Statement {
	var result []runtime.Statement
	for _, program := range component {
		switch program := program.(type) {
		case *runtime.Slot:
			if definedPrograms, ok := defines[program.Name]; ok {
				result = append(result, definedPrograms...)
				continue
			}
			result = append(result, program.Children...)
		default:
			result = append(result, program)
		}
	}

	return result
}

func foldTexts(statements []runtime.Statement) []runtime.Statement {
	var result []runtime.Statement
	for i, statement := range statements {
		if i == 0 {
			result = append(result, statement)
			continue
		}

		if text, ok := statement.(*runtime.Text); ok {
			if lastText, ok := result[len(result)-1].(*runtime.Text); ok {
				result[len(result)-1] = &runtime.Text{Content: lastText.Content + text.Content}
				continue
			}
		}

		result = append(result, statement)
	}

	return result
}
