package compiler

import (
	"fmt"
	"strings"
	"vine-lang/ast"
	"vine-lang/bytecode"
	"vine-lang/env"
	"vine-lang/object/store"
)

type compileFunc func(node ast.Node) error

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
}

type EmittedInstruction struct {
	Opcode   bytecode.Opcode
	Position int
}

func (c *Compiler) GetConstantRaw() []any {
	return c.constants
}

func (c *Compiler) Compile(node ast.Node) error {
	return c.CallStmtHandler(node)
}

func (c *Compiler) RegisterStmtHandler(nodeType ast.NodeType, handler compileFunc) {
	c.handlers[nodeType] = handler
}

func (c *Compiler) CallStmtHandler(node ast.Node) error {
	handler, ok := c.handlers[node.NodeType()]
	if !ok {
		return nil
	}

	return handler(node)
}

// 添加常量，并返回常量的位置
func (c *Compiler) AddConstant(constant any) int {
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
	pos := c.AddInstruction(ins)
	return pos
}

func (c Compiler) Dismassemble() string {
	var out strings.Builder
	for i, ins := range c.instructions {
		fmt.Fprintf(&out, "%04d %s\n", i, strings.Join(bytecode.Disassemble(ins, c.constants), " "))
	}
	return out.String()
}

// 创建一个 CompiledFunction 对象
func (c *Compiler) Bytecode() *bytecode.CompiledFunction {
	// 将所有指令合并为一个指令序列
	instructions := make(bytecode.Instructions, 0)
	for _, ins := range c.instructions {
		instructions = append(instructions, ins...)
	}
	return &bytecode.CompiledFunction{
		Instructions:  instructions,
		NumLocals:     0,
		NumParameters: 0,
	}
}
