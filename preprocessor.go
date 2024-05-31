package socks

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/tokenizer"
	"slices"
)

type Preprocessor struct {
	files         map[string]string
	nativeMap     map[string]string
	processed     map[string]string
	staticContext map[string]any
	sanitizer     func(string) string
}

type filePreprocessor struct {
	preprocessor *Preprocessor
	filename     string
	programs     []Statement
	result       []Statement
	i            int
}

func New(files map[string]string, nativeMap map[string]string, staticContext map[string]interface{}, sanitizer func(string) string) *Preprocessor {
	return &Preprocessor{
		files:         files,
		staticContext: staticContext,
		sanitizer:     sanitizer,
		nativeMap:     nativeMap,
	}
}

func (p *Preprocessor) Preprocess(filename string, keepSlots bool) ([]Statement, error) {
	filePreprocessor := &filePreprocessor{
		preprocessor: p,
		filename:     filename,
		result:       make([]Statement, 0),
		programs:     make([]Statement, 0),
		i:            0,
	}
	return filePreprocessor.preprocess(keepSlots)
}

func (fp *filePreprocessor) preprocess(keepSlots bool) (res []Statement, err error) {
	content, ok := fp.preprocessor.files[fp.filename]
	if !ok {
		return nil, fmt.Errorf("template `%s` not found", fp.filename)
	}

	nativeName := fp.preprocessor.nativeMap[fp.filename]

	elements, err := tokenizer.Tokenize(nativeName, content)
	if err != nil {
		return nil, err
	}

	fp.programs, err = Parse(helpers.File{Name: nativeName, Content: content}, elements)
	if err != nil {
		return nil, err
	}

	var extends string

	fp.i = 0
	for fp.i < len(fp.programs) {
		program := fp.programs[fp.i]

		switch program.Kind() {
		case "extend":
			extends = program.(*ExtendStatement).Template
			fp.i++
		case "template":
			if err := fp.evaluateTemplateStatement(); err != nil {
				return nil, err
			}
		case "end":
			end := program.(*EndStatement)
			fp.i++
			if !keepSlots && end.ClosedStatement.Kind() == "slot" {
				continue
			}
			fp.result = append(fp.result, program)
		case "slot":
			if !keepSlots {
				fp.i++
				continue
			}
			fallthrough
		default:
			fp.result = append(fp.result, program)
			fp.i++
		}
	}

	if extends != "" {
		if err := fp.extendTemplate(extends); err != nil {
			return nil, err
		}
	}

	var evaluationResult helpers.Queue[Statement]
	staticEvaluator := newStaticEvaluator(helpers.File{nativeName, content}, &evaluationResult, fp.result, fp.preprocessor.sanitizer)

	if err := staticEvaluator.evaluate(nil, fp.preprocessor.staticContext); err != nil {
		fp.foldText()
		return fp.result, nil
	}

	fp.result = evaluationResult
	fp.foldText()

	return fp.result, nil
}

func (fp *filePreprocessor) foldText() {
	for i := 1; i < len(fp.result); i++ {
		if fp.result[i].Kind() == "text" && fp.result[i-1].Kind() == "text" {
			textLeft := fp.result[i-1].(*Text)
			textRight := fp.result[i].(*Text)
			textLeft.Content += textRight.Content
			fp.result = append(fp.result[:i], fp.result[i+1:]...)
			i--
		}
	}
}

func (fp *filePreprocessor) evaluateTemplateStatement() error {
	templateStatement := fp.programs[fp.i].(*TemplateStatement)
	templateName := templateStatement.Template

	resolvedPath := helpers.ResolvePath(fp.filename, templateName)

	includedPrograms, err := fp.preprocessor.Preprocess(resolvedPath, true)
	if err != nil {
		return err
	}

	fp.i++

	defines := map[string][]Statement{}
	for ; fp.program() != templateStatement.EndStatement; fp.i++ {
		defineStatement, ok := fp.program().(*DefineStatement)
		if !ok || defineStatement.Parent != templateStatement {
			continue
		}
		fp.i++
		for ; fp.program() != defineStatement.EndStatement; fp.i++ {
			defines[defineStatement.Name] = append(defines[defineStatement.Name], fp.program())
		}
	}

	fp.i++

	for i := 0; i < len(includedPrograms); i++ {
		includedProgram := includedPrograms[i]
		slotStatement, ok := includedProgram.(*SlotStatement)
		if !ok || slices.Contains(fp.result, slotStatement.Parent) {
			fp.result = append(fp.result, includedProgram)
			continue
		}

		definedPrograms := defines[slotStatement.Name]

		i++
		for includedPrograms[i] != slotStatement.EndStatement {
			if definedPrograms == nil {
				fp.result = append(fp.result, includedPrograms[i])
			}
			i++
		}

		if definedPrograms != nil {
			fp.result = append(fp.result, definedPrograms...)
		}
	}

	return nil
}

func (fp *filePreprocessor) extendTemplate(parentTemplate string) error {
	resolvedPath := helpers.ResolvePath(fp.filename, parentTemplate)

	parentPrograms, err := fp.preprocessor.Preprocess(resolvedPath, true)
	if err != nil {
		return err
	}

	merged := make([]Statement, 0)

	for i := 0; i < len(parentPrograms); i++ {
		// find all parent's slots that can be filled by the child template
		slotStatement, ok := parentPrograms[i].(*SlotStatement)
		if !ok || slices.Contains(merged, slotStatement.Parent) {
			merged = append(merged, parentPrograms[i])
			continue
		}

		i++
		defineFound := false

		// swap the contents of the slot with the contents of the define statement
		for j := 0; j < len(fp.result); j++ {
			defineStatement, ok := fp.result[j].(*DefineStatement)
			if !ok || defineStatement.Name != slotStatement.Name || slices.Contains(merged, defineStatement.Parent) {
				continue
			}

			defineFound = true

			j++
			for fp.result[j] != defineStatement.EndStatement {
				merged = append(merged, fp.result[j])
				j++
			}

			break
		}

		for parentPrograms[i] != slotStatement.EndStatement {
			if !defineFound {
				merged = append(merged, parentPrograms[i])
			}
			i++
		}
	}

	fp.result = merged

	return nil
}

func (fp *filePreprocessor) program() Statement {
	return fp.programs[fp.i]
}
