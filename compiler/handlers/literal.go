package handlers

import (
	"fmt"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleLiteral(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.Literal)
	var pos int
	switch n.Value.Type {
	case token.TRUE:
		c.Emit(bytecode.OpTrue)
	case token.FALSE:
		c.Emit(bytecode.OpFalse)
	case token.FLOAT:
		n, _ := n.Value.GetFloat()
		pos = c.AddConstant(n)
		c.Emit(bytecode.OpConstant, pos)
	case token.INT:
		n, _ := n.Value.GetInt()
		pos = c.AddConstant(n)
		c.Emit(bytecode.OpConstant, pos)
	case token.STRING:
		v := n.Value.Value
		pos = c.AddConstant(v)
		c.Emit(bytecode.OpConstant, pos)
	case token.IDENT:
		// 使用 ResolveVariable 解析变量
		isLocal, index := c.ResolveVariable(n.Value.Value)
		if isLocal {
			// 局部变量
			c.Emit(bytecode.OpGetLocal, index)
		} else {
			// 全局变量
			pos = c.AddConstant(n.Value.Value)
			c.Emit(bytecode.OpGetGlobal, pos)
		}
	default:
		return nil, fmt.Errorf("unknown literal type: %s", n.Value.Type)
	}
	return pos, nil
}
