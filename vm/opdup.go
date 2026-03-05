package vm

import (
	"vine-lang/bytecode"
)

// OpDupHandler handles the OpDup opcode
func (v *VM) OpDupHandler(op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	// 复制栈顶元素
	if v.sp > 0 {
		v.stack[v.sp] = v.stack[v.sp-1]
		v.sp++
	}
	return nil, nil
}
