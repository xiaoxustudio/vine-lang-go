package handlers

import (
	"errors"
	"fmt"
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandlePlus 处理加法运算
func HandlePlus(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()

	var result any
	var err error

	switch left := left.(type) {
	case string:
		if rightStr, ok := right.(string); ok {
			result = left + rightStr
		} else {
			result = fmt.Sprintf("%s%v", left, right)
		}
	case int64:
		switch right := right.(type) {
		case int64:
			result = left + right
		case float64:
			result = float64(left) + right
		default:
			err = errors.New("unsupported types for addition")
		}
	case float64:
		switch right := right.(type) {
		case int64:
			result = left + float64(right)
		case float64:
			result = left + right
		default:
			err = errors.New("unsupported types for addition")
		}
	default:
		err = errors.New("unsupported types for addition")
	}

	if err != nil {
		return nil, err
	}

	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
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
		return nil, err
	}

	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleMul 处理乘法运算
func HandleMul(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()

	var result any
	var err error

	switch left := left.(type) {
	case int64:
		switch right := right.(type) {
		case int64:
			result = left * right
		case float64:
			result = float64(left) * right
		default:
			err = errors.New("unsupported types for multiplication")
		}
	case float64:
		switch right := right.(type) {
		case int64:
			result = left * float64(right)
		case float64:
			result = left * right
		default:
			err = errors.New("unsupported types for multiplication")
		}
	default:
		err = errors.New("unsupported types for multiplication")
	}

	if err != nil {
		return nil, err
	}

	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
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
	val := v.Pop()
	switch val := val.(type) {
	case int64:
		v.Push(val + 1)
	case float64:
		v.Push(val + 1)
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
