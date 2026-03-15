package handlers

import (
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandleTo 处理to表达式
// OpTo操作码用于将一个函数标记为to表达式，以便在任务完成后调用
func HandleTo(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	stack := v.GetStack()
	sp := v.GetSP()

	// 从栈中弹出编译好的函数对象
	sp--
	fn, ok := stack[sp].(*bytecode.CompiledFunction)
	if !ok {
		return nil, nil // 如果不是函数，直接返回
	}

	// 将函数对象压入栈
	stack[sp] = fn
	v.SetSP(sp + 1)

	frame.Ip += 1

	return fn, nil
}
