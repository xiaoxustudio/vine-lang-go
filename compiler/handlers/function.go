package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleFunctionDeclStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.FunctionDecl)

	fn, err := compileFunctionBody(c, n.Body, n.Arguments.Arguments, len(n.Arguments.Arguments))
	if err != nil {
		return nil, err
	}

	pos := c.AddConstant(fn)

	if n.ID != nil && n.ID.Value.Type == token.IDENT {
		funcName := n.ID.Value.Value
		namePos := c.AddConstant(funcName)
		c.Emit(bytecode.OpConstant, pos)
		c.Emit(bytecode.OpSetGlobal, namePos)
	} else {
		c.Emit(bytecode.OpConstant, pos)
	}

	return fn, nil
}
