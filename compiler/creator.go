package compiler

import (
	"fmt"
	"vine-lang/ast"
	"vine-lang/bytecode"
	"vine-lang/env"
	"vine-lang/object/store"
	"vine-lang/token"
	"vine-lang/types"
)

func NewCompiler(e *env.Environment) *Compiler {
	c := &Compiler{
		handlers:     make(map[ast.NodeType]compileFunc),
		constants:    make([]any, 0),
		instructions: make([]bytecode.Instructions, 0),
		symbolTable:  store.NewStoreObject(),
		scopes:       []*CompilationScope{},
		scopeIndex:   -1,
	}

	c.NewScope(*e)

	c.RegisterStmtHandler(ast.NodeTypeProgramStmt, func(node ast.Node) (any, error) {
		n := node.(*ast.ProgramStmt)
		for _, s := range n.Body {
			_, err := c.Compile(s)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeBlockStmt, func(node ast.Node) (any, error) {
		n := node.(*ast.BlockStmt)
		for _, s := range n.Body {
			_, err := c.Compile(s)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeExpressionStmt, func(node ast.Node) (any, error) {
		n := node.(*ast.ExpressionStmt)
		return c.Compile(n.Expression)
	})

	c.RegisterStmtHandler(ast.NodeTypeBinaryExpr, func(node ast.Node) (any, error) {
		n := node.(*ast.BinaryExpr)
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
	})

	c.RegisterStmtHandler(ast.NodeTypeLiteral, func(node ast.Node) (any, error) {
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
			// 添加全局变量
			pos = c.AddConstant(n.Value.Value)
			c.Emit(bytecode.OpGetGlobal, pos)
		default:
			return nil, fmt.Errorf("unknown literal type: %s", n.Value.Type)
		}
		return pos, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeVariableDecl, func(node ast.Node) (any, error) {
		n := node.(*ast.VariableDecl)

		// 编译变量的值
		_, err := c.Compile(n.Value)
		if err != nil {
			return nil, err
		}

		varName := n.Name.Value.Value
		pos := c.AddConstant(varName)

		if n.IsConst {
			c.Emit(bytecode.OpSetConst, pos)
		} else {
			c.Emit(bytecode.OpSetGlobal, pos)
		}

		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeAssignmentExpr, func(node ast.Node) (any, error) {
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
				pos := c.AddConstant(varName)
				c.Emit(bytecode.OpSetGlobal, pos)
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
	})

	c.RegisterStmtHandler(ast.NodeTypeUnaryExpr, func(node ast.Node) (any, error) {
		n := node.(*ast.UnaryExpr)
		_, err := c.Compile(n.Value)
		if err != nil {
			return nil, err
		}
		switch n.Operator.Type {
		case token.INC:
			c.Emit(bytecode.OpIncrement)
		case token.DEC:
			c.Emit(bytecode.OpDecrement)
		default:
			return nil, nil
		}
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeUseSpecifier, func(node ast.Node) (any, error) {
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeUseDecl, func(node ast.Node) (any, error) {
		n := node.(*ast.UseDecl)

		// 获取模块名称
		moduleName := n.Source.Value.Value

		// 获取当前作用域的环境
		currentScope := c.scopes[c.scopeIndex]
		env := currentScope.env

		// 导入模块到环境中
		mod, err := env.ImportModule(moduleName)
		if err != nil {
			return nil, err
		}

		// 根据不同的导入模式处理
		switch n.Mode {
		case token.USE:
			// 1. use "module" - 导入整个模块
			env.DefineFast(moduleName, mod)

		case token.AS:
			// 2. use "module" as alias - 导入模块并设置别名
			if len(n.Specifiers) != 1 {
				return nil, fmt.Errorf("use as requires exactly one alias")
			}
			var alias string
			if aliasLit, ok := n.Specifiers[0].(*ast.Literal); ok {
				if aliasLit.Value.Type != token.IDENT {
					return nil, fmt.Errorf("alias must be an identifier")
				}
				alias = aliasLit.Value.Value
			} else {
				return nil, fmt.Errorf("invalid alias specifier")
			}
			env.DefineFast(alias, mod)

		case token.PICK:
			// 3. use "module" pick (fn1, fn2) - 从模块中选择性地导入函数
			if mod, ok := mod.(types.LibsModule); ok {
				for _, sp := range n.Specifiers {
					var fnName, localName string

					if lit, ok := sp.(*ast.Literal); ok {
						if lit.Value.Type != token.IDENT {
							return nil, fmt.Errorf("pick target must be an identifier")
						}
						fnName = lit.Value.Value
						localName = fnName
					} else if us, ok := sp.(*ast.UseSpecifier); ok {
						if us.Remote == nil || us.Remote.Value == nil || us.Remote.Value.Type != token.IDENT {
							return nil, fmt.Errorf("invalid pick specifier")
						}
						fnName = us.Remote.Value.Value
						if us.Local != nil && us.Local.Value != nil {
							if us.Local.Value.Type != token.IDENT {
								return nil, fmt.Errorf("alias must be an identifier")
							}
							localName = us.Local.Value.Value
						} else {
							localName = fnName
						}
					} else {
						return nil, fmt.Errorf("invalid pick specifier")
					}

					if fn, ok := mod.Get(token.Token{Type: token.IDENT, Value: fnName}); ok {
						env.DefineFast(localName, fn)
					} else {
						return nil, fmt.Errorf("function %s not found in module %s", fnName, moduleName)
					}
				}
			} else {
				return nil, fmt.Errorf("invalid module type for pick")
			}

		default:
			return nil, fmt.Errorf("unknown use mode: %s", n.Mode)
		}

		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeArgsExpr, func(node ast.Node) (any, error) {
		n := node.(*ast.ArgsExpr)
		for _, arg := range n.Arguments {
			_, err := c.Compile(arg)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeCallExpr, func(node ast.Node) (any, error) {
		n := node.(*ast.CallExpr)
		_, err := c.Compile(n.Callee)
		if err != nil {
			return nil, err
		}

		_, err = c.Compile(&n.Args)
		if err != nil {
			return nil, err
		}

		c.Emit(bytecode.OpCall, len(n.Args.Arguments))
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeMemberExpr, func(node ast.Node) (any, error) {
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
	})

	c.RegisterStmtHandler(ast.NodeTypeProperty, func(node ast.Node) (any, error) {
		n := node.(*ast.Property)
		// 对于数组元素，只需要编译 Value，不需要编译 Key（索引）
		_, err := c.Compile(n.Value)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeArrayExpr, func(node ast.Node) (any, error) {
		n := node.(*ast.ArrayExpr)
		for _, elem := range n.Items {
			_, err := c.Compile(elem)
			if err != nil {
				return nil, err
			}
		}
		c.Emit(bytecode.OpArray, len(n.Items))
		return nil, nil
	})

	return c
}
