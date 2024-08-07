package expression

const (
	OpConstant = iota
	OpJmp

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

	OpOptionalChain // OpOptionalChain | <<JUMP_BY_IF_NIL>>
	OpElvis         // OpElvis | <<JUMP_BY_IF_NOT_NIL>>
	OpTernary

	OpModulo
	OpNot
	OpNegate
	OpPropertyAccess
	OpChain
	OpArray
	OpIn

	// OpGet | <<LITERAL_CONST_ID>>
	OpGet
	OpCall
	OpPop
)
