package handlers

import (
	"fmt"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleCompareExpr(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.CompareExpr)
	_, err := c.Compile(n.Left)
	if err != nil {
		return nil, err
	}
	_, err = c.Compile(n.Right)
	if err != nil {
		return nil, err
	}

	switch n.Operator.Type {
	case token.EQ:
		c.Emit(bytecode.OpEqual)
	case token.NOT_EQ:
		c.Emit(bytecode.OpNotEqual)
	case token.LESS:
		c.Emit(bytecode.OpLessThan)
	case token.LESS_EQ:
		c.Emit(bytecode.OpLessEqual)
	case token.GREATER:
		c.Emit(bytecode.OpGreaterThan)
	case token.GREATER_EQ:
		c.Emit(bytecode.OpGreaterEqual)
	default:
		return nil, fmt.Errorf("unknown comparison operator: %s", n.Operator.Type)
	}
	return nil, nil
}
