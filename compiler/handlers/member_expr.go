package handlers

import (
	"fmt"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleMemberExpr(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.MemberExpr)

	// 检查Object是否为Literal类型（标识符）
	if literal, ok := n.Object.(*ast.Literal); ok && literal.Value.Type == token.IDENT {
		// 使用 ResolveVariable 解析变量
		identName := literal.Value.Value
		isLocal, index := c.ResolveVariable(identName)
		if isLocal {
			// 局部变量
			c.Emit(bytecode.OpGetLocal, index)
		} else {
			// 全局变量
			pos := c.AddConstant(identName)
			c.Emit(bytecode.OpGetGlobal, pos)
		}
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
		// 处理非计算属性 xxx,xxx
		if literal, ok := n.Property.(*ast.Literal); ok {
			// 处理字面量属性（包括标识符、数字、字符串）
			var propName string
			switch literal.Value.Type {
			case token.IDENT:
				propName = literal.Value.Value
			case token.INT:
				propName = literal.Value.Value
			case token.STRING:
				propName = literal.Value.Value
			default:
				return nil, fmt.Errorf("member cannot resolve the key %s", literal.Value.Value)
			}
			pos := c.AddConstant(propName)
			c.Emit(bytecode.OpGetMember, pos)
		} else {
			// 对于非字面量情况，正常编译
			_, err := c.Compile(n.Property)
			if err != nil {
				return nil, err
			}
			c.Emit(bytecode.OpGetMember)
		}
	}
	return nil, nil
}
