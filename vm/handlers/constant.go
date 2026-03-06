package handlers

import (
	"encoding/binary"
	"fmt"
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandleConstant 处理常量操作码
func HandleConstant(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	constIndex := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	constants := v.GetConstants()
	if constIndex >= len(constants) {
		return nil, fmt.Errorf("constant index %d out of range", constIndex)
	}
	constant := constants[constIndex]
	v.Push(constant)
	frame.Ip += 3
	return constant, nil
}

// HandleTrue 处理true常量
func HandleTrue(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	v.Push(true)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return true, nil
}

// HandleFalse 处理false常量
func HandleFalse(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	v.Push(false)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return false, nil
}
