package preprocessor

import (
	"errors"
	"fmt"
	errors2 "github.com/terawatthour/socks/errors"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/parser"
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
	programs     []parser.Program
	result       []parser.Program
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

func (p *Preprocessor) Preprocess(filename string, keepSlots bool) ([]parser.Program, error) {
	filePreprocessor := &filePreprocessor{
		preprocessor: p,
		filename:     filename,
		result:       make([]parser.Program, 0),
		programs:     make([]parser.Program, 0),
		i:            0,
	}
	return filePreprocessor.preprocess(keepSlots)
}

func (fp *filePreprocessor) preprocess(keepSlots bool) (res []parser.Program, err error) {
	content, ok := fp.preprocessor.files[fp.filename]
	if !ok {
		return nil, fmt.Errorf("template `%s` not found", fp.filename)
	}

	nativeName := fp.preprocessor.nativeMap[fp.filename]

	elements, err := tokenizer.Tokenize(content)
	if err != nil {
		var tokenizerError *errors2.Error
		errors.As(err, &tokenizerError)
		tokenizerError.File = nativeName

		return nil, tokenizerError
	}

	fp.programs, err = parser.Parse(elements)
	if err != nil {
		var parserError *errors2.Error
		errors.As(err, &parserError)
		parserError.File = nativeName

		return nil, parserError
	}

	var extends = ""

	fp.i = 0
	for fp.i < len(fp.programs) {
		program := fp.programs[fp.i]

		switch program.Kind() {
		case "extend":
			extends = program.(*parser.ExtendStatement).Template
			if extends == "" {
				return nil, errors2.New("extend statement must take a valid file name as an argument", program.Location())
			}
			fp.i++
		case "template":
			if err := fp.evaluateTemplateStatement(); err != nil {
				return nil, err
			}
		case "end":
			end := program.(*parser.EndStatement)
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

	evaluationResult, err := evaluate(fp.result, fp.preprocessor.staticContext, fp.preprocessor.sanitizer)
	if err != nil {
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
			textLeft := fp.result[i-1].(*parser.Text)
			textRight := fp.result[i].(*parser.Text)
			textLeft.Content += textRight.Content
			fp.result = append(fp.result[:i], fp.result[i+1:]...)
			i--
		}
	}
}

func (fp *filePreprocessor) evaluateTemplateStatement() error {
	templateStatement := fp.programs[fp.i].(*parser.TemplateStatement)
	templateName := templateStatement.Template

	resolvedPath := helpers.ResolvePath(fp.filename, templateName)

	includedPrograms, err := fp.preprocessor.Preprocess(resolvedPath, true)
	if err != nil {
		return err
	}

	fp.i++

	defines := map[string][]parser.Program{}
	for ; fp.program() != templateStatement.EndStatement; fp.i++ {
		defineStatement, ok := fp.program().(*parser.DefineStatement)
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
		slotStatement, ok := includedProgram.(*parser.SlotStatement)
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

	merged := make([]parser.Program, 0)

	for i := 0; i < len(parentPrograms); i++ {
		// find all parent's slots that can be filled by the child template
		slotStatement, ok := parentPrograms[i].(*parser.SlotStatement)
		if !ok || slices.Contains(merged, parser.Program(slotStatement.Parent)) {
			merged = append(merged, parentPrograms[i])
			continue
		}

		i++
		defineFound := false

		// swap the contents of the slot with the contents of the define statement
		for j := 0; j < len(fp.result); j++ {
			defineStatement, ok := fp.result[j].(*parser.DefineStatement)
			if !ok || defineStatement.Name != slotStatement.Name || slices.Contains(merged, parser.Program(defineStatement.Parent)) {
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

func (fp *filePreprocessor) program() parser.Program {
	return fp.programs[fp.i]
}
