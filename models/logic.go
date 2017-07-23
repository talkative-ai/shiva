package models

type LBlock struct {
	AlwaysExec *uint64         `json:"always"`
	Statements *[][]LStatement `json:"statements"`
}

type LStatement struct {
	Operators *OpArray `json:"conditions"`
	Exec      []uint64 `json:"then"`
}

type VarValMap map[int]interface{}
type OpArray []map[OperatorStr]VarValMap

type OperatorStr string

const (
	OpStrEQ OperatorStr = "eq"
	OpStrLT OperatorStr = "lt"
	OpStrGT OperatorStr = "gt"
	OpStrLE OperatorStr = "le"
	OpStrGE OperatorStr = "ge"
	OpStrNE OperatorStr = "ne"
)

type OperatorInt int8

const (
	OpIntEQ OperatorInt = 1 << iota
	OpIntLT
	OpIntGT
	OpIntLE
	OpIntGE
	OpIntNE
)

func GenerateOperatorStrIntMap() map[OperatorStr]OperatorInt {
	return map[OperatorStr]OperatorInt{
		OpStrEQ: OpIntEQ,
		OpStrLT: OpIntLT,
		OpStrGT: OpIntGT,
		OpStrLE: OpIntLE,
		OpStrGE: OpIntGE,
		OpStrNE: OpIntNE,
	}
}

type StatementInt int8

const (
	StatementIF StatementInt = 1 << iota
	StatementELIF
	StatementELSE
)
