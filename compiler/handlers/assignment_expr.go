package handlers

import (
	"fmt"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleAssignmentExpr(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.AssignmentExpr)

	// 复合赋值操作符需要先获取变量的当前值
	switch n.Operator.Type {
	case token.INC_EQ, token.DEC_EQ, token.MUL_EQ, token.DIV_EQ:
		// 先编译左边，获取变量的当前值
		_, err := c.Compile(n.Left)
		if err != nil {
			return nil, err
		}
	}

	// 编译右边的表达式
	_, err := c.Compile(n.Right)
	if err != nil {
		return nil, err
	}

	// 根据操作符类型生成相应的指令
	switch n.Operator.Type {
	case token.ASSIGN:
		// 简单赋值，不需要额外操作
	case token.INC_EQ:
		// += 操作
		c.Emit(bytecode.OpPlus)
	case token.DEC_EQ:
		// -= 操作
		c.Emit(bytecode.OpMinus)
	case token.MUL_EQ:
		// *= 操作
		c.Emit(bytecode.OpMul)
	case token.DIV_EQ:
		// /= 操作
		c.Emit(bytecode.OpDiv)
	default:
		return nil, fmt.Errorf("unknown assignment operator: %s", n.Operator.Type)
	}

	// 处理左边的赋值目标
	switch left := n.Left.(type) {
	case *ast.Literal:
		if left.Value.Type == token.IDENT {
			// 标识符赋值
			varName := left.Value.Value
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
	case *ast.MemberExpr:
		// 成员赋值
		_, err := c.Compile(left.Object)
		if err != nil {
			return nil, err
		}
		if left.Computed {
			// 计算属性赋值 obj[key] = value
			_, err := c.Compile(left.Property)
			if err != nil {
				return nil, err
			}
			// 使用OpSetIndex指令设置索引
			c.Emit(bytecode.OpSetIndex)
		} else {
			// 非计算属性赋值 obj.prop = value
			if literal, ok := left.Property.(*ast.Literal); ok && literal.Value.Type == token.IDENT {
				propName := literal.Value.Value
				pos := c.AddConstant(propName)
				// 使用OpSetMember指令设置成员
				c.Emit(bytecode.OpSetMember, pos)
			}
		}
	default:
		return nil, fmt.Errorf("invalid assignment target")
	}

	return nil, nil
}
