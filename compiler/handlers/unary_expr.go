package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleUnaryExpr(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.UnaryExpr)

	// 处理后缀自增和自减操作（如 i++, i--）
	if n.IsSuffix {
		// 先编译操作数，获取变量的当前值
		_, err := c.Compile(n.Value)
		if err != nil {
			return nil, err
		}

		// 复制当前值作为返回值（后缀操作返回操作前的值）
		c.Emit(bytecode.OpDup)

		// 发出自增或自减指令
		switch n.Operator.Type {
		case token.INC:
			c.Emit(bytecode.OpIncrement)
		case token.DEC:
			c.Emit(bytecode.OpDecrement)
		default:
			return nil, nil
		}

		// 将结果保存回变量
		if literal, ok := n.Value.(*ast.Literal); ok && literal.Value.Type == token.IDENT {
			varName := literal.Value.Value
			// 检查是否为局部变量
			currentScope := c.CurrentScope()
			if localIndex, ok := currentScope.SymbolTable[varName]; ok {
				// 局部变量
				c.Emit(bytecode.OpSetLocal, localIndex)
			} else {
				// 全局变量
				pos := c.AddConstant(varName)
				c.Emit(bytecode.OpSetGlobal, pos)
			}
		}
		// 弹出保存后的值，保留复制的原始值
		c.Emit(bytecode.OpPop)
		return nil, nil
	}

	// 处理前缀自增和自减操作（如 ++i, --i）
	// 先编译操作数，获取变量的当前值
	_, err := c.Compile(n.Value)
	if err != nil {
		return nil, err
	}

	// 发出自增或自减指令
	switch n.Operator.Type {
	case token.INC:
		c.Emit(bytecode.OpIncrement)
	case token.DEC:
		c.Emit(bytecode.OpDecrement)
	default:
		return nil, nil
	}

	// 将结果保存回变量
	if literal, ok := n.Value.(*ast.Literal); ok && literal.Value.Type == token.IDENT {
		varName := literal.Value.Value
		// 检查是否为局部变量
		currentScope := c.CurrentScope()
		if localIndex, ok := currentScope.SymbolTable[varName]; ok {
			// 局部变量
			c.Emit(bytecode.OpSetLocal, localIndex)
		} else {
			// 全局变量
			pos := c.AddConstant(varName)
			c.Emit(bytecode.OpSetGlobal, pos)
		}
	}

	return nil, nil
}
