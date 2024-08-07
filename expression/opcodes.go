package expression

const (
	OpConstant = iota
	OpJmp
	OpPop

	OpAnd
	OpOr

	OpEq
	OpNeq
	OpGt
	OpGte
	OpLt
	OpLte
	OpIn

	OpAdd
	OpSubtract
	OpMultiply
	OpModulo
	OpDivide
	OpPower

	// OpGet | <<LITERAL_CONST_ID>>
	OpGet
	OpCall
	OpChain
	OpOptionalChain // OpOptionalChain | <<JUMP_BY_IF_NIL>>
	OpElvis         // OpElvis | <<JUMP_BY_IF_NOT_NIL>>
	OpPropertyAccess

	OpTernary

	OpNot
	OpNegate
	OpArray

	OpNil
)
