package preprocessor

import (
	"errors"
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	errors2 "github.com/terawatthour/socks/pkg/errors"
	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/tokenizer"
)

type Preprocessor struct {
	files         map[string]string
	processed     map[string]string
	staticContext map[string]interface{}
}

type filePreprocessor struct {
	preprocessor *Preprocessor
	filename     string
	programs     []parser.Program
	result       []parser.Program
	i            int
}

func New(files map[string]string, staticContext map[string]interface{}) *Preprocessor {
	return &Preprocessor{
		files:         files,
		staticContext: staticContext,
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
		return nil, errors2.NewError(fmt.Sprintf("template `%s` not found", fp.filename))
	}

	elements, err := tokenizer.Tokenize(content)
	if err != nil {
		var tokenizerError *errors2.Error
		errors.As(err, &tokenizerError)
		tokenizerError.File = fp.filename

		return nil, tokenizerError
	}

	fp.programs, err = parser.Parse(elements)
	if err != nil {
		var parserError *errors2.Error
		errors.As(err, &parserError)
		parserError.File = fp.filename

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
				return nil, errors2.NewErrorWithLocation("extend statement must have a valid file name", program.Location())
			}
		case "template":
			if err := fp.evaluateTemplateStatement(); err != nil {
				return nil, err
			}
		case "slot":
			if keepSlots {
				fp.result = append(fp.result, program)
			} else if program.(*parser.SlotStatement).Parent != nil {
				program.(*parser.SlotStatement).Parent.ChangeProgramCount(-1)
			}
		default:
			fp.result = append(fp.result, program)
		}
		fp.i++
	}

	if extends != "" {
		if err := fp.extendTemplate(extends); err != nil {
			return nil, err
		}
	}

	return fp.result, nil
}

func (fp *filePreprocessor) evaluateTemplateStatement() error {
	templateStatement := fp.programs[fp.i].(*parser.TemplateStatement)
	templateName := templateStatement.Template

	resolvedPath := helpers.ResolvePath(fp.filename, templateName)

	includedPrograms, err := fp.preprocessor.Preprocess(resolvedPath, true)
	if err != nil {
		return err
	}

	// find all defines within the nested template block
	defines := map[string][]parser.Program{}
	for i := 0; i < templateStatement.Programs; i++ {
		program := fp.programs[fp.i+1+i]
		if program.Kind() != "define" {
			continue
		}
		defineStatement := program.(*parser.DefineStatement)
		if defineStatement.Depth-1 != templateStatement.Depth {
			continue
		}
		defines[defineStatement.Name] = fp.programs[fp.i+2+i : fp.i+2+i+defineStatement.Programs]
		i += defineStatement.Programs
	}

	fp.i += templateStatement.Programs

	beforeCount := len(fp.result)

	for i := 0; i < len(includedPrograms); i++ {
		includedProgram := includedPrograms[i]
		if includedProgram.Kind() != "slot" {
			fp.result = append(fp.result, includedProgram)
			continue
		}
		slotStatement := includedProgram.(*parser.SlotStatement)

		// slot is nested within something else
		if slotStatement.Depth != 0 {
			fp.result = append(fp.result, includedProgram)
			continue
		}

		definedPrograms := defines[slotStatement.Name]
		if definedPrograms == nil {
			continue
		}

		// swap the contents of the slot with the contents of the define statement
		fp.result = append(fp.result, definedPrograms...)

		// skipping fallback content of the slot if it is overwritten
		i += slotStatement.Programs
	}

	delta := len(fp.result) - beforeCount - templateStatement.Programs - 1
	templateStatement.ChangeProgramCount(delta)

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
		// accept only slots the top level
		parentProgram := parentPrograms[i]
		if parentProgram.Kind() != "slot" {
			merged = append(merged, parentProgram)
			continue
		}
		slotStatement := parentProgram.(*parser.SlotStatement)
		if slotStatement.Depth != 0 && slotStatement.Parent != nil && slotStatement.Parent.Kind() != "define" {
			merged = append(merged, parentProgram)
			continue
		}

		defineFound := false

		// swap the contents of the slot with the contents of the define statement
		for j := 0; j < len(fp.result); j++ {
			program := fp.result[j]
			if program.Kind() != "define" {
				continue
			}
			defineStatement := program.(*parser.DefineStatement)
			if defineStatement.Name != slotStatement.Name || defineStatement.Depth != 0 {
				continue
			}

			defineFound = true
			for k := 0; k < defineStatement.Programs; k++ {
				merged = append(merged, fp.result[j+1+k])
			}
		}

		// skipping fallback content of the slot if it is overwritten
		if defineFound {
			i += slotStatement.Programs
		}
	}

	fp.result = merged

	return nil
}
