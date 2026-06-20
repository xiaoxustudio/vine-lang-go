package handlers

import (
	"fmt"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

var _ = iface.CompilerInterface(nil)

func HandleTaskStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.TaskStmt)

	fn, err := c.Compile(&n.Fn)
	if err != nil {
		return nil, err
	}

	compiledFn, ok := fn.(*bytecode.CompiledFunction)
	if !ok {
		return nil, fmt.Errorf("task function must be a compiled function")
	}
	pos := c.AddConstant(compiledFn)

	if n.Fn.ID != nil && n.Fn.ID.Value.Type == token.IDENT {
		taskName := n.Fn.ID.Value.Value
		namePos := c.AddConstant(taskName)
		c.Emit(bytecode.OpConstant, pos)
		c.Emit(bytecode.OpTask, 0)
		c.Emit(bytecode.OpSetGlobal, namePos)
	} else {
		c.Emit(bytecode.OpConstant, pos)
		c.Emit(bytecode.OpTask, 0)
	}

	return nil, nil
}
