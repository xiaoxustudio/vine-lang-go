package handlers

import (
	"vine-lang/ast"
	iface "vine-lang/compiler/interface"
)

func HandleBlockStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.BlockStmt)

	currentScope := c.CurrentScope()
	currentEnv := c.GetScopeEnv(currentScope)
	c.EnterScope(currentEnv)

	newScope := c.CurrentScope()
	for name, index := range currentScope.SymbolTable {
		if _, ok := newScope.SymbolTable[name]; !ok {
			newScope.SymbolTable[name] = index
		}
	}

	for _, s := range n.Body {
		if _, ok := s.(*ast.CommentStmt); ok {
			continue
		}
		_, err := c.Compile(s)
		if err != nil {
			return nil, err
		}
	}

	scope := c.LeaveScope()

	if scope != nil {
		scopeInstructions := c.GetScopeInstructions(scope)
		if len(scopeInstructions) > 0 {
			parentScope := c.CurrentScope()
			if parentScope != nil {
				c.AppendScopeInstructions(parentScope, scopeInstructions)
			}
		}
	}

	return nil, nil
}
