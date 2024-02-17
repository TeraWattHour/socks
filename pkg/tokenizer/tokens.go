package tokenizer

const (
	TokIdent  = "ident"
	TokNumber = "number"
	TokString = "string"
	TokComma  = "comma"
	TokEnd    = "end"
	TokAt     = "at"

	TokLparen = "lparen"
	TokRparen = "rparen"
	TokLbrack = "lbrack"
	TokRbrack = "rbrack"
	TokLbrace = "lbrace"
	TokRbrace = "rbrace"

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
	TokFloorDiv      = "floor_div"
	TokColon         = "colon"
	TokQuestion      = "question"
	TokDot           = "dot"
	TokOptionalChain = "optional_chain"

	TokFor   = "for"
	TokIn    = "in"
	TokIf    = "if"
	TokTrue  = "true"
	TokNot   = "not"
	TokFalse = "false"
	TokAnd   = "and"
	TokOr    = "or"

	TokExtend   = "extend"
	TokSlot     = "slot"
	TokTemplate = "template"
	TokDefine   = "define"
)

var Keywords = []string{
	TokFor,
	TokIn,
	TokExtend,
	TokSlot,
	TokEnd,
	TokDefine,
	TokTemplate,
	TokIf,
	TokTrue,
	TokFalse,
	TokAnd,
	TokOr,
	TokNot,
}

var Instructions = []string{
	TokFor,
	TokIf,
	TokDefine,
	TokExtend,
	TokSlot,
	TokTemplate,
	"endif",
	"endfor",
	"enddefine",
	"endtemplate",
	"endslot",
}
