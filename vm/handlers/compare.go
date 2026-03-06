package handlers

import (
	"reflect"
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// compareValues 比较两个值，返回比较结果
func compareValues(left, right any, operator string) bool {
	// 快速路径：使用类型开关
	switch left := left.(type) {
	case int64:
		switch right := right.(type) {
		case int64:
			switch operator {
			case "==":
				return left == right
			case "!=":
				return left != right
			case "<":
				return left < right
			case "<=":
				return left <= right
			case ">":
				return left > right
			case ">=":
				return left >= right
			}
		case float64:
			leftFloat := float64(left)
			switch operator {
			case "==":
				return leftFloat == right
			case "!=":
				return leftFloat != right
			case "<":
				return leftFloat < right
			case "<=":
				return leftFloat <= right
			case ">":
				return leftFloat > right
			case ">=":
				return leftFloat >= right
			}
		}
	case float64:
		switch right := right.(type) {
		case int64:
			rightFloat := float64(right)
			switch operator {
			case "==":
				return left == rightFloat
			case "!=":
				return left != rightFloat
			case "<":
				return left < rightFloat
			case "<=":
				return left <= rightFloat
			case ">":
				return left > rightFloat
			case ">=":
				return left >= rightFloat
			}
		case float64:
			switch operator {
			case "==":
				return left == right
			case "!=":
				return left != right
			case "<":
				return left < right
			case "<=":
				return left <= right
			case ">":
				return left > right
			case ">=":
				return left >= right
			}
		}
	case string:
		if right, ok := right.(string); ok {
			switch operator {
			case "==":
				return left == right
			case "!=":
				return left != right
			case "<":
				return left < right
			case "<=":
				return left <= right
			case ">":
				return left > right
			case ">=":
				return left >= right
			}
		}
	case bool:
		if right, ok := right.(bool); ok {
			switch operator {
			case "==":
				return left == right
			case "!=":
				return left != right
			}
		}
	}

	// 使用 reflect 进行通用比较
	switch operator {
	case "==":
		return reflect.DeepEqual(left, right)
	case "!=":
		return !reflect.DeepEqual(left, right)
	default:
		return false
	}
}

// HandleEqual 处理相等比较
func HandleEqual(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := compareValues(left, right, "==")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleNotEqual 处理不等比较
func HandleNotEqual(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := compareValues(left, right, "!=")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleLessThan 处理小于比较
func HandleLessThan(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := compareValues(left, right, "<")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleLessEqual 处理小于等于比较
func HandleLessEqual(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := compareValues(left, right, "<=")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleGreaterThan 处理大于比较
func HandleGreaterThan(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := compareValues(left, right, ">")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}

// HandleGreaterEqual 处理大于等于比较
func HandleGreaterEqual(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	right := v.Pop()
	left := v.Pop()
	result := compareValues(left, right, ">=")
	v.Push(result)
	frame := v.CurrentFrame()
	frame.Ip += 1
	return result, nil
}
