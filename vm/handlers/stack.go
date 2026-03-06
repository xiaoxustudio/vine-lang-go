package handlers

import (
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandlePop 处理弹出栈顶元素
func HandlePop(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	v.Pop()
	frame := v.CurrentFrame()
	frame.Ip += 1
	return nil, nil
}

// HandleDup 处理复制栈顶元素
func HandleDup(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	stack := v.GetStack()
	sp := v.GetSP()
	if sp > 0 {
		stack[sp] = stack[sp-1]
		v.SetSP(sp + 1)
	}
	frame := v.CurrentFrame()
	frame.Ip += 1
	return nil, nil
}
