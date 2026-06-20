package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func HandleLambdaFunctionDeclStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.LambdaFunctionDecl)

	fn, err := compileFunctionBody(c, &n.Body, n.Args.Arguments, len(n.Args.Arguments))
	if err != nil {
		return nil, err
	}

	pos := c.AddConstant(fn)
	c.Emit(bytecode.OpConstant, pos)

	return fn, nil
}
