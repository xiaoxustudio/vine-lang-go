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

	c.RegisterStmtHandler(ast.NodeTypeProgramStmt, func(node ast.Node) error {
		n := node.(*ast.ProgramStmt)
		for _, s := range n.Body {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
		return nil
	})

	c.RegisterStmtHandler(ast.NodeTypeBlockStmt, func(node ast.Node) error {
		n := node.(*ast.BlockStmt)
		for _, s := range n.Body {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
		return nil
	})

	c.RegisterStmtHandler(ast.NodeTypeExpressionStmt, func(node ast.Node) error {
		n := node.(*ast.ExpressionStmt)
		return c.Compile(n.Expression)
	})

	c.RegisterStmtHandler(ast.NodeTypeBinaryExpr, func(node ast.Node) error {
		n := node.(*ast.BinaryExpr)
		err := c.Compile(n.Left)
		if err != nil {
			return err
		}
		err = c.Compile(n.Right)
		if err != nil {
			return err
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
		return nil
	})

	c.RegisterStmtHandler(ast.NodeTypeLiteral, func(node ast.Node) error {
		n := node.(*ast.Literal)
		switch n.Value.Type {
		case token.TRUE:
			c.Emit(bytecode.OpTrue)
		case token.FALSE:
			c.Emit(bytecode.OpFalse)
		case token.FLOAT:
			n, _ := n.Value.GetFloat()
			pos := c.AddConstant(n)
			c.Emit(bytecode.OpConstant, pos)
		case token.INT:
			n, _ := n.Value.GetInt()
			pos := c.AddConstant(n)
			c.Emit(bytecode.OpConstant, pos)
		case token.STRING:
			v := n.Value.Value
			pos := c.AddConstant(v)
			c.Emit(bytecode.OpConstant, pos)
		case token.IDENT:
			// 添加全局变量
			pos := c.AddConstant(n.Value.Value)
			c.Emit(bytecode.OpGetGlobal, pos)
		}
		return nil
	})

	c.RegisterStmtHandler(ast.NodeTypeUnaryExpr, func(node ast.Node) error {
		n := node.(*ast.UnaryExpr)
		err := c.Compile(n.Value)
		if err != nil {
			return err
		}
		switch n.Operator.Type {
		case token.INC:
			c.Emit(bytecode.OpIncrement)
		case token.DEC:
			c.Emit(bytecode.OpDecrement)
		default:
			return nil
		}
		return nil
	})

	c.RegisterStmtHandler(ast.NodeTypeUseSpecifier, func(node ast.Node) error {
		return nil
	})

	c.RegisterStmtHandler(ast.NodeTypeUseDecl, func(node ast.Node) error {
		n := node.(*ast.UseDecl)

		// 获取模块名称
		moduleName := n.Source.Value.Value

		// 获取当前作用域的环境
		currentScope := c.scopes[c.scopeIndex]
		env := currentScope.env

		// 导入模块到环境中
		mod, err := env.ImportModule(moduleName)
		if err != nil {
			return err
		}

		// 根据不同的导入模式处理
		switch n.Mode {
		case token.USE:
			// 1. use "module" - 导入整个模块
			if mod_, ok := mod.(types.LibsModule); ok {
				mod_.ForEach(func(tk token.Token, val any) {
					env.DefineFast(tk.Value, val)
				})
			} else if modObj, ok := mod.(*store.StoreObject); ok {
				modObj.ForEach(func(tk token.Token, val any) {
					env.DefineFast(tk.Value, val)
				})
			}

		case token.AS:
			// 2. use "module" as alias - 导入模块并设置别名
			if len(n.Specifiers) != 1 {
				return fmt.Errorf("use as requires exactly one alias")
			}
			var alias string
			if aliasLit, ok := n.Specifiers[0].(*ast.Literal); ok {
				if aliasLit.Value.Type != token.IDENT {
					return fmt.Errorf("alias must be an identifier")
				}
				alias = aliasLit.Value.Value
			} else {
				return fmt.Errorf("invalid alias specifier")
			}
			env.DefineFast(alias, mod)

		case token.PICK:
			// 3. use "module" pick (fn1, fn2) - 从模块中选择性地导入函数
			if mod, ok := mod.(types.LibsModule); ok {
				for _, sp := range n.Specifiers {
					var fnName, localName string

					if lit, ok := sp.(*ast.Literal); ok {
						if lit.Value.Type != token.IDENT {
							return fmt.Errorf("pick target must be an identifier")
						}
						fnName = lit.Value.Value
						localName = fnName
					} else if us, ok := sp.(*ast.UseSpecifier); ok {
						if us.Remote == nil || us.Remote.Value == nil || us.Remote.Value.Type != token.IDENT {
							return fmt.Errorf("invalid pick specifier")
						}
						fnName = us.Remote.Value.Value
						if us.Local != nil && us.Local.Value != nil {
							if us.Local.Value.Type != token.IDENT {
								return fmt.Errorf("alias must be an identifier")
							}
							localName = us.Local.Value.Value
						} else {
							localName = fnName
						}
					} else {
						return fmt.Errorf("invalid pick specifier")
					}

					if fn, ok := mod.Get(token.Token{Type: token.IDENT, Value: fnName}); ok {
						env.DefineFast(localName, fn)
					} else {
						return fmt.Errorf("function %s not found in module %s", fnName, moduleName)
					}
				}
			} else {
				return fmt.Errorf("invalid module type for pick")
			}

		default:
			return fmt.Errorf("unknown use mode: %s", n.Mode)
		}

		return nil
	})

	c.RegisterStmtHandler(ast.NodeTypeArgsExpr, func(node ast.Node) error {
		n := node.(*ast.ArgsExpr)
		for _, arg := range n.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}
		return nil
	})

	c.RegisterStmtHandler(ast.NodeTypeCallExpr, func(node ast.Node) error {
		n := node.(*ast.CallExpr)
		err := c.Compile(n.Callee)
		if err != nil {
			return err
		}

		err = c.Compile(&n.Args)
		if err != nil {
			return err
		}

		c.Emit(bytecode.OpCall, len(n.Args.Arguments))
		return nil
	})

	return c
}
