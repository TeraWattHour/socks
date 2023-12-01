package preprocessor

// Preprocessor is responsible for preprocessing the template before it is evaluated and removing potential comments.
// Preprocessing should only be done once per template.
// The features of this preprocessor are:
// - extend: extend another template
// - template: embed a template into another template
// - slot: define a slot in a template
// - define: define a block from an extended template

import (
	"fmt"
	"github.com/terawatthour/socks/internal/helpers"
	"github.com/terawatthour/socks/pkg/evaluator"
	"github.com/terawatthour/socks/pkg/parser"
	"github.com/terawatthour/socks/pkg/tokenizer"
)

type Preprocessor struct {
	files         map[string]string
	processed     map[string]string
	staticContext map[string]interface{}
}

type FilePreprocessor struct {
	Preprocessor *Preprocessor
	Parser       *parser.Parser
	Filename     string
	Result       string
	i            int
}

type TagPreprocessor struct {
	FilePreprocessor *FilePreprocessor
	Program          *parser.TagProgram
}

func NewPreprocessor(files map[string]string, staticContext map[string]interface{}) *Preprocessor {
	return &Preprocessor{
		files:         files,
		staticContext: staticContext,
	}
}

func NewFilePreprocessor(filename string, preprocessor *Preprocessor) *FilePreprocessor {
	return &FilePreprocessor{
		Preprocessor: preprocessor,
		Filename:     filename,
		Result:       "",
		i:            0,
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
	content, ok := fp.Preprocessor.files[fp.Filename]
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

	fp.Result = par.Tokenizer.Template
	fp.Parser = par

	fp.i = 0
	for fp.i < len(fp.Parser.Programs) {
		program := fp.Parser.Programs[fp.i]

		if program.Tag.Kind == tokenizer.CommentKind {
			fp.Result = fp.Result[:program.Tag.Start] + fp.Result[program.Tag.End+1:]
			fp.i += 1
		} else if program.Tag.Kind != tokenizer.PreprocessorKind || program.Statement.Kind() == "slot" || program.Statement.Kind() == "end" {
			// skip non-preprocessor tags and slot tags which need to be replaced by the parent template
			// end tags must be handled by their opening tag

			fp.i += 1
			continue
		} else {
			tagPreprocessor := NewTagPreprocessor(fp, &program)
			err := tagPreprocessor.evaluateProgram()
			if err != nil {
				return "", err
			}
		}

		tok := tokenizer.NewTokenizer(fp.Result)
		if err := tok.Tokenize(); err != nil {
			return "", err
		}

		fp.Parser = parser.NewParser(tok)
		if err := fp.Parser.Parse(); err != nil {
			return "", err
		}

		fp.i = 0
	}

	result, err := evaluator.NewEvaluator(fp.Parser, evaluator.StaticMode).Evaluate(fp.Preprocessor.staticContext)
	if err != nil {
		return "", err
	}

	tok = tokenizer.NewTokenizer(result)
	if err := tok.Tokenize(); err != nil {
		return "", err
	}

	fp.Parser = parser.NewParser(tok)
	if err := fp.Parser.Parse(); err != nil {
		return "", err
	}

	return result, nil
}

func (tp *TagPreprocessor) evaluateProgram() error {
	switch tp.Program.Statement.Kind() {
	case "extend":
		return tp.evaluateExtendStatement()
	case "template":
		return tp.evaluateTemplateStatement()
	}

	return nil
}

func (tp *TagPreprocessor) evaluateTemplateStatement() error {
	templateStatement := tp.Program.Statement.(*parser.TemplateStatement)
	embeddedTemplateName := templateStatement.Template

	embeddedTemplate, err := tp.FilePreprocessor.Preprocessor.Preprocess(embeddedTemplateName)
	if err != nil {
		return err
	}

	var currentOffset int

	defines := []*parser.DefineStatement{}
	for _, program := range tp.FilePreprocessor.Parser.Programs {
		if program.Statement.Kind() == "define" {
			parents := program.Statement.(*parser.DefineStatement).Parents
			if len(parents) == 0 {
				continue
			}
			directParent := parents[len(parents)-1]
			if directParent.Kind() != "template" {
				continue
			}

			if directParent.(*parser.TemplateStatement).StartTag.Start == tp.Program.Statement.(*parser.TemplateStatement).StartTag.Start {
				defines = append(defines, program.Statement.(*parser.DefineStatement))
			}
		}
	}

	tok := tokenizer.NewTokenizer(embeddedTemplate)
	if err := tok.Tokenize(); err != nil {
		return err
	}

	par := parser.NewParser(tok)
	if err := par.Parse(); err != nil {
		return err
	}

	offset := 0
slotLoop:
	for _, program := range par.Programs {
		if program.Statement.Kind() == "slot" {
			slotStatement := program.Statement.(*parser.SlotStatement)
			for _, defineStatement := range defines {
				if defineStatement.Name != slotStatement.Name {
					continue
				}

				ru := tp.FilePreprocessor.Parser.Tokenizer.Runes
				innerContent := ru[defineStatement.StartTag.End+1 : defineStatement.EndTag.Start]
				embeddedTemplate, currentOffset = helpers.SwapInnerText([]rune(embeddedTemplate), slotStatement.StartTag.Start+offset, slotStatement.EndTag.End+1+offset, innerContent)
				offset += currentOffset
				continue slotLoop
			}

			// use fallback if no define statement is found
			ru := []rune(embeddedTemplate)
			innerContent := ru[slotStatement.StartTag.End+1+offset : slotStatement.EndTag.Start+offset]
			embeddedTemplate, currentOffset = helpers.SwapInnerText(ru, slotStatement.StartTag.End+1+offset, slotStatement.EndTag.Start+offset, innerContent)
			offset += currentOffset
		}
	}

	ru := tp.FilePreprocessor.Parser.Tokenizer.Runes
	tp.FilePreprocessor.Result, _ = helpers.SwapInnerText(ru, templateStatement.StartTag.Start, templateStatement.EndTag.End+1, []rune(embeddedTemplate))

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

	extendedTemplate, err := tp.FilePreprocessor.Preprocessor.Preprocess(extendedTemplateName)
	if err != nil {
		return err
	}

	result, err := extendTemplate(tp.FilePreprocessor.Result, extendedTemplate)
	if err != nil {
		return err
	}

	tp.FilePreprocessor.Result = result

	return nil
}

func extendTemplate(baseTemplate, extendedTemplate string) (string, error) {
	var currentOffset int

	baseTokenizer := tokenizer.NewTokenizer(baseTemplate)
	if err := baseTokenizer.Tokenize(); err != nil {
		return "", err
	}

	baseParser := parser.NewParser(baseTokenizer)
	if err := baseParser.Parse(); err != nil {
		return "", err
	}

	extendedTokenizer := tokenizer.NewTokenizer(extendedTemplate)
	if err := extendedTokenizer.Tokenize(); err != nil {
		return "", err
	}

	extendedParser := parser.NewParser(extendedTokenizer)
	if err := extendedParser.Parse(); err != nil {
		return "", err
	}

	offset := 0

slotsLoop:
	for _, program := range extendedParser.Programs {
		if program.Statement.Kind() == "slot" {
			slot := program.Statement.(*parser.SlotStatement)
			for _, program := range baseParser.Programs {
				if program.Statement.Kind() != "define" {
					continue
				}

				defineStatement := program.Statement.(*parser.DefineStatement)
				if defineStatement.Name != slot.Name || len(defineStatement.Parents) != 0 {
					continue
				}

				innerContent := baseParser.Tokenizer.Runes[defineStatement.StartTag.End+1 : defineStatement.EndTag.Start]
				extendedTemplate, currentOffset = helpers.SwapInnerText([]rune(extendedTemplate), slot.StartTag.Start+offset, slot.EndTag.End+1+offset, innerContent)
				offset += currentOffset
				continue slotsLoop
			}
		}
	}

	return extendedTemplate, nil
}
