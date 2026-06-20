package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func HandleCallTaskStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.CallTaskFn)
	_, err := c.Compile(n.Target.Callee)
	if err != nil {
		return nil, err
	}

	_, err = c.Compile(&n.Target.Args)
	if err != nil {
		return nil, err
	}

	var currentTo *ast.ToExpr = &n.To
	for currentTo != nil {
		_, err := c.Compile(currentTo)
		if err != nil {
			return nil, err
		}
		c.Emit(bytecode.OpTo)
		if currentTo.Next != nil {
			currentTo = currentTo.Next
		} else {
			break
		}
	}

	c.Emit(bytecode.OpCallTask, 0)
	return nil, nil
}
