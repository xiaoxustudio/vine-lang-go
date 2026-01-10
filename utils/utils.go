package utils

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"vine-lang/token"
	"vine-lang/types"
)

func IsIdentifier(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func IsDigit(ch rune) bool {
	return unicode.IsDigit(ch) || ch == '.'
}

func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

/* 转换字符 */
func TrasformPrintString(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	var s string
	for i := range args {
		switch current := args[i].(type) {
		case string:
			return current
		case token.Token:
			return current.Value
		default:
			return fmt.Sprint(args...)
		}
	}
	return s
}

type ColorPrint struct {
	Red    string
	Green  string
	Yellow string
	Blue   string
	Cyan   string
}

var Color = ColorPrint{
	Red:    "\033[31m",
	Green:  "\033[32m",
	Yellow: "\033[33m",
	Blue:   "\033[34m",
	Cyan:   "\033[36m",
}

func TrasformPrintStringWithColor(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	var s string
	for i := range args {
		switch current := args[i].(type) {
		case string:
			return fmt.Sprintf("%s%s%s", Color.Yellow, current, "\033[0m")
		case int, int64:
			return fmt.Sprintf("%s%d%s", Color.Blue, current, "\033[0m")
		case float32, float64:
			return fmt.Sprintf("%s%g%s", Color.Green, current, "\033[0m")
		case token.Token:
			switch current.Type {
			case token.NUMBER:
				val, _, err := GetNumberAndType(current)
				if err == nil {
					return TrasformPrintStringWithColor(val)
				}
			case token.NIL:
				return fmt.Sprintf("%s%s%s", Color.Cyan, "nil", "\033[0m")
			}
			return TrasformPrintStringWithColor(current.Value)
		default:
			return fmt.Sprint(args...)
		}
	}
	return s
}

/* 获取数字和类型 */
func GetNumberAndType(v any) (any, types.GetNumberAndTypeENUM, error) {
	switch val := v.(type) {
	case token.Token:
		if strings.Contains(val.Value, ".") {
			f, err := strconv.ParseFloat(val.Value, 64)
			if err != nil {
				return 0, types.GNT_FLOAT, err
			}
			return f, types.GNT_FLOAT, nil
		} else {
			i, err := strconv.Atoi(val.Value)
			if err != nil {
				return 0, types.GNT_INT, err
			}
			return int64(i), types.GNT_INT, nil
		}
	case int64:
		return val, types.GNT_INT, nil
	case float64:
		return val, types.GNT_FLOAT, nil
	case int:
		return int64(val), types.GNT_INT, nil
	default:
		return -1, types.GNT_Unknown, fmt.Errorf("runtime error: unknown operand type %T", v)
	}
}

type ComparableType int

const (
	TypeFloat64 ComparableType = iota
	TypeString
	TypeBool
	TypeInvalid
)

type InternalValue struct {
	Kind  ComparableType
	Value any // 实际存储 float64, string 或 bool
}

func CompareVal(leftVal any, op token.TokenType, rightVal any) (bool, error) {
	// 第一步：解析左侧值
	left, err := ResolveValue(leftVal)
	if err != nil {
		return false, fmt.Errorf("left param error: %v", err)
	}

	// 第二步：解析右侧值
	right, err := ResolveValue(rightVal)
	if err != nil {
		return false, fmt.Errorf("right param error: %v", err)
	}

	// 第三步：检查类型是否兼容
	// 只有相同类型才能比较 (例如数字不能和字符串比较)
	if left.Kind != right.Kind {
		return false, fmt.Errorf("type mismatch: cannot compare %v with %v", left.Kind, right.Kind)
	}

	// 第四步：根据类型分发比较逻辑
	switch left.Kind {
	case TypeFloat64:
		return compareNumbers(left.Value.(float64), op, right.Value.(float64))
	case TypeString:
		return compareStrings(left.Value.(string), op, right.Value.(string))
	case TypeBool:
		// 布尔值通常只支持相等性判断 (==, !=)
		// 如果你需要 true > false 这样的逻辑，可以在这里扩展
		return compareBools(left.Value.(bool), op, right.Value.(bool))
	default:
		return false, fmt.Errorf("unsupported type for comparison")
	}
}

func ResolveValue(val any) (InternalValue, error) {
	switch v := val.(type) {
	case *token.Token:
		return ParseTokenToValue(v)
	case token.Token:
		return ParseTokenToValue(&v)
	case int, int64, int32:
		return InternalValue{Kind: TypeFloat64, Value: toFloat64(v)}, nil
	case float64, float32:
		return InternalValue{Kind: TypeFloat64, Value: toFloat64(v)}, nil
	case string:
		return InternalValue{Kind: TypeString, Value: v}, nil
	case bool:
		return InternalValue{Kind: TypeBool, Value: v}, nil
	default:
		return InternalValue{}, fmt.Errorf("unsupported input type: %T", val)
	}
}

func ParseTokenToValue(t *token.Token) (InternalValue, error) {
	switch t.Type {
	case token.NUMBER:
		// 尝试将字符串值转为 float64
		f, err := strconv.ParseFloat(t.Value, 64)
		if err != nil {
			return InternalValue{}, fmt.Errorf("invalid number token '%s'", t.Value)
		}
		return InternalValue{Kind: TypeFloat64, Value: f}, nil

	case token.STRING:
		return InternalValue{Kind: TypeString, Value: t.Value}, nil

	case token.TRUE, token.FALSE:
		// 尝试解析布尔字符串 "true" 或 "false"
		b, err := strconv.ParseBool(t.Value)
		if err != nil {
			return InternalValue{}, fmt.Errorf("invalid bool token '%s'", t.Value)
		}
		return InternalValue{Kind: TypeBool, Value: b}, nil

	default:
		return InternalValue{}, fmt.Errorf("unknown token type: %v", t.Type)
	}
}

func compareNumbers(left float64, op token.TokenType, right float64) (bool, error) {
	switch op {
	case token.EQ:
		return left == right, nil
	case token.NOT_EQ:
		return left != right, nil
	case token.LESS:
		return left < right, nil
	case token.LESS_EQ:
		return left <= right, nil
	case token.GREATER:
		return left > right, nil
	case token.GREATER_EQ:
		return left >= right, nil
	default:
		return false, fmt.Errorf("invalid operator '%v' for numbers", op)
	}
}

func compareStrings(left string, op token.TokenType, right string) (bool, error) {
	switch op {
	case token.EQ:
		return left == right, nil
	case token.NOT_EQ:
		return left != right, nil
	case token.LESS:
		return left < right, nil
	case token.LESS_EQ:
		return left <= right, nil
	case token.GREATER:
		return left > right, nil
	case token.GREATER_EQ:
		return left >= right, nil
	default:
		return false, fmt.Errorf("invalid operator '%v' for strings", op)
	}
}

func compareBools(left bool, op token.TokenType, right bool) (bool, error) {
	switch op {
	case token.EQ:
		return left == right, nil
	case token.NOT_EQ:
		return left != right, nil
	default:
		return false, fmt.Errorf("invalid operator '%v' for booleans", op)
	}
}

func toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case float64:
		return v
	case float32:
		return float64(v)
	default:
		return 0
	}
}
