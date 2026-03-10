package handlers

import (
	"vine-lang/ast"
	iface "vine-lang/compiler/interface"
)

// HandleBlockStmt 处理块语句
func HandleBlockStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.BlockStmt)

	// 使用当前作用域的环境
	currentScope := c.CurrentScope()
	currentEnv := c.GetScopeEnv(currentScope)
	// 进入新的作用域
	c.EnterScope(currentEnv)

	for _, s := range n.Body {
		// 跳过注释语句
		if _, ok := s.(*ast.CommentStmt); ok {
			continue
		}
		_, err := c.Compile(s)
		if err != nil {
			return nil, err
		}
	}

	// 退出作用域，并保存指令
	scope := c.LeaveScope()

	// 将指令传递给父作用域
	if scope != nil {
		scopeInstructions := c.GetScopeInstructions(scope)
		if len(scopeInstructions) > 0 {
			// 将当前作用域的指令添加到父作用域
			parentScope := c.CurrentScope()
			if parentScope != nil {
				c.AppendScopeInstructions(parentScope, scopeInstructions)
			}
		}
	}

	return nil, nil
}
