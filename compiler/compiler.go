package compiler

import (
	"fmt"
	"reflect"
	"strings"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/env"
	"vine-lang/object/store"
)

type Compiler struct {
	handlers     map[ast.NodeType]iface.CompileFunc // 编译器处理函数
	constants    []any                              // 对应的常量池
	instructions []bytecode.Instructions            // 字节码指令集

	// 作用域 / 符号表
	symbolTable *store.StoreObject
	scopes      []*iface.CompilationScope
	scopeIndex  int
}

func (c *Compiler) GetConstantRaw() []any {
	return c.constants
}

// EnterScope 进入新的作用域
func (c *Compiler) EnterScope(e env.Environment) *iface.CompilationScope {
	scope := &iface.CompilationScope{
		Instructions: bytecode.Instructions{},
		Env:          e,
		SymbolTable:  make(map[string]int),
	}

	// 如果有父作用域，设置父作用域
	if c.scopeIndex >= 0 {
		scope.Parent = c.scopes[c.scopeIndex]
	}

	// 先将作用域添加到数组中
	c.scopes = append(c.scopes, scope)
	// 再更新索引，指向新添加的作用域
	c.scopeIndex = len(c.scopes) - 1
	return scope
}

// LeaveScope 退出当前作用域，返回当前作用域
func (c *Compiler) LeaveScope() *iface.CompilationScope {
	if c.scopeIndex < 0 {
		return nil
	}

	scope := c.scopes[c.scopeIndex]
	c.scopes = c.scopes[:c.scopeIndex]
	c.scopeIndex--
	return scope
}

// NewScope 兼容旧的方法，调用 EnterScope
func (c *Compiler) NewScope(e env.Environment) *iface.CompilationScope {
	return c.EnterScope(e)
}

func (c *Compiler) Compile(node ast.Node) (any, error) {
	return c.CallStmtHandler(node)
}

func (c *Compiler) RegisterStmtHandler(nodeType ast.NodeType, handler iface.CompileFunc) {
	c.handlers[nodeType] = handler
}

func (c *Compiler) CallStmtHandler(node ast.Node) (any, error) {
	handler, ok := c.handlers[node.NodeType()]
	if !ok {
		return nil, fmt.Errorf("no handler for node type : %v", node.NodeType())
	}

	return handler(node)
}

// 添加常量，并返回常量的位置
func (c *Compiler) AddConstant(constant any) int {
	// 检查常量是否已经存在于常量池中
	for i, existing := range c.constants {
		// 对于字符串类型，使用字符串比较
		if str, ok := constant.(string); ok {
			if existingStr, ok := existing.(string); ok && existingStr == str {
				return i
			}
		} else if reflect.DeepEqual(existing, constant) {
			return i
		}
	}
	// 常量不存在，添加到常量池
	c.constants = append(c.constants, constant)
	return len(c.constants) - 1
}

// 添加指令，并返回指令的位置
func (c *Compiler) AddInstruction(ins bytecode.Instructions) int {
	c.instructions = append(c.instructions, ins)
	return len(c.instructions) - 1
}

func (c *Compiler) Emit(op bytecode.Opcode, operands ...int) int {
	ins := bytecode.Make(op, operands...)
	// 将指令添加到当前作用域的 instructions 中
	currentScope := c.scopes[c.scopeIndex]
	currentScope.Instructions = append(currentScope.Instructions, ins...)
	return len(currentScope.Instructions) - len(ins)
}

// DefineLocal 在当前作用域定义局部变量
func (c *Compiler) DefineLocal(name string) int {
	currentScope := c.scopes[c.scopeIndex]
	index := len(currentScope.SymbolTable)
	currentScope.SymbolTable[name] = index
	return index
}

// ResolveVariable 解析变量，返回变量的类型（全局或局部）和索引
func (c *Compiler) ResolveVariable(name string) (isLocal bool, index int) {
	// 从当前作用域开始，向上查找
	for i := c.scopeIndex; i >= 0; i-- {
		scope := c.scopes[i]
		if idx, ok := scope.SymbolTable[name]; ok {
			// 如果在当前作用域或其父作用域中找到，则是局部变量
			// 对于函数作用域，所有参数和局部变量都是局部的
			return true, idx
		}
	}

	// 没有在任何作用域中找到，是全局变量
	return false, -1
}

// CurrentScope 返回当前作用域
func (c *Compiler) CurrentScope() *iface.CompilationScope {
	if c.scopeIndex < 0 {
		return nil
	}
	return c.scopes[c.scopeIndex]
}

func (c *Compiler) GetScopeIndex() int {
	return c.scopeIndex
}

// ParentScope 返回当前作用域的父作用域
func (c *Compiler) ParentScope() *iface.CompilationScope {
	if c.scopeIndex < 1 {
		return nil
	}
	return c.scopes[c.scopeIndex-1]
}

func (c *Compiler) GetIndexScope(index int) *iface.CompilationScope {
	return c.scopes[index]
}

func (c Compiler) Dismassemble() string {
	var out strings.Builder
	// 只显示主函数（第一个作用域）的指令
	mainScope := c.scopes[0]
	// 遍历指令，每次读取一条完整的指令
	ip := 0
	for ip < len(mainScope.Instructions) {
		// 获取操作码定义
		def, err := bytecode.Lookup(bytecode.Opcode(mainScope.Instructions[ip]), c.constants)
		if err != nil {
			fmt.Fprintf(&out, "%04d ERROR: %s\n", ip, err.Error())
			ip++
			continue
		}
		// 计算操作数宽度
		operandsWidth := 0
		for _, w := range def.OperandWidths {
			operandsWidth += w
		}
		// 检查是否有足够的字节来读取操作数
		if ip+1+operandsWidth > len(mainScope.Instructions) {
			fmt.Fprintf(&out, "%04d ERROR: insufficient operands\n", ip)
			ip++
			continue
		}
		// 读取操作数
		operands, _ := bytecode.ReadOperands(mainScope.Instructions[ip+1:ip+1+operandsWidth], 0)
		// 生成指令字符串
		ins := bytecode.Make(def.Op, operands...)
		fmt.Fprintf(&out, "%04d %s\n", ip, strings.Join(bytecode.Disassemble(ins, c.constants), " "))
		// 移动到下一条指令
		ip += 1 + operandsWidth
	}
	return out.String()
}

// GetScopeEnv 获取作用域的环境
func (c *Compiler) GetScopeEnv(scope *iface.CompilationScope) env.Environment {
	return scope.Env
}

// GetScopeInstructions 获取作用域的指令
func (c *Compiler) GetScopeInstructions(scope *iface.CompilationScope) bytecode.Instructions {
	return scope.Instructions
}

// SetScopeInstructions 设置作用域的指令
func (c *Compiler) SetScopeInstructions(scope *iface.CompilationScope, ins bytecode.Instructions) {
	scope.Instructions = ins
}

// AppendScopeInstructions 追加指令到作用域
func (c *Compiler) AppendScopeInstructions(scope *iface.CompilationScope, ins bytecode.Instructions) {
	scope.Instructions = append(scope.Instructions, ins...)
}

// GetScopeConstantValues 获取作用域的常量值
func (c *Compiler) GetScopeConstantValues(scope *iface.CompilationScope) map[int]any {
	return scope.ConstantValues
}

// SetScopeConstantValue 设置作用域的常量值
func (c *Compiler) SetScopeConstantValue(scope *iface.CompilationScope, index int, value any) {
	if scope.ConstantValues == nil {
		scope.ConstantValues = make(map[int]any)
	}
	scope.ConstantValues[index] = value
}

// 创建一个 CompiledFunction 对象
func (c *Compiler) Bytecode() *bytecode.CompiledFunction {
	// 返回主函数（第一个作用域）的指令
	mainScope := c.scopes[0]
	return &bytecode.CompiledFunction{
		Instructions:  mainScope.Instructions,
		NumLocals:     0,
		NumParameters: 0,
	}
}
