package compiler

import (
	"encoding/binary"
	"fmt"
	"vine-lang/ast"
	"vine-lang/bytecode"
	"vine-lang/env"
	"vine-lang/object/store"
	"vine-lang/token"
	"vine-lang/types"
	"vine-lang/utils"
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
			// 跳过注释语句
			if _, ok := s.(*ast.CommentStmt); ok {
				continue
			}
			_, err := c.Compile(s)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeCommentStmt, func(node ast.Node) (any, error) {
		// 注释语句在编译时被忽略，不生成任何字节码
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeBlockStmt, func(node ast.Node) (any, error) {
		n := node.(*ast.BlockStmt)

		// 进入新的作用域
		// 使用当前作用域的环境
		currentEnv := c.scopes[c.scopeIndex].env
		c.EnterScope(currentEnv)

		for _, s := range n.Body {
			// 跳过注释语句
			if _, ok := s.(*ast.CommentStmt); ok {
				continue
			}
			_, err := c.Compile(s)
			if err != nil {
				return nil, err
			}
		}

		// 退出作用域，并保存指令
		scope := c.LeaveScope()

		// 将指令传递给父作用域
		if scope != nil && len(scope.instructions) > 0 {
			// 将当前作用域的指令添加到父作用域
			parentScope := c.CurrentScope()
			if parentScope != nil {
				parentScope.instructions = append(parentScope.instructions, scope.instructions...)
			}
		}

		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeVariableDecl, func(node ast.Node) (any, error) {
		n := node.(*ast.VariableDecl)

		varName := n.Name.Value.Value

		// 检查是否为局部变量
		currentScope := c.scopes[c.scopeIndex]
		if localIndex, ok := currentScope.symbolTable[varName]; ok {
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
				if currentScope.constantValues == nil {
					currentScope.constantValues = make(map[int]any)
				}
				if lit.Value.Type == token.INT {
					val, _ := lit.Value.GetInt()
					currentScope.constantValues[localIndex] = val
				} else {
					val, _ := lit.Value.GetFloat()
					currentScope.constantValues[localIndex] = val
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
	})

	c.RegisterStmtHandler(ast.NodeTypeFunctionDecl, func(node ast.Node) (any, error) {
		n := node.(*ast.FunctionDecl)

		// 进入新的编译作用域用于函数体
		// 使用当前作用域的环境
		currentEnv := c.scopes[c.scopeIndex].env
		c.EnterScope(currentEnv)

		// 获取函数作用域
		funcScope := c.CurrentScope()

		// 处理函数参数，将参数名添加到符号表
		for _, arg := range n.Arguments.Arguments {
			if lit, ok := arg.(*ast.Literal); ok && lit.Value.Type == token.IDENT {
				paramName := lit.Value.Value
				c.DefineLocal(paramName)
			}
		}

		var allInstructions []byte

		// 保存编译前的 scopeIndex
		beforeCompileScopeIndex := c.scopeIndex

		// 编译函数体
		_, err := c.Compile(n.Body)
		if err != nil {
			return nil, err
		}

		// 检查是否有新的作用域被创建
		if c.scopeIndex > beforeCompileScopeIndex {
			currentScope := c.CurrentScope()
			if len(currentScope.instructions) > 0 {
				allInstructions = append(allInstructions, currentScope.instructions...)
			}
		} else {
			// 没有新的作用域被创建，直接使用当前作用域的指令
			currentScope := c.CurrentScope()
			allInstructions = currentScope.instructions
		}

		// 退出所有嵌套的作用域，直到回到函数声明之前的作用域
		for c.scopeIndex >= 0 && c.scopes[c.scopeIndex] != funcScope.parent {
			c.LeaveScope()
		}

		// 创建函数对象
		fn := &bytecode.CompiledFunction{
			Instructions:  allInstructions,
			NumLocals:     len(funcScope.symbolTable),
			NumParameters: len(n.Arguments.Arguments),
		}

		// 将函数对象添加到常量池
		pos := c.AddConstant(fn)

		// 如果函数有名称，则将其定义为全局变量
		if n.ID != nil && n.ID.Value.Type == token.IDENT {
			funcName := n.ID.Value.Value
			namePos := c.AddConstant(funcName)
			c.Emit(bytecode.OpConstant, pos)
			c.Emit(bytecode.OpSetGlobal, namePos)
		} else {
			// 匿名函数，直接压入栈
			c.Emit(bytecode.OpConstant, pos)
		}

		return fn, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeLambdaFunctionDecl, func(node ast.Node) (any, error) {
		n := node.(*ast.LambdaFunctionDecl)

		// 进入新的编译作用域用于函数体
		// 使用当前作用域的环境
		currentEnv := c.scopes[c.scopeIndex].env
		c.EnterScope(currentEnv)

		// 获取函数作用域
		funcScope := c.CurrentScope()

		// 处理函数参数，将参数名添加到符号表
		for _, arg := range n.Args.Arguments {
			if lit, ok := arg.(*ast.Literal); ok && lit.Value.Type == token.IDENT {
				paramName := lit.Value.Value
				c.DefineLocal(paramName)
			}
		}

		var allInstructions []byte

		// 保存编译前的 scopeIndex
		beforeCompileScopeIndex := c.scopeIndex

		// 编译函数体
		_, err := c.Compile(&n.Body)
		if err != nil {
			return nil, err
		}

		// 检查是否有新的作用域被创建
		if c.scopeIndex > beforeCompileScopeIndex {
			currentScope := c.CurrentScope()
			if len(currentScope.instructions) > 0 {
				allInstructions = append(allInstructions, currentScope.instructions...)
			}
		} else {
			// 没有新的作用域被创建，直接使用当前作用域的指令
			currentScope := c.CurrentScope()
			allInstructions = currentScope.instructions
		}

		// 退出所有嵌套的作用域，直到回到函数声明之前的作用域
		for c.scopeIndex >= 0 && c.scopes[c.scopeIndex] != funcScope.parent {
			c.LeaveScope()
		}

		// 创建函数对象
		fn := &bytecode.CompiledFunction{
			Instructions:  allInstructions,
			NumLocals:     len(funcScope.symbolTable),
			NumParameters: len(n.Args.Arguments),
		}

		// 将函数对象添加到常量池
		pos := c.AddConstant(fn)

		// 匿名函数，直接压入栈
		c.Emit(bytecode.OpConstant, pos)

		return fn, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeReturnStmt, func(node ast.Node) (any, error) {
		n := node.(*ast.ReturnStmt)
		if n.Value != nil {
			_, err := c.Compile(n.Value)
			if err != nil {
				return nil, err
			}
		}
		c.Emit(bytecode.OpReturn)
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeIfStmt, func(node ast.Node) (any, error) {
		n := node.(*ast.IfStmt)

		// 编译条件表达式
		_, err := c.Compile(n.Test)
		if err != nil {
			return nil, err
		}

		// 占位指令
		jumpIfFalsePos := c.Emit(bytecode.OpJumpIfFalse, 9999)

		// 编译 consequent 块
		_, err = c.Compile(n.Consequent)
		if err != nil {
			return nil, err
		}

		if n.Alternate != nil {
			// 占位指令
			jumpPos := c.Emit(bytecode.OpJump, 9999)

			currentScope := c.CurrentScope()
			offset := uint16(len(currentScope.instructions) - jumpIfFalsePos)
			binary.LittleEndian.PutUint16(currentScope.instructions[jumpIfFalsePos+1:], offset)

			// 编译 else 或 else if 分支
			_, err = c.Compile(n.Alternate)
			if err != nil {
				return nil, err
			}

			currentScope = c.CurrentScope()
			offset = uint16(len(currentScope.instructions) - jumpPos)
			binary.LittleEndian.PutUint16(currentScope.instructions[jumpPos+1:], offset)
		} else {
			currentScope := c.CurrentScope()
			offset := uint16(len(currentScope.instructions) - jumpIfFalsePos)
			binary.LittleEndian.PutUint16(currentScope.instructions[jumpIfFalsePos+1:], offset)
		}

		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeUseSpecifier, func(node ast.Node) (any, error) {
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeForStmt, func(node ast.Node) (any, error) {
		n := node.(*ast.ForStmt)

		// 检查是否是 for in 循环
		if n.Range != nil {
			// for i in array 形式的循环
			// 编译数组表达式
			_, err := c.Compile(n.Range)
			if err != nil {
				return nil, err
			}

			// 保存循环开始位置
			loopStartPos := len(c.CurrentScope().instructions)

			// 编译循环变量（如果存在）
			if n.Init != nil {
				_, err = c.Compile(n.Init)
				if err != nil {
					return nil, err
				}
			}

			// 发出循环跳转指令（占位）
			loopJumpPos := c.Emit(bytecode.OpLoop, 9999)

			// 为循环体创建新的作用域
			currentEnv := c.scopes[c.scopeIndex].env
			c.EnterScope(currentEnv)

			// 编译循环体
			_, err = c.Compile(&n.Body)
			if err != nil {
				return nil, err
			}

			// 退出循环体作用域，并保存指令
			loopScope := c.LeaveScope()

			// 将循环体指令传递给父作用域
			if loopScope != nil && len(loopScope.instructions) > 0 {
				parentScope := c.CurrentScope()
				if parentScope != nil {
					parentScope.instructions = append(parentScope.instructions, loopScope.instructions...)
				}
			}

			// 发出跳转到循环开始的指令
			currentScope := c.CurrentScope()
			offset := uint16(loopStartPos - len(currentScope.instructions))
			c.Emit(bytecode.OpLoop, int(offset))

			// 修改循环跳转指令，使其跳转到循环结束位置
			offset = uint16(len(currentScope.instructions) - loopJumpPos)
			binary.LittleEndian.PutUint16(currentScope.instructions[loopJumpPos+1:], offset)

			return nil, nil
		} else {
			// for i := 0; i < 10; i++ 形式的循环
			// 编译初始化表达式（在当前作用域中，不进入新作用域）
			if n.Init != nil {
				_, err := c.Compile(n.Init)
				if err != nil {
					return nil, err
				}
			}

			// 保存循环开始位置
			loopStartPos := len(c.CurrentScope().instructions)

			// 声明条件跳转位置变量
			var jumpIfFalsePos int

			// 编译条件表达式
			if n.Value != nil {
				_, err := c.Compile(n.Value)
				if err != nil {
					return nil, err
				}

				// 发出条件跳转指令（如果条件为假，跳出循环）
				// 此时还不知道跳转位置，先发出一个占位指令
				jumpIfFalsePos = c.Emit(bytecode.OpJumpIfFalse, 9999)
			}

			// 保存父作用域的符号表大小
			parentScope := c.CurrentScope()
			parentSymbolCount := len(parentScope.symbolTable)

			// 为循环体创建新的作用域
			currentEnv := c.scopes[c.scopeIndex].env
			c.EnterScope(currentEnv)

			// 编译循环体
			_, err := c.Compile(&n.Body)
			if err != nil {
				return nil, err
			}

			// 退出循环体作用域，并保存指令
			loopScope := c.LeaveScope()

			// 将循环体指令传递给父作用域
			if loopScope != nil && len(loopScope.instructions) > 0 {
				// 调整局部变量索引，加上父作用域的符号表大小
				for i := 0; i < len(loopScope.instructions); i++ {
					op := bytecode.Opcode(loopScope.instructions[i])
					if op == bytecode.OpGetLocal || op == bytecode.OpSetLocal {
						// 读取局部变量索引
						localIndex := int(binary.LittleEndian.Uint16(loopScope.instructions[i+1:]))
						// 调整索引
						newIndex := localIndex + parentSymbolCount
						binary.LittleEndian.PutUint16(loopScope.instructions[i+1:], uint16(newIndex))
					}
				}
				// 将调整后的指令添加到父作用域
				parentScope.instructions = append(parentScope.instructions, loopScope.instructions...)
				// 将循环体的符号表添加到父作用域
				for name, index := range loopScope.symbolTable {
					parentScope.symbolTable[name] = index + parentSymbolCount
				}
			}

			// 编译更新表达式（在循环体之后，但在循环跳转之前）
			if n.Update != nil {
				_, err := c.Compile(n.Update)
				if err != nil {
					return nil, err
				}
			}

			// 发出跳转到循环开始的指令
			currentScope := c.CurrentScope()
			offset := uint16(loopStartPos - len(currentScope.instructions))
			c.Emit(bytecode.OpLoop, int(offset))

			// 修改条件跳转指令，使其跳转到循环结束位置
			if n.Value != nil {
				offset = uint16(len(currentScope.instructions) - jumpIfFalsePos)
				binary.LittleEndian.PutUint16(currentScope.instructions[jumpIfFalsePos+1:], offset)
			}

			return nil, nil
		}
	})

	c.RegisterStmtHandler(ast.NodeTypeBreakStmt, func(node ast.Node) (any, error) {
		c.Emit(bytecode.OpBreak)
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeContinueStmt, func(node ast.Node) (any, error) {
		c.Emit(bytecode.OpContinue)
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

	c.RegisterStmtHandler(ast.NodeTypeExpressionStmt, func(node ast.Node) (any, error) {
		n := node.(*ast.ExpressionStmt)
		return c.Compile(n.Expression)
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
				// 检查是否为局部变量
				currentScope := c.scopes[c.scopeIndex]
				if localIndex, ok := currentScope.symbolTable[varName]; ok {
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
	})

	c.RegisterStmtHandler(ast.NodeTypeUnaryExpr, func(node ast.Node) (any, error) {
		n := node.(*ast.UnaryExpr)

		// 处理后缀自增和自减操作（如 i++, i--）
		if n.IsSuffix {
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
				currentScope := c.scopes[c.scopeIndex]
				if localIndex, ok := currentScope.symbolTable[varName]; ok {
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
			currentScope := c.scopes[c.scopeIndex]
			if localIndex, ok := currentScope.symbolTable[varName]; ok {
				// 局部变量
				c.Emit(bytecode.OpSetLocal, localIndex)
			} else {
				// 全局变量
				pos := c.AddConstant(varName)
				c.Emit(bytecode.OpSetGlobal, pos)
			}
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
		// 编译值
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

	c.RegisterStmtHandler(ast.NodeTypeObjectExpr, func(node ast.Node) (any, error) {
		n := node.(*ast.ObjectExpr)
		// 编译每个属性的键和值
		for _, prop := range n.Properties {
			// 编译键
			if prop.Key.Value.Type == token.IDENT {
				// 标识符键，直接压入字符串
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
	})

	c.RegisterStmtHandler(ast.NodeTypeBinaryExpr, func(node ast.Node) (any, error) {
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
			currentScope := c.scopes[c.scopeIndex]
			isLocal, localIndex := c.ResolveVariable(leftLit.Value.Value)
			if isLocal {
				if constantValues, ok := currentScope.constantValues[localIndex]; ok {
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
			currentScope := c.scopes[c.scopeIndex]
			isLocal, localIndex := c.ResolveVariable(rightLit.Value.Value)
			if isLocal {
				if constantValues, ok := currentScope.constantValues[localIndex]; ok {
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
	})

	c.RegisterStmtHandler(ast.NodeTypeCompareExpr, func(node ast.Node) (any, error) {
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
	})

	return c
}
