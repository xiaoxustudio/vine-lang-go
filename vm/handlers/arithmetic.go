package handlers

import (
	"errors"
	"fmt"
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandlePlus 处理加法运算
func HandlePlus(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	stack := v.GetStack()
	sp := v.GetSP()
	right := stack[sp-1]
	left := stack[sp-2]
	sp -= 2
	var result any
	if leftInt, ok := left.(int64); ok {
		if rightInt, ok := right.(int64); ok {
			result = leftInt + rightInt
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		if rightFloat, ok := right.(float64); ok {
			result = float64(leftInt) + rightFloat
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		return nil, errors.New("unsupported types for addition")
	}
	if leftFloat, ok := left.(float64); ok {
		if rightInt, ok := right.(int64); ok {
			result = leftFloat + float64(rightInt)
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		if rightFloat, ok := right.(float64); ok {
			result = leftFloat + rightFloat
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		return nil, errors.New("unsupported types for addition")
	}
	if leftStr, ok := left.(string); ok {
		if rightStr, ok := right.(string); ok {
			result = leftStr + rightStr
		} else {
			result = fmt.Sprintf("%s%v", leftStr, right)
		}
		stack[sp] = result
		v.SetSP(sp + 1)
		frame := v.CurrentFrame()
		frame.Ip += 1
		return result, nil
	}
	return nil, errors.New("unsupported types for addition")
}

// HandleMinus 处理减法运算
func HandleMinus(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()

	var result any
	var err error

	switch left := left.(type) {
	case int64:
		switch right := right.(type) {
		case int64:
			result = left - right
		case float64:
			result = float64(left) - right
		default:
			err = errors.New("unsupported types for subtraction")
		}
	case float64:
		switch right := right.(type) {
		case int64:
			result = left - float64(right)
		case float64:
			result = left - right
		default:
			err = errors.New("unsupported types for subtraction")
		}
	default:
		err = errors.New("unsupported types for subtraction")
	}

	if err != nil {
		return nil, fmt.Errorf("%w: left=%T right=%T", err, left, right)
	}

	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleMul 处理乘法运算
func HandleMul(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	stack := v.GetStack()
	sp := v.GetSP()
	right := stack[sp-1]
	left := stack[sp-2]
	sp -= 2
	var result any
	if leftInt, ok := left.(int64); ok {
		if rightInt, ok := right.(int64); ok {
			result = leftInt * rightInt
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		if rightFloat, ok := right.(float64); ok {
			result = float64(leftInt) * rightFloat
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		return nil, errors.New("unsupported types for multiplication")
	}
	if leftFloat, ok := left.(float64); ok {
		if rightInt, ok := right.(int64); ok {
			result = leftFloat * float64(rightInt)
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		if rightFloat, ok := right.(float64); ok {
			result = leftFloat * rightFloat
			stack[sp] = result
			v.SetSP(sp + 1)
			frame := v.CurrentFrame()
			frame.Ip += 1
			return result, nil
		}
		return nil, errors.New("unsupported types for multiplication")
	}
	return nil, errors.New("unsupported types for multiplication")
}

// HandleDiv 处理除法运算
func HandleDiv(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()

	var result any
	var err error

	// 检查除数是否为零
	switch right := right.(type) {
	case int64:
		if right == 0 {
			return nil, errors.New("division by zero")
		}
	case float64:
		if right == 0 {
			return nil, errors.New("division by zero")
		}
	}

	switch left := left.(type) {
	case int64:
		switch right := right.(type) {
		case int64:
			if left%right == 0 {
				result = left / right
			} else {
				result = float64(left) / float64(right)
			}
		case float64:
			result = float64(left) / right
		default:
			err = errors.New("unsupported types for division")
		}
	case float64:
		switch right := right.(type) {
		case int64:
			result = left / float64(right)
		case float64:
			result = left / right
		default:
			err = errors.New("unsupported types for division")
		}
	default:
		err = errors.New("unsupported types for division")
	}

	if err != nil {
		return nil, err
	}

	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleIncrement 处理自增运算
func HandleIncrement(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	stack := v.GetStack()
	sp := v.GetSP() - 1
	val := stack[sp]
	switch val := val.(type) {
	case int64:
		stack[sp] = val + 1
	case float64:
		stack[sp] = val + 1
	default:
		return nil, errors.New("unsupported types for increment")
	}
	frame := v.CurrentFrame()
	frame.Ip += 1
	return nil, nil
}

// HandleDecrement 处理自减运算
func HandleDecrement(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	val := v.Pop()
	switch val := val.(type) {
	case int64:
		v.Push(val - 1)
	case float64:
		v.Push(val - 1)
	default:
		return nil, errors.New("unsupported types for decrement")
	}
	frame := v.CurrentFrame()
	frame.Ip += 1
	return nil, nil
}
