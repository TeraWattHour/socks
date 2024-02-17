package expression

// BEGIN OPCODES
const (
	OpConstant = iota
	OpEq
	OpNeq
	OpGt
	OpGte
	OpLt
	OpLte
	OpAdd
	OpSubtract
	OpMultiply
	OpDivide
	OpExponent
	OpAnd
	OpOr
	OpModulo
	OpNot
	OpNegate
	OpArrayAccess
	OpChain
	OpOptionalChain
	OpArray
	OpIn
	OpGet
	OpCall
	OpBuiltin1
	OpBuiltin2
	OpBuiltin3

	OpCodeCount
)

// END OPCODES

// BEGIN LOOKUP
var opcodesLookup = map[int]string{
	OpConstant:    "CONSTANT",
	OpEq:          "EQ",
	OpNeq:         "NEQ",
	OpGt:          "GT",
	OpGte:         "GTE",
	OpLt:          "LT",
	OpLte:         "LTE",
	OpAdd:         "ADD",
	OpSubtract:    "SUBTRACT",
	OpMultiply:    "MULTIPLY",
	OpDivide:      "DIVIDE",
	OpExponent:    "EXPONENT",
	OpAnd:         "AND",
	OpOr:          "OR",
	OpNot:         "NOT",
	OpNegate:      "NEGATE",
	OpArrayAccess: "ARRAYACCESS",
	OpChain:       "CHAIN",
	OpArray:       "ARRAY",
	OpIn:          "IN",
	OpGet:         "GET",
	OpCall:        "CALL",
	OpBuiltin1:    "BUILTIN1",
	OpBuiltin2:    "BUILTIN2",
	OpBuiltin3:    "BUILTIN3",
}

// END LOOKUP
