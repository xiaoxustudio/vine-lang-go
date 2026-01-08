package utils

import (
	"fmt"
	"math"
	"strconv"
	"unicode"
	"vine-lang/token"
)

func IsIdentifier(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func IsDigit(ch rune) bool {
	return unicode.IsDigit(ch)
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

/* 获取数字和类型 */
func GetNumberAndType(v any) (float64, bool, error) {
	switch val := v.(type) {
	case token.Token:
		f, err := strconv.ParseFloat(val.Value, 64)
		if err != nil {
			return 0, false, err
		}
		isInt := f == math.Trunc(f)
		return f, isInt, nil
	case int64:
		return float64(val), true, nil
	case float64:
		return val, false, nil
	case int:
		return float64(val), true, nil
	default:
		return 0, false, fmt.Errorf("runtime error: unknown operand type %T", v)
	}
}
