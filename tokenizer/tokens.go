package tokenizer

const (
	TokIdent   = "identifier"
	TokNumeric = "numeric"

	TokString = "string"
	TokComma  = "comma"
	TokAt     = "at"

	TokLparen = "lparen"
	TokRparen = "rparen"
	TokLbrack = "lbrack"
	TokRbrack = "rbrack"

	TokLt  = "lt"
	TokGt  = "gt"
	TokEq  = "eq"
	TokNeq = "neq"
	TokLte = "lte"
	TokGte = "gte"

	TokBang          = "bang"
	TokPlus          = "plus"
	TokMinus         = "minus"
	TokAsterisk      = "asterisk"
	TokSlash         = "slash"
	TokModulo        = "modulo"
	TokPower         = "power"
	TokColon         = "colon"
	TokQuestion      = "question"
	TokDot           = "dot"
	TokOptionalChain = "optional_chain"
	TokElvis         = "elvis"

	TokIn    = "in"
	TokTrue  = "true"
	TokNot   = "not"
	TokFalse = "false"
	TokAnd   = "and"
	TokOr    = "or"
	TokWith  = "with"
	TokNil   = "nil"
)

var Keywords = []string{
	TokIn,
	TokTrue,
	TokFalse,
	TokAnd,
	TokOr,
	TokNot,
	TokNil,
	TokWith,
}

var Instructions = []string{
	"for",
	"if",
	"define",
	"extend",
	"slot",
	"template",
	"endif",
	"endfor",
	"enddefine",
	"endtemplate",
	"endslot",
}
