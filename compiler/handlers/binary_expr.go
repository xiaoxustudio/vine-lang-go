package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
	"vine-lang/utils"
)

func HandleBinaryExpr(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.BinaryExpr)

	// 尝试常量折叠优化
	leftLit, leftIsLit := n.Left.(*ast.Literal)
	rightLit, rightIsLit := n.Right.(*ast.Literal)

	// 如果左右都是数字字面量，进行常量折叠
	if leftIsLit && rightIsLit {
		// 使用utils中的BinaryVal函数进行常量折叠
		result, err := utils.BinaryVal(&leftLit.Value, n.Operator.Type, &rightLit.Value)
		if err == nil {
			// 将计算结果作为常量发出
			pos := c.AddConstant(result)
			c.Emit(bytecode.OpConstant, pos)
			return nil, nil
		}
	}

	// 尝试常量传播优化
	// 如果左操作数是常量变量，直接使用常量值
	if leftLit, ok := n.Left.(*ast.Literal); ok && leftLit.Value.Type == token.IDENT {
		currentScope := c.CurrentScope()
		isLocal, localIndex := c.ResolveVariable(leftLit.Value.Value)
		if isLocal {
			if constantValues, ok := currentScope.ConstantValues[localIndex]; ok {
				// 变量是常量，直接使用常量值
				if rightLit, ok := n.Right.(*ast.Literal); ok {
					// 使用utils中的BinaryVal函数进行常量折叠
					result, err := utils.BinaryVal(constantValues, n.Operator.Type, &rightLit.Value)
					if err == nil {
						// 将计算结果作为常量发出
						pos := c.AddConstant(result)
						c.Emit(bytecode.OpConstant, pos)
						return nil, nil
					}
				}
			}
		}
	}

	// 如果右操作数是常量变量，直接使用常量值
	if rightLit, ok := n.Right.(*ast.Literal); ok && rightLit.Value.Type == token.IDENT {
		currentScope := c.CurrentScope()
		isLocal, localIndex := c.ResolveVariable(rightLit.Value.Value)
		if isLocal {
			if constantValues, ok := currentScope.ConstantValues[localIndex]; ok {
				// 变量是常量，直接使用常量值
				if leftLit, ok := n.Left.(*ast.Literal); ok {
					// 使用utils中的BinaryVal函数进行常量折叠
					result, err := utils.BinaryVal(&leftLit.Value, n.Operator.Type, constantValues)
					if err == nil {
						// 将计算结果作为常量发出
						pos := c.AddConstant(result)
						c.Emit(bytecode.OpConstant, pos)
						return nil, nil
					}
				}
			}
		}
	}

	// 正常编译流程
	_, err := c.Compile(n.Left)
	if err != nil {
		return nil, err
	}
	_, err = c.Compile(n.Right)
	if err != nil {
		return nil, err
	}

	switch n.Operator.Type {
	case token.PLUS:
		c.Emit(bytecode.OpPlus)
	case token.MINUS:
		c.Emit(bytecode.OpMinus)
	case token.MUL:
		c.Emit(bytecode.OpMul)
	case token.DIV:
		c.Emit(bytecode.OpDiv)
	}
	return nil, nil
}
