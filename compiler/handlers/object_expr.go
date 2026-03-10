package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleObjectExpr(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.ObjectExpr)
	// 编译每个属性的键和值
	for _, prop := range n.Properties {
		// 编译键
		if prop.Key.Value.Type == token.IDENT {
			// 标识符键，直接压入字符串
			// 如果不处理会尝试解析为标识符（同理它会进行获取变量的操作），这是不合理的
			keyName := prop.Key.Value.Value
			pos := c.AddConstant(keyName)
			c.Emit(bytecode.OpConstant, pos)
		} else {
			// 字面量键，编译键表达式
			_, err := c.Compile(prop.Key)
			if err != nil {
				return nil, err
			}
		}
		// 编译值
		_, err := c.Compile(prop.Value)
		if err != nil {
			return nil, err
		}
	}
	c.Emit(bytecode.OpObject, len(n.Properties))
	return nil, nil
}
