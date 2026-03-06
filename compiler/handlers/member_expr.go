package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleMemberExpr(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.MemberExpr)

	// 检查Object是否为Literal类型（标识符）
	if literal, ok := n.Object.(*ast.Literal); ok && literal.Value.Type == token.IDENT {
		// 直接处理标识符，不进行递归编译
		identName := literal.Value.Value
		pos := c.AddConstant(identName)
		c.Emit(bytecode.OpGetGlobal, pos)
	} else {
		// 对于非标识符情况，正常编译
		_, err := c.Compile(n.Object)
		if err != nil {
			return nil, err
		}
	}

	if n.Computed {
		_, err := c.Compile(n.Property)
		if err != nil {
			return nil, err
		}
		c.Emit(bytecode.OpIndex)
	} else {
		// 处理非计算属性
		if literal, ok := n.Property.(*ast.Literal); ok && literal.Value.Type == token.IDENT {
			// 直接处理标识符属性
			propName := literal.Value.Value
			pos := c.AddConstant(propName)
			c.Emit(bytecode.OpGetMember, pos)
		} else {
			// 对于非标识符情况，正常编译
			_, err := c.Compile(n.Property)
			if err != nil {
				return nil, err
			}
			c.Emit(bytecode.OpGetMember)
		}
	}
	return nil, nil
}
