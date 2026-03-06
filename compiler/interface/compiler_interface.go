package iface

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	"vine-lang/env"
)

// CompilerInterface 编译器接口
type CompilerInterface interface {
	// Compile 编译节点
	Compile(node ast.Node) (any, error)
	// Emit 发出指令
	Emit(op bytecode.Opcode, operands ...int) int
	// AddConstant 添加常量
	AddConstant(constant any) int
	// DefineLocal 定义局部变量
	DefineLocal(name string) int
	// ResolveVariable 解析变量
	ResolveVariable(name string) (isLocal bool, index int)
	// CurrentScope 获取当前作用域
	CurrentScope() *CompilationScope
	// GetIndexScope 获取指定索引的作用域
	GetIndexScope(index int) *CompilationScope
	// GetScopeIndex 获取作用域索引
	GetScopeIndex() int
	// EnterScope 进入作用域
	EnterScope(e env.Environment) *CompilationScope
	// LeaveScope 退出作用域
	LeaveScope() *CompilationScope
	// GetScopeEnv 获取作用域的环境
	GetScopeEnv(scope *CompilationScope) env.Environment
	// GetScopeInstructions 获取作用域的指令
	GetScopeInstructions(scope *CompilationScope) bytecode.Instructions
	// SetScopeInstructions 设置作用域的指令
	SetScopeInstructions(scope *CompilationScope, ins bytecode.Instructions)
	// AppendScopeInstructions 追加指令到作用域
	AppendScopeInstructions(scope *CompilationScope, ins bytecode.Instructions)
	// GetScopeConstantValues 获取作用域的常量值
	GetScopeConstantValues(scope *CompilationScope) map[int]any
	// SetScopeConstantValue 设置作用域的常量值
	SetScopeConstantValue(scope *CompilationScope, index int, value any)
}

// CompilationScope 编译作用域
type CompilationScope struct {
	Instructions   bytecode.Instructions
	LastIns        EmittedInstruction
	Env            env.Environment
	SymbolTable    map[string]int    // 局部变量符号表，记录变量名到索引的映射
	Parent         *CompilationScope // 父作用域
	ConstantValues map[int]any       // 局部变量的常量值，用于常量传播优化
	JumpPositions  []int             // 需要修复的跳转位置列表
	DefaultCasePos int               // default case的位置
}

type EmittedInstruction struct {
	Opcode   bytecode.Opcode
	Position int
}

// CompileFunc 编译函数
type CompileFunc func(node ast.Node) (any, error)
