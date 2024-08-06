package expression

type TokenKind int

func (t TokenKind) String() string {
	if t < 0 || int(t) >= len(TokenKinds) {
		panic("invalid token kind, malformed token table")
	}

	return TokenKinds[t]
}

const (
	TokEmpty TokenKind = iota
	TokEof
	TokIdent
	TokNumeric
	TokString
	TokComma
	TokAt
	TokLparen
	TokRparen
	TokLbrack
	TokRbrack
	TokLt
	TokGt
	TokEq
	TokNeq
	TokLte
	TokGte
	TokBang
	TokPlus
	TokMinus
	TokAsterisk
	TokSlash
	TokModulo
	TokPower
	TokColon
	TokQuestion
	TokDot
	TokOptionalChain
	TokElvis
	TokIn
	TokTrue
	TokNot
	TokFalse
	TokAnd
	TokOr
	TokWith
	TokNil
)

var TokenKinds = []string{
	"empty",
	"eof",
	"identifier",
	"numeric",
	"string",
	"comma",
	"at",
	"lparen",
	"rparen",
	"lbrack",
	"rbrack",
	"lt",
	"gt",
	"eq",
	"neq",
	"lte",
	"gte",
	"bang",
	"plus",
	"minus",
	"asterisk",
	"slash",
	"modulo",
	"power",
	"colon",
	"question",
	"dot",
	"optional_chain",
	"elvis",
	"in",
	"true",
	"not",
	"false",
	"and",
	"or",
	"with",
	"nil",
}

var Keywords = []TokenKind{
	TokIn,
	TokTrue,
	TokFalse,
	TokAnd,
	TokOr,
	TokNot,
	TokNil,
	TokWith,
}
