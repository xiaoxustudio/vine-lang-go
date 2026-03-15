package handlers

import (
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandleTask 处理任务创建
// OpTask操作码用于创建一个任务对象，该对象包含一个编译好的函数
func HandleTask(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	stack := v.GetStack()
	sp := v.GetSP()

	// 检查栈是否为空
	if sp <= frame.BasePointer {
		return nil, nil // 如果栈为空，直接返回
	}

	// 从栈中弹出编译好的函数对象
	sp--
	fn := stack[sp]

	// 确保弹出的对象是一个编译好的函数
	_, ok := fn.(*bytecode.CompiledFunction)
	if !ok {
		return nil, nil // 如果不是函数，直接返回
	}

	// 创建任务对象
	task := &Task{
		Fn: fn.(*bytecode.CompiledFunction),
	}

	// 将任务对象压入栈
	stack[sp] = task
	v.SetSP(sp + 1)

	// 更新指令指针
	// OpTask指令有2字节的操作数
	frame.Ip += 3

	return task, nil
}

// Task 表示一个任务对象
type Task struct {
	Fn *bytecode.CompiledFunction
}
