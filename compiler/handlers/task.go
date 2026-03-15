package handlers

import (
	"fmt"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleTaskStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.TaskStmt)

	currentEnv := c.CurrentScope().Env
	c.EnterScope(currentEnv)

	// 目前只支持纯函数任务
	fn, err := c.Compile(&n.Fn)
	if err != nil {
		return nil, err
	}

	c.LeaveScope()

	// 将函数对象添加到常量池
	compiledFn, ok := fn.(*bytecode.CompiledFunction)
	if !ok {
		return nil, fmt.Errorf("task function must be a compiled function")
	}
	pos := c.AddConstant(compiledFn)

	// 如果任务有名称，则将其定义为全局变量
	if n.Fn.ID != nil && n.Fn.ID.Value.Type == token.IDENT {
		taskName := n.Fn.ID.Value.Value
		namePos := c.AddConstant(taskName)
		c.Emit(bytecode.OpConstant, pos)
		c.Emit(bytecode.OpTask, 0)
		c.Emit(bytecode.OpSetGlobal, namePos)
	} else {
		// 匿名任务，直接压入栈
		c.Emit(bytecode.OpConstant, pos)
		c.Emit(bytecode.OpTask, 0)
	}

	return nil, nil
}
