package expression

const (
	OpConstant = iota
	OpJmp
	OpOptionalChain
	OpNil
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
	OpPower
	OpAnd
	OpOr
	OpElvis
	OpTernary
	OpModulo
	OpNot
	OpNegate
	OpPropertyAccess
	OpChain
	OpArray
	OpIn
	OpGet
	OpCall
	OpPop

	OpCodeCount
)

var opcodesLookup = map[int]string{
	OpConstant:       "Constant",
	OpJmp:            "Jmp",
	OpPop:            "Pop",
	OpEq:             "Eq",
	OpNeq:            "Neq",
	OpGt:             "Gt",
	OpGte:            "Gte",
	OpLt:             "Lt",
	OpLte:            "Lte",
	OpAdd:            "Add",
	OpSubtract:       "Subtract",
	OpMultiply:       "Multiply",
	OpDivide:         "Divide",
	OpPower:          "Exponent",
	OpElvis:          "Elvis",
	OpTernary:        "Ternary",
	OpAnd:            "And",
	OpOr:             "Or",
	OpNot:            "Not",
	OpNegate:         "Negate",
	OpChain:          "Chain",
	OpOptionalChain:  "OptionalChain",
	OpPropertyAccess: "PropertyAccess",
	OpArray:          "Array",
	OpIn:             "In",
	OpGet:            "Get",
	OpCall:           "Call",
	OpNil:            "Nil",
}
