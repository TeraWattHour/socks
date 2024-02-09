package expression

const (
	OpConstant = iota
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterThanOrEqual
	OpLessThan
	OpLessThanOrEqual
	OpAdd
	OpSubtract
	OpMultiply
	OpDivide
	OpExponent
	OpAnd
	OpOr
	OpNot
	OpNegate
	OpGet
	OpCall
	OpBuiltin1
	OpBuiltin2
	OpBuiltin3
	OpArrayAccess
	OpChain
	OpArray
	OpIn
	OpCodeCount
)
