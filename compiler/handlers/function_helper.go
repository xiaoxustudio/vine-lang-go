package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func compileFunctionBody(c iface.CompilerInterface, body ast.Node, args []ast.Expr, numParams int) (*bytecode.CompiledFunction, error) {
	currentEnv := c.CurrentScope().Env
	c.EnterScope(currentEnv)

	funcScope := c.CurrentScope()

	for _, arg := range args {
		if lit, ok := arg.(*ast.Literal); ok && lit.Value.Type == token.IDENT {
			c.DefineLocal(lit.Value.Value)
		}
	}

	var allInstructions []byte
	beforeCompileScopeIndex := c.GetScopeIndex()

	_, err := c.Compile(body)
	if err != nil {
		return nil, err
	}

	if c.GetScopeIndex() > beforeCompileScopeIndex {
		currentScope := c.CurrentScope()
		if len(currentScope.Instructions) > 0 {
			allInstructions = append(allInstructions, currentScope.Instructions...)
		}
	} else {
		currentScope := c.CurrentScope()
		allInstructions = currentScope.Instructions
	}

	for c.GetScopeIndex() >= 0 && c.CurrentScope() != funcScope.Parent {
		c.LeaveScope()
	}

	return &bytecode.CompiledFunction{
		Instructions:  allInstructions,
		NumLocals:     len(funcScope.SymbolTable),
		NumParameters: numParams,
	}, nil
}
