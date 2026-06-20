package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func patchJumpOffset(scope *iface.CompilationScope, jumpPos int) {
	offset := uint16(len(scope.Instructions) - jumpPos)
	scope.Instructions[jumpPos+1] = byte(offset)
	scope.Instructions[jumpPos+2] = byte(offset >> 8)
}

func HandleIfStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.IfStmt)

	_, err := c.Compile(n.Test)
	if err != nil {
		return nil, err
	}

	jumpIfFalsePos := c.Emit(bytecode.OpJumpIfFalse, 9999)

	_, err = c.Compile(n.Consequent)
	if err != nil {
		return nil, err
	}

	if n.Alternate != nil {
		jumpPos := c.Emit(bytecode.OpJump, 9999)

		patchJumpOffset(c.CurrentScope(), jumpIfFalsePos)

		_, err = c.Compile(n.Alternate)
		if err != nil {
			return nil, err
		}

		patchJumpOffset(c.CurrentScope(), jumpPos)
	} else {
		patchJumpOffset(c.CurrentScope(), jumpIfFalsePos)
	}

	return nil, nil
}
