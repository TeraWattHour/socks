package expression

// BEGIN OPCODES
const (
	OpConstant = iota
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
	OpConstant:      "Constant",
	OpEq:            "Eq",
	OpNeq:           "Neq",
	OpGt:            "Gt",
	OpGte:           "Gte",
	OpLt:            "Lt",
	OpLte:           "Lte",
	OpAdd:           "Add",
	OpSubtract:      "Subtract",
	OpMultiply:      "Multiply",
	OpDivide:        "Divide",
	OpPower:         "Exponent",
	OpElvis:         "Elvis",
	OpAnd:           "And",
	OpOr:            "Or",
	OpNot:           "Not",
	OpNegate:        "Negate",
	OpArrayAccess:   "FieldAccess",
	OpChain:         "Chain",
	OpOptionalChain: "OptionalChain",
	OpArray:         "Array",
	OpIn:            "In",
	OpGet:           "Get",
	OpCall:          "Call",
	OpBuiltin1:      "Builtin1",
	OpBuiltin2:      "Builtin2",
	OpBuiltin3:      "Builtin3",
	OpNil:           "Nil",
}

// END LOOKUP
