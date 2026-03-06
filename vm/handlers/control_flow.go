package handlers

import (
	"encoding/binary"
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandleJump 处理无条件跳转
func HandleJump(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	offset := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	frame.Ip += offset
	return nil, nil
}

// HandleJumpIfFalse 处理条件为假时跳转
func HandleJumpIfFalse(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	condition := v.Pop()
	offset := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	if !isTruthy(condition) {
		frame.Ip += offset
	} else {
		frame.Ip += 3
	}
	return nil, nil
}

// HandleJumpIfTrue 处理条件为真时跳转
func HandleJumpIfTrue(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	condition := v.Pop()
	offset := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	if isTruthy(condition) {
		frame.Ip += offset
	} else {
		frame.Ip += 3
	}
	return nil, nil
}

// HandleLoop 处理循环跳转
func HandleLoop(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	offset := int16(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	frame.Ip += int(offset)
	return nil, nil
}

// HandleBreak 处理跳出循环
func HandleBreak(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	frame.Ip += 1
	return nil, nil
}

// HandleContinue 处理继续循环
func HandleContinue(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	frame.Ip += 1
	return nil, nil
}

// isTruthy 判断值是否为真
func isTruthy(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	case nil:
		return false
	default:
		return true
	}
}
