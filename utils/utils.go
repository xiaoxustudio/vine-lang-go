package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"
	"vine-lang/token"
	"vine-lang/types"
)

func IsIdentifier(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func IsDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

func IsDigitOrDot(ch rune) bool {
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
		case bool:
			return fmt.Sprintf("%s%v%s", Color.Cyan, current, "\033[0m")
		case int, int64:
			return fmt.Sprintf("%s%d%s", Color.Blue, current, "\033[0m")
		case float32, float64:
			return fmt.Sprintf("%s%g%s", Color.Green, current, "\033[0m")
		case token.Token:
			switch current.Type {
			case token.INT, token.FLOAT:
				val, _, err := GetNumberAndType(current)
				if err == nil {
					return TrasformPrintStringWithColor(val)
				}
			case token.NIL:
				return fmt.Sprintf("%s%s%s", Color.Cyan, "nil", "\033[0m")
			case token.TRUE, token.FALSE:
				return fmt.Sprintf("%s%s%s", Color.Cyan, current.Value, "\033[0m")
			}
			return TrasformPrintStringWithColor(current.Value)
		case reflect.Value:
			switch current.Kind() {
			case reflect.Bool:
				return fmt.Sprintf("%s%v%s", Color.Cyan, current.Bool(), "\033[0m")
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return fmt.Sprintf("%s%d%s", Color.Blue, current.Int(), "\033[0m")
			case reflect.Float32, reflect.Float64:
				return fmt.Sprintf("%s%g%s", Color.Green, current.Float(), "\033[0m")
			case reflect.String:
				return fmt.Sprintf("%s%s%s", Color.Yellow, current.String(), "\033[0m")
			case reflect.Slice:
				return fmt.Sprintf("%s%s%s", Color.Yellow, current.String(), "\033[0m")
			case reflect.Map:
				return fmt.Sprintf("%s%s%s", Color.Yellow, current.String(), "\033[0m")
			default:
				return fmt.Sprintf("%s%s%s", Color.Yellow, current.String(), "\033[0m")
			}
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
		switch val.Type {
		case token.INT:
			i, err := val.GetInt()
			if err != nil {
				return 0, types.GNT_INT, err
			}
			return i, types.GNT_INT, nil
		case token.FLOAT:
			f, err := val.GetFloat()
			if err != nil {
				return 0, types.GNT_FLOAT, err
			}
			return f, types.GNT_FLOAT, nil
		case token.STRING:
			return val.Value, types.GNT_STRING, nil
		default:
			return val.Value, types.GNT_Unknown, fmt.Errorf("runtime error: unknown operand type %T", v)
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

func BinaryVal(leftVal any, op token.TokenType, rightVal any) (any, error) {
	left, err := ResolveValue(leftVal)
	if err != nil {
		return false, fmt.Errorf("left param error: %v", err)
	}

	right, err := ResolveValue(rightVal)
	if err != nil {
		return false, fmt.Errorf("right param error: %v", err)
	}

	// 处理类型不匹配的情况
	if left.Kind != right.Kind {
		// 整数和浮点数之间的运算，提升为浮点数
		if (left.Kind == TypeInt64 && right.Kind == TypeFloat64) || (left.Kind == TypeFloat64 && right.Kind == TypeInt64) {
			leftFloat := toFloat64(left.Value)
			rightFloat := toFloat64(right.Value)
			return binaryNumbers(leftFloat, op, rightFloat)
		}
		// 字符串和数字的拼接
		if left.Kind == TypeString && (right.Kind == TypeFloat64 || right.Kind == TypeInt64) {
			rightStr := ""
			if right.Kind == TypeFloat64 {
				rightStr = strconv.FormatFloat(right.Value.(float64), 'f', -1, 64)
			} else {
				rightStr = strconv.FormatInt(right.Value.(int64), 10)
			}
			return binaryStrings(left.Value.(string), op, rightStr)
		} else if right.Kind == TypeString && (left.Kind == TypeFloat64 || left.Kind == TypeInt64) {
			leftStr := ""
			if left.Kind == TypeFloat64 {
				leftStr = strconv.FormatFloat(left.Value.(float64), 'f', -1, 64)
			} else {
				leftStr = strconv.FormatInt(left.Value.(int64), 10)
			}
			return binaryStrings(leftStr, op, right.Value.(string))
		}
		return false, fmt.Errorf("type mismatch: cannot calc %v with %v", left.Kind, right.Kind)
	}

	switch left.Kind {
	case TypeInt64:
		return binaryIntegers(left.Value.(int64), op, right.Value.(int64))
	case TypeFloat64:
		return binaryNumbers(left.Value.(float64), op, right.Value.(float64))
	case TypeString:
		return binaryStrings(left.Value.(string), op, right.Value.(string))
	case TypeBool:
		return false, fmt.Errorf("invalid operator '%v' for booleans", op)
	default:
		return false, fmt.Errorf("unsupported type for comparison")
	}
}

func binaryIntegers(left int64, op token.TokenType, right int64) (any, error) {
	switch op {
	case token.PLUS:
		return left + right, nil
	case token.MINUS:
		return left - right, nil
	case token.MUL:
		return left * right, nil
	case token.DIV:
		// 整数除法，如果不能整除则返回浮点数
		if right == 0 {
			return false, fmt.Errorf("division by zero")
		}
		if left%right == 0 {
			return left / right, nil
		}
		return float64(left) / float64(right), nil
	default:
		return false, fmt.Errorf("invalid operator '%v' for numbers", op)
	}
}

func binaryNumbers(left float64, op token.TokenType, right float64) (any, error) {
	switch op {
	case token.PLUS:
		return left + right, nil
	case token.MINUS:
		return left - right, nil
	case token.MUL:
		return left * right, nil
	case token.DIV:
		if right == 0 {
			return false, fmt.Errorf("division by zero")
		}
		return left / right, nil
	default:
		return false, fmt.Errorf("invalid operator '%v' for numbers", op)
	}
}

func binaryStrings(left string, op token.TokenType, right string) (string, error) {
	switch op {
	case token.PLUS:
		return left + right, nil
	case token.MINUS, token.MUL, token.DIV:
		return "", fmt.Errorf("invalid operator '%v' for strings", op)
	default:
		return "", fmt.Errorf("invalid operator '%v' for strings", op)
	}
}

type ComparableType int

const (
	TypeInt64 ComparableType = iota
	TypeFloat64
	TypeString
	TypeBool
	TypeInvalid
)

type InternalValue struct {
	Kind  ComparableType
	Value any // 实际存储 int64, float64, string 或 bool
}

func CompareVal(leftVal any, op token.TokenType, rightVal any) (bool, error) {
	left, err := ResolveValue(leftVal)
	if err != nil {
		return false, fmt.Errorf("left param error: %v", err)
	}

	right, err := ResolveValue(rightVal)
	if err != nil {
		return false, fmt.Errorf("right param error: %v", err)
	}

	// 处理类型不匹配的情况
	if left.Kind != right.Kind {
		// 整数和浮点数之间的比较，提升为浮点数
		if (left.Kind == TypeInt64 && right.Kind == TypeFloat64) || (left.Kind == TypeFloat64 && right.Kind == TypeInt64) {
			leftFloat := toFloat64(left.Value)
			rightFloat := toFloat64(right.Value)
			return compareNumbers(leftFloat, op, rightFloat)
		}
		return false, fmt.Errorf("type mismatch: cannot compare %v with %v", left.Kind, right.Kind)
	}

	switch left.Kind {
	case TypeInt64:
		return compareIntegers(left.Value.(int64), op, right.Value.(int64))
	case TypeFloat64:
		return compareNumbers(left.Value.(float64), op, right.Value.(float64))
	case TypeString:
		return compareStrings(left.Value.(string), op, right.Value.(string))
	case TypeBool:
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
	case int:
		return InternalValue{Kind: TypeInt64, Value: int64(v)}, nil
	case int64:
		return InternalValue{Kind: TypeInt64, Value: v}, nil
	case int32:
		return InternalValue{Kind: TypeInt64, Value: int64(v)}, nil
	case float64:
		return InternalValue{Kind: TypeFloat64, Value: v}, nil
	case float32:
		return InternalValue{Kind: TypeFloat64, Value: float64(v)}, nil
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
	case token.INT:
		i, err := t.GetInt()
		if err != nil {
			return InternalValue{}, fmt.Errorf("invalid int token '%s'", t.Value)
		}
		return InternalValue{Kind: TypeInt64, Value: i}, nil
	case token.FLOAT:
		f, err := t.GetFloat()
		if err != nil {
			return InternalValue{}, fmt.Errorf("invalid float token '%s'", t.Value)
		}
		return InternalValue{Kind: TypeFloat64, Value: f}, nil

	case token.STRING:
		return InternalValue{Kind: TypeString, Value: t.Value}, nil

	case token.TRUE, token.FALSE:
		b, err := strconv.ParseBool(t.Value)
		if err != nil {
			return InternalValue{}, fmt.Errorf("invalid bool token '%s'", t.Value)
		}
		return InternalValue{Kind: TypeBool, Value: b}, nil

	default:
		return InternalValue{}, fmt.Errorf("unknown token type: %v", t.Type)
	}
}

func compareIntegers(left int64, op token.TokenType, right int64) (bool, error) {
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

func toFloat64(val any) float64 {
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
