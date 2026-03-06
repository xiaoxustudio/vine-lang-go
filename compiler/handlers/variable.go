package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleVariableDeclStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.VariableDecl)

	varName := n.Name.Value.Value

	// 检查是否为局部变量
	currentScope := c.CurrentScope()
	if localIndex, ok := currentScope.SymbolTable[varName]; ok {
		// 局部变量已存在，直接使用
		_, err := c.Compile(n.Value)
		if err != nil {
			return nil, err
		}
		c.Emit(bytecode.OpSetLocal, localIndex)
	} else {
		// 检查是否为常量表达式
		if lit, ok := n.Value.(*ast.Literal); ok && (lit.Value.Type == token.INT || lit.Value.Type == token.FLOAT) {
			// 定义为新的局部变量
			localIndex = c.DefineLocal(varName)
			// 直接发出常量，不进行编译
			var pos int
			if lit.Value.Type == token.INT {
				val, _ := lit.Value.GetInt()
				pos = c.AddConstant(val)
			} else {
				val, _ := lit.Value.GetFloat()
				pos = c.AddConstant(val)
			}
			c.Emit(bytecode.OpConstant, pos)
			c.Emit(bytecode.OpSetLocal, localIndex)
			// 记录变量的常量值
			if currentScope.ConstantValues == nil {
				currentScope.ConstantValues = make(map[int]any)
			}
			if lit.Value.Type == token.INT {
				val, _ := lit.Value.GetInt()
				currentScope.ConstantValues[localIndex] = val
			} else {
				val, _ := lit.Value.GetFloat()
				currentScope.ConstantValues[localIndex] = val
			}
		} else {
			// 定义为新的局部变量
			localIndex = c.DefineLocal(varName)
			// 编译变量的值
			_, err := c.Compile(n.Value)
			if err != nil {
				return nil, err
			}
			c.Emit(bytecode.OpSetLocal, localIndex)
		}
	}

	return nil, nil
}
