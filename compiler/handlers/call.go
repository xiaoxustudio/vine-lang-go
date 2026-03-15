package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func HandleCallStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.CallExpr)
	_, err := c.Compile(n.Callee)
	if err != nil {
		return nil, err
	}

	_, err = c.Compile(&n.Args)
	if err != nil {
		return nil, err
	}

	c.Emit(bytecode.OpCall, len(n.Args.Arguments))
	return nil, nil
}
