package handlers

import (
	"encoding/binary"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func HandleIfStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.IfStmt)

	// 编译条件表达式
	_, err := c.Compile(n.Test)
	if err != nil {
		return nil, err
	}

	// 占位指令
	jumpIfFalsePos := c.Emit(bytecode.OpJumpIfFalse, 9999)

	// 编译 consequent 块
	_, err = c.Compile(n.Consequent)
	if err != nil {
		return nil, err
	}

	if n.Alternate != nil {
		// 占位指令
		jumpPos := c.Emit(bytecode.OpJump, 9999)

		currentScope := c.CurrentScope()
		offset := uint16(len(currentScope.Instructions) - jumpIfFalsePos)
		binary.LittleEndian.PutUint16(currentScope.Instructions[jumpIfFalsePos+1:], offset)

		// 编译 else 或 else if 分支
		_, err = c.Compile(n.Alternate)
		if err != nil {
			return nil, err
		}

		currentScope = c.CurrentScope()
		offset = uint16(len(currentScope.Instructions) - jumpPos)
		binary.LittleEndian.PutUint16(currentScope.Instructions[jumpPos+1:], offset)
	} else {
		currentScope := c.CurrentScope()
		offset := uint16(len(currentScope.Instructions) - jumpIfFalsePos)
		binary.LittleEndian.PutUint16(currentScope.Instructions[jumpIfFalsePos+1:], offset)
	}

	return nil, nil
}
