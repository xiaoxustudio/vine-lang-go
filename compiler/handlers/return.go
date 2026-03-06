package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func HandleReturnStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.ReturnStmt)
	if n.Value != nil {
		_, err := c.Compile(n.Value)
		if err != nil {
			return nil, err
		}
	}
	c.Emit(bytecode.OpReturn)
	return nil, nil
}
