package compiler

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	"vine-lang/compiler/handlers"
	iface "vine-lang/compiler/interface"
	"vine-lang/env"
	"vine-lang/object/store"
)

func NewCompiler(e *env.Environment) *Compiler {
	c := &Compiler{
		handlers:     make(map[ast.NodeType]iface.CompileFunc),
		constants:    make([]any, 0),
		symbolTable:  store.NewStoreObject(),
		scopes:       []*iface.CompilationScope{},
		scopeIndex:   -1,
	}

	c.NewScope(*e)

	c.RegisterStmtHandler(ast.NodeTypeProgramStmt, func(node ast.Node) (any, error) {
		return handlers.HandleProgramStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeCommentStmt, func(node ast.Node) (any, error) {
		return handlers.HandleCommentStmt(node)
	})

	c.RegisterStmtHandler(ast.NodeTypeBlockStmt, func(node ast.Node) (any, error) {
		return handlers.HandleBlockStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeVariableDecl, func(node ast.Node) (any, error) {
		return handlers.HandleVariableDeclStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeTaskStmt, func(node ast.Node) (any, error) {
		return handlers.HandleTaskStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeFunctionDecl, func(node ast.Node) (any, error) {
		return handlers.HandleFunctionDeclStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeLambdaFunctionDecl, func(node ast.Node) (any, error) {
		return handlers.HandleLambdaFunctionDeclStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeCallTaskFn, func(node ast.Node) (any, error) {
		return handlers.HandleCallTaskStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeReturnStmt, func(node ast.Node) (any, error) {
		return handlers.HandleReturnStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeIfStmt, func(node ast.Node) (any, error) {
		return handlers.HandleIfStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeUseSpecifier, func(node ast.Node) (any, error) {
		return nil, nil
	})

	c.RegisterStmtHandler(ast.NodeTypeExposeStmt, func(node ast.Node) (any, error) {
		return handlers.HandleExposeStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeForStmt, func(node ast.Node) (any, error) {
		return handlers.HandleForStmt(c, node)
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
		return handlers.HandleUseDeclStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeExpressionStmt, func(node ast.Node) (any, error) {
		n := node.(*ast.ExpressionStmt)
		return c.Compile(n.Expression)
	})

	c.RegisterStmtHandler(ast.NodeTypeSwitchCase, func(node ast.Node) (any, error) {
		return handlers.HandleSwitchCaseStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeSwitchStmt, func(node ast.Node) (any, error) {
		return handlers.HandleSwitchStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeToExpr, func(node ast.Node) (any, error) {
		return handlers.HandleToExpr(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeAssignmentExpr, func(node ast.Node) (any, error) {
		return handlers.HandleAssignmentExpr(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeUnaryExpr, func(node ast.Node) (any, error) {
		return handlers.HandleUnaryExpr(c, node)
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
		return handlers.HandleCallStmt(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeMemberExpr, func(node ast.Node) (any, error) {
		return handlers.HandleMemberExpr(c, node)
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
		return handlers.HandleObjectExpr(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeBinaryExpr, func(node ast.Node) (any, error) {
		return handlers.HandleBinaryExpr(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeCompareExpr, func(node ast.Node) (any, error) {
		return handlers.HandleCompareExpr(c, node)
	})

	c.RegisterStmtHandler(ast.NodeTypeLiteral, func(node ast.Node) (any, error) {
		return handlers.HandleLiteral(c, node)
	})

	return c
}
