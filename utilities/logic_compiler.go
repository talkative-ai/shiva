package utilities

import (
	"github.com/artificial-universe-maker/shiva/models"
)

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
				compiled = append(compiled, byte(vr))
				switch v := val.(type) {
				case string:
					compiled = append(compiled, uint8(0))
					compiled = append(compiled, byte(uint16(len(v))))
					compiled = append(compiled, []byte(v)...)
					break
				case int:
					compiled = append(compiled, uint8(1))
					compiled = append(compiled, byte(uint32(v)))
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
	compiled = append(compiled, uint8(len(c.Exec)))
	for _, execID := range c.Exec {
		compiled = append(compiled, byte(execID))
	}
	return compiled
}

func compileStatementArray(c *[]models.LStatement) []byte {
	compiled := []byte{}

	compiled = append(compiled, byte(len(*c)))

	for _, stmt := range *c {
		compiled = append(compiled, compileStatement(&stmt)...)
	}
	return compiled
}

func Compile(logic *models.LBlock) []byte {
	compiled := []byte{}

	compiled = append(compiled, uint8(len(logic.AlwaysExec)))

	for _, id := range logic.AlwaysExec {
		compiled = append(compiled, byte(id))
	}

	compiled = append(compiled, uint8(len(logic.Conditionals)))

	for _, conditional := range logic.Conditionals {
		var expectedEnum models.StatementInt

		if conditional.StatementIF != nil {
			expectedEnum |= models.StatementIF
		}
		if conditional.StatementELIF != nil {
			expectedEnum |= models.StatementELIF
		}
		if conditional.StatementELSE != nil {
			expectedEnum |= models.StatementELSE
		}

		compiled = append(compiled, uint8(expectedEnum))

		if expectedEnum&models.StatementIF > 0 {
			compiled = append(compiled, uint8(len(*conditional.StatementIF.Operators)))
		}
		if expectedEnum&models.StatementELIF > 0 {
			compiled = append(compiled, uint8(len(*conditional.StatementELIF)))
			for _, elif := range *conditional.StatementELIF {
				compiled = append(compiled, uint8(len(*elif.Operators)))
			}
		}
		if expectedEnum&models.StatementELSE > 0 {
			if conditional.StatementELSE.Operators != nil {
				compiled = append(compiled, uint8(len(*conditional.StatementELSE.Operators)))
			} else {
				compiled = append(compiled, 0)
			}
		}

		compiled = append(compiled, compileStatement(conditional.StatementIF)...)
		compiled = append(compiled, compileStatementArray(conditional.StatementELIF)...)
		compiled = append(compiled, compileStatement(conditional.StatementELSE)...)
	}

	return compiled
}
