package utilities

import (
	"encoding/binary"
	"fmt"

	"github.com/artificial-universe-maker/shiva/models"
)

type lConditionaIndex struct {
	Conditional models.LConditional
	Index       int
}

type bSliceIndex struct {
	Bslice []byte
	Index  int
}

func compileHelper(o *models.OpArray) []byte {
	compiled := []byte{}
	OperatorStrIntMap := models.GenerateOperatorStrIntMap()

	for _, OrGroup := range *o {
		var availableConditionals models.OperatorInt
		for c := range OrGroup {
			availableConditionals |= OperatorStrIntMap[c]
		}
		compiled = append(compiled, byte(availableConditionals))
		for _, group := range OrGroup {
			for vr, val := range group {
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(vr))
				compiled = append(compiled, b...)
				switch v := val.(type) {
				case string:
					compiled = append(compiled, uint8(0))
					b := make([]byte, 2)
					binary.LittleEndian.PutUint16(b, uint16(len(v)))
					compiled = append(compiled, b...)
					compiled = append(compiled, []byte(v)...)
					break
				case int:
					compiled = append(compiled, uint8(1))
					b := make([]byte, 4)
					binary.LittleEndian.PutUint32(b, uint32(v))
					compiled = append(compiled, b...)
					break
				}
			}
		}
	}

	return compiled
}

func compileStatement(c *models.LStatement) []byte {
	compiled := []byte{}

	if c.Operators != nil {
		compiled = append(compiled, compileHelper(c.Operators)...)
	}
	compiled = append(compiled, byte(len(c.Exec)))
	for _, execID := range c.Exec {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(execID))
		compiled = append(compiled, b...)
	}
	return compiled
}

func compileStatementArray(c *[]*models.LStatement) []byte {
	compiled := []byte{}

	compiled = append(compiled, uint8(len(*c)))

	for _, stmt := range *c {
		compiled = append(compiled, compileStatement(stmt)...)
	}
	return compiled
}

func conditionalCompile(cidx *lConditionaIndex, c chan bSliceIndex) {
	bslice := []byte{}

	var expectedEnum models.StatementInt

	if cidx.Conditional.StatementIF != nil {
		expectedEnum |= models.StatementIF
	}
	if cidx.Conditional.StatementELIF != nil {
		expectedEnum |= models.StatementELIF
	}
	if cidx.Conditional.StatementELSE != nil {
		expectedEnum |= models.StatementELSE
	}

	bslice = append(bslice, uint8(expectedEnum))

	if expectedEnum&models.StatementIF > 0 {
		bslice = append(bslice, uint8(len(*cidx.Conditional.StatementIF.Operators)))
	}
	if expectedEnum&models.StatementELIF > 0 {
		bslice = append(bslice, uint8(len(*cidx.Conditional.StatementELIF)))
		for _, elif := range *cidx.Conditional.StatementELIF {
			bslice = append(bslice, uint8(len(*elif.Operators)))
		}
	}
	if expectedEnum&models.StatementELSE > 0 {
		if cidx.Conditional.StatementELSE.Operators != nil {
			bslice = append(bslice, uint8(len(*cidx.Conditional.StatementELSE.Operators)))
		} else {
			bslice = append(bslice, 0)
		}
	}

	bslice = append(bslice, compileStatement(cidx.Conditional.StatementIF)...)
	bslice = append(bslice, compileStatementArray(cidx.Conditional.StatementELIF)...)
	bslice = append(bslice, compileStatement(cidx.Conditional.StatementELSE)...)

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(len(bslice)))
	bslice = append(bslice, b...)

	bsliceidx := bSliceIndex{
		Bslice: bslice,
		Index:  cidx.Index,
	}
	c <- bsliceidx
}

func Compile(logic *models.LBlock) []byte {
	compiled := []byte{}

	if logic.AlwaysExec != nil {
		compiled = append(compiled, 1)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, *logic.AlwaysExec)
		compiled = append(compiled, b...)
	}

	compiled = append(compiled, uint8(len(logic.Conditionals)))

	c := make(chan bSliceIndex)
	for idx, conditional := range logic.Conditionals {
		go conditionalCompile(&lConditionaIndex{
			Conditional: conditional,
			Index:       idx,
		}, c)
	}

	newBytes := make([][]byte, len(logic.Conditionals))

	reg := 0
	for bslice := range c {
		fmt.Println("Channel", bslice.Index)
		newBytes[bslice.Index] = bslice.Bslice
		reg++
		if reg == len(logic.Conditionals) {
			close(c)
		}
	}

	for _, bslice := range newBytes {
		compiled = append(compiled, bslice...)
	}

	return compiled
}
