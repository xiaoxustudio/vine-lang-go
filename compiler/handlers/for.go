package handlers

import (
	"encoding/binary"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func HandleForStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.ForStmt)

	if n.Range != nil {
		arrVarName := "__for_arr_" + n.Init.(*ast.Literal).Value.Value
		arrVarIndex := c.DefineLocal(arrVarName)

		_, err := c.Compile(n.Range)
		if err != nil {
			return nil, err
		}
		c.Emit(bytecode.OpSetLocal, arrVarIndex)

		varName := n.Init.(*ast.Literal).Value.Value
		loopVarIndex := c.DefineLocal(varName)

		lengthVarName := "__for_len_" + varName
		lengthVarIndex := c.DefineLocal(lengthVarName)
		c.Emit(bytecode.OpGetLocal, arrVarIndex)
		c.Emit(bytecode.OpLen)
		c.Emit(bytecode.OpSetLocal, lengthVarIndex)

		indexVarName := "__for_idx_" + varName
		indexVarIndex := c.DefineLocal(indexVarName)
		c.Emit(bytecode.OpConstant, c.AddConstant(int64(0)))
		c.Emit(bytecode.OpSetLocal, indexVarIndex)

		loopStartPos := len(c.CurrentScope().Instructions)

		c.Emit(bytecode.OpGetLocal, indexVarIndex)
		c.Emit(bytecode.OpGetLocal, lengthVarIndex)
		c.Emit(bytecode.OpLessThan)
		jumpIfFalsePos := c.Emit(bytecode.OpJumpIfFalse, 9999)

		c.Emit(bytecode.OpGetLocal, arrVarIndex)
		c.Emit(bytecode.OpGetLocal, indexVarIndex)
		c.Emit(bytecode.OpIndex)
		c.Emit(bytecode.OpSetLocal, loopVarIndex)

		_, err = c.Compile(&n.Body)
		if err != nil {
			return nil, err
		}

		c.Emit(bytecode.OpGetLocal, indexVarIndex)
		c.Emit(bytecode.OpConstant, c.AddConstant(int64(1)))
		c.Emit(bytecode.OpPlus)
		c.Emit(bytecode.OpSetLocal, indexVarIndex)

		currentScope := c.CurrentScope()
		offset := uint16(loopStartPos - len(currentScope.Instructions))
		c.Emit(bytecode.OpLoop, int(offset))

		offset = uint16(len(currentScope.Instructions) - jumpIfFalsePos)
		binary.LittleEndian.PutUint16(currentScope.Instructions[jumpIfFalsePos+1:], offset)

		return nil, nil
	}

	if n.Init != nil {
		_, err := c.Compile(n.Init)
		if err != nil {
			return nil, err
		}
	}

	loopStartPos := len(c.CurrentScope().Instructions)
	var jumpIfFalsePos int

	if n.Value != nil {
		_, err := c.Compile(n.Value)
		if err != nil {
			return nil, err
		}
		jumpIfFalsePos = c.Emit(bytecode.OpJumpIfFalse, 9999)
	}

	parentScope := c.CurrentScope()
	currentEnv := c.CurrentScope().Env
	c.EnterScope(currentEnv)

	newScope := c.CurrentScope()
	for name, index := range parentScope.SymbolTable {
		if _, ok := newScope.SymbolTable[name]; !ok {
			newScope.SymbolTable[name] = index
		}
	}

	_, err := c.Compile(&n.Body)
	if err != nil {
		return nil, err
	}

	loopScope := c.LeaveScope()

	if loopScope != nil && len(loopScope.Instructions) > 0 {
		parentScope.Instructions = append(parentScope.Instructions, loopScope.Instructions...)
		for name, index := range loopScope.SymbolTable {
			if _, ok := parentScope.SymbolTable[name]; !ok {
				parentScope.SymbolTable[name] = index
			}
		}
	}

	if n.Update != nil {
		_, err := c.Compile(n.Update)
		if err != nil {
			return nil, err
		}
	}

	currentScope := c.CurrentScope()
	offset := uint16(loopStartPos - len(currentScope.Instructions))
	c.Emit(bytecode.OpLoop, int(offset))

	if n.Value != nil {
		offset = uint16(len(currentScope.Instructions) - jumpIfFalsePos)
		binary.LittleEndian.PutUint16(currentScope.Instructions[jumpIfFalsePos+1:], offset)
	}

	return nil, nil
}
