package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/token"
)

func HandleFunctionDeclStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.FunctionDecl)

	// 进入新的编译作用域用于函数体
	// 使用当前作用域的环境
	currentEnv := c.CurrentScope().Env
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
	beforeCompileScopeIndex := c.GetScopeIndex()

	// 编译函数体
	_, err := c.Compile(n.Body)
	if err != nil {
		return nil, err
	}

	// 检查是否有新的作用域被创建
	if c.GetScopeIndex() > beforeCompileScopeIndex {
		currentScope := c.CurrentScope()
		if len(currentScope.Instructions) > 0 {
			allInstructions = append(allInstructions, currentScope.Instructions...)
		}
	} else {
		// 没有新的作用域被创建，直接使用当前作用域的指令
		currentScope := c.CurrentScope()
		allInstructions = currentScope.Instructions
	}

	// 退出所有嵌套的作用域，直到回到函数声明之前的作用域
	for c.GetScopeIndex() >= 0 && c.CurrentScope() != funcScope.Parent {
		c.LeaveScope()
	}

	// 创建函数对象
	fn := &bytecode.CompiledFunction{
		Instructions:  allInstructions,
		NumLocals:     len(funcScope.SymbolTable),
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
}
