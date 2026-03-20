package handlers

import (
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
	"vine-lang/vm/vutils"
)

// HandleEqual 处理相等比较
func HandleEqual(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := vutils.CompareValues(left, right, "==")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleNotEqual 处理不等比较
func HandleNotEqual(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := vutils.CompareValues(left, right, "!=")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleLessThan 处理小于比较
func HandleLessThan(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	stack := v.GetStack()
	sp := v.GetSP()
	right := stack[sp-1]
	left := stack[sp-2]
	sp -= 2
	var result bool
	if leftInt, ok := left.(int64); ok {
		if rightInt, ok := right.(int64); ok {
			result = leftInt < rightInt
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		if rightFloat, ok := right.(float64); ok {
			result = float64(leftInt) < rightFloat
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
	}
	if leftFloat, ok := left.(float64); ok {
		if rightInt, ok := right.(int64); ok {
			result = leftFloat < float64(rightInt)
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		if rightFloat, ok := right.(float64); ok {
			result = leftFloat < rightFloat
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
	}
	if leftStr, ok := left.(string); ok {
		if rightStr, ok := right.(string); ok {
			result = leftStr < rightStr
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
	}
	result = vutils.CompareValues(left, right, "<")
	stack[sp] = result
	v.SetSP(sp + 1)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleLessEqual 处理小于等于比较
func HandleLessEqual(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := vutils.CompareValues(left, right, "<=")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleGreaterThan 处理大于比较
func HandleGreaterThan(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := vutils.CompareValues(left, right, ">")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleGreaterEqual 处理大于等于比较
func HandleGreaterEqual(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := vutils.CompareValues(left, right, ">=")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}
