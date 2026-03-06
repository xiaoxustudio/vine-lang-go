package vutils

import "reflect"

// IsTruthy 判断值是否为真
func IsTruthy(value any) bool {
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
		// 对于其他类型，假设为真
		return true
	}
}

// CompareValues 比较两个值，返回比较结果
func CompareValues(left, right any, operator string) bool {
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
