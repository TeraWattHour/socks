package preprocessor

import (
	"fmt"
	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/tokenizer"
)

type Preprocessor struct {
	Files map[string]string
}

type FilePreprocessor struct {
	Preprocessor *Preprocessor
	Parser       *parser.Parser
	Filename     string
	Result       string
}

type TagPreprocessor struct {
	FilePreprocessor *FilePreprocessor
	Program          *parser.TagProgram
}

func NewPreprocessor(files map[string]string) *Preprocessor {
	return &Preprocessor{
		Files: files,
	}
}

func NewFilePreprocessor(filename string, preprocessor *Preprocessor) *FilePreprocessor {
	return &FilePreprocessor{
		Preprocessor: preprocessor,
		Filename:     filename,
		Result:       "",
	}
}

func NewTagPreprocessor(filePreprocessor *FilePreprocessor, program *parser.TagProgram) *TagPreprocessor {
	return &TagPreprocessor{
		FilePreprocessor: filePreprocessor,
		Program:          program,
	}
}

func (p *Preprocessor) Preprocess(filename string) (string, error) {
	return NewFilePreprocessor(filename, p).preprocess()
}

func (fp *FilePreprocessor) preprocess() (string, error) {
	content, ok := fp.Preprocessor.Files[fp.Filename]
	if !ok {
		return "", fmt.Errorf("template %s not found", fp.Filename)
	}

	tok := tokenizer.NewTokenizer(content)
	if err := tok.Tokenize(); err != nil {
		return "", err
	}

	par := parser.NewParser(tok)
	if err := par.Parse(); err != nil {
		return "", err
	}

	fp.Parser = par

	for _, program := range par.Programs {
		if program.Tag.Kind != "preprocessor" {
			continue
		}

		err := NewTagPreprocessor(fp, &program).evaluateProgram()
		if err != nil {
			return "", err
		}
	}

	return fp.Result, nil
}

func (tp *TagPreprocessor) evaluateProgram() error {
	switch tp.Program.Statement.Kind() {
	case "extend":
		return tp.evaluateExtendStatement()
	}

	return nil
}

func (tp *TagPreprocessor) evaluateSlotStatement() error {
	slotStatement := tp.Program.Statement.(*parser.SlotStatement)
	ru := tp.FilePreprocessor.Parser.Tokenizer.Runes
	tp.FilePreprocessor.Result = string(ru[:slotStatement.StartTag.Start]) + string(ru[slotStatement.StartTag.End+1:slotStatement.EndTag.Start]) + string(ru[slotStatement.EndTag.End+1:])
	return nil
}

func (tp *TagPreprocessor) evaluateExtendStatement() error {
	extendedTemplateName := tp.Program.Statement.(*parser.ExtendStatement).Template

	extendedTemplate, ok := tp.FilePreprocessor.Preprocessor.Files[extendedTemplateName]
	if !ok {
		return fmt.Errorf("template %s not found", extendedTemplateName)
	}
	tp.FilePreprocessor.Result = extendedTemplate

	for {
		newTemplate, updated := tp.FilePreprocessor.replaceSlots()
		if !updated {
			break
		}

		tp.FilePreprocessor.Result = newTemplate
	}

	return nil
}

// replaceSlots replaces one slot at a time, if there are no slots left to replace, it returns "", false
func (fp *FilePreprocessor) replaceSlots() (string, bool) {
	tok := tokenizer.NewTokenizer(fp.Result)
	if err := tok.Tokenize(); err != nil {
		return "", false
	}

	var slot *parser.SlotStatement
	par := parser.NewParser(tok)
	if err := par.Parse(); err != nil {
		return "", false
	}

	for _, program := range par.Programs {
		if program.Tag.Kind == "preprocessor" && program.Statement.Kind() == "slot" {
			slot = program.Statement.(*parser.SlotStatement)
			break
		}
	}

	// no slots left to replace
	if slot == nil {
		return "", false
	}

	for _, program := range fp.Parser.Programs {
		if program.Statement.Kind() == "define" && program.Statement.(*parser.DefineStatement).Name == slot.Name {
			defineStatement := program.Statement.(*parser.DefineStatement)
			newContent := string(fp.Parser.Tokenizer.Runes[defineStatement.StartTag.End+1 : defineStatement.EndTag.Start])
			newTemplate := fmt.Sprintf("%s%s%s", string(tok.Runes[:slot.StartTag.Start]), newContent, string(tok.Runes[slot.EndTag.End+1:]))
			return newTemplate, true
		}
	}

	fallback := string(tok.Runes[slot.StartTag.End+1 : slot.EndTag.Start])
	newTemplate := fmt.Sprintf("%s%s%s", string(tok.Runes[:slot.StartTag.Start]), fallback, string(tok.Runes[slot.EndTag.End+1:]))
	return newTemplate, true
}
