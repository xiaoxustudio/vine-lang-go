package compiler

import (
	"fmt"
	"reflect"
	"strings"
	"vine-lang/ast"
	"vine-lang/bytecode"
	"vine-lang/env"
	"vine-lang/object/store"
)

type compileFunc func(node ast.Node) (any, error)

type Compiler struct {
	handlers     map[ast.NodeType]compileFunc // 编译器处理函数
	constants    []any                        // 对应的常量池
	instructions []bytecode.Instructions      // 字节码指令集

	// 作用域 / 符号表
	symbolTable *store.StoreObject
	scopes      []*CompilationScope
	scopeIndex  int
}

type CompilationScope struct {
	instructions bytecode.Instructions
	lastIns      EmittedInstruction
	env          env.Environment
	symbolTable  map[string]int // 局部变量符号表，记录变量名到索引的映射
	parent       *CompilationScope // 父作用域
}

type EmittedInstruction struct {
	Opcode   bytecode.Opcode
	Position int
}

func (c *Compiler) GetConstantRaw() []any {
	return c.constants
}

// EnterScope 进入新的作用域
func (c *Compiler) EnterScope(e env.Environment) *CompilationScope {
	scope := &CompilationScope{
		instructions: bytecode.Instructions{},
		env:          e,
		symbolTable:  make(map[string]int),
	}

	// 如果有父作用域，设置父作用域
	if c.scopeIndex >= 0 {
		scope.parent = c.scopes[c.scopeIndex]
	}

	// 先将作用域添加到数组中
	c.scopes = append(c.scopes, scope)
	// 再更新索引，指向新添加的作用域
	c.scopeIndex = len(c.scopes) - 1
	return scope
}

// LeaveScope 退出当前作用域，返回当前作用域
func (c *Compiler) LeaveScope() *CompilationScope {
	if c.scopeIndex < 0 {
		return nil
	}

	scope := c.scopes[c.scopeIndex]
	c.scopes = c.scopes[:c.scopeIndex]
	c.scopeIndex--
	return scope
}

// NewScope 兼容旧的方法，调用 EnterScope
func (c *Compiler) NewScope(e env.Environment) *CompilationScope {
	return c.EnterScope(e)
}

func (c *Compiler) Compile(node ast.Node) (any, error) {
	return c.CallStmtHandler(node)
}

func (c *Compiler) RegisterStmtHandler(nodeType ast.NodeType, handler compileFunc) {
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
	for i, c := range c.constants {
		// 对于字符串类型，使用字符串比较
		if str, ok := constant.(string); ok {
			if existingStr, ok := c.(string); ok && existingStr == str {
				return i
			}
		} else if reflect.DeepEqual(c, constant) {
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
	currentScope.instructions = append(currentScope.instructions, ins...)
	return len(currentScope.instructions) - len(ins)
}

// DefineLocal 在当前作用域定义局部变量
func (c *Compiler) DefineLocal(name string) int {
	currentScope := c.scopes[c.scopeIndex]
	index := len(currentScope.symbolTable)
	currentScope.symbolTable[name] = index
	return index
}

// ResolveVariable 解析变量，返回变量的类型（全局或局部）和索引
func (c *Compiler) ResolveVariable(name string) (isLocal bool, index int) {
	// 从当前作用域开始，向上查找
	for i := c.scopeIndex; i >= 0; i-- {
		scope := c.scopes[i]
		if idx, ok := scope.symbolTable[name]; ok {
			// 如果在当前作用域或其父作用域中找到，则是局部变量
			// 对于函数作用域，所有参数和局部变量都是局部的
			return true, idx
		}
	}

	// 没有在任何作用域中找到，是全局变量
	return false, -1
}

// CurrentScope 返回当前作用域
func (c *Compiler) CurrentScope() *CompilationScope {
	if c.scopeIndex < 0 {
		return nil
	}
	return c.scopes[c.scopeIndex]
}

// ParentScope 返回当前作用域的父作用域
func (c *Compiler) ParentScope() *CompilationScope {
	if c.scopeIndex < 1 {
		return nil
	}
	return c.scopes[c.scopeIndex-1]
}

func (c Compiler) Dismassemble() string {
	var out strings.Builder
	// 只显示主函数（第一个作用域）的指令
	mainScope := c.scopes[0]
	// 遍历指令，每次读取一条完整的指令
	ip := 0
	for ip < len(mainScope.instructions) {
		// 获取操作码定义
		def, err := bytecode.Lookup(bytecode.Opcode(mainScope.instructions[ip]), c.constants)
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
		// 读取操作数
		operands, _ := bytecode.ReadOperands(mainScope.instructions[ip+1:ip+1+operandsWidth], 0)
		// 生成指令字符串
		ins := bytecode.Make(def.Op, operands...)
		fmt.Fprintf(&out, "%04d %s\n", ip, strings.Join(bytecode.Disassemble(ins, c.constants), " "))
		// 移动到下一条指令
		ip += 1 + operandsWidth
	}
	return out.String()
}

// 创建一个 CompiledFunction 对象
func (c *Compiler) Bytecode() *bytecode.CompiledFunction {
	// 返回主函数（第一个作用域）的指令
	mainScope := c.scopes[0]
	return &bytecode.CompiledFunction{
		Instructions:  mainScope.instructions,
		NumLocals:     0,
		NumParameters: 0,
	}
}
