package compiler

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	"vine-lang/object/store"
	"vine-lang/token"
)

func NewCompiler() *Compiler {
	c := &Compiler{
		handlers:     make(map[ast.NodeType]compileFunc),
		constants:    make([]any, 0),
		instructions: make([]bytecode.Instructions, 0),
		symbolTable:  store.NewStoreObject(),
		scopes:       []*CompilationScope{},
		scopeIndex:   -1,
	}

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

	return c
}
