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
	switch args[0].(type) {
	case string:
		return args[0].(string)
	case token.Token:
		return args[0].(token.Token).Value
	default:
		return fmt.Sprint(args...)
	}
}

/* 获取数字和类型 */
func GetNumberAndType(v any) (float64, bool, error) {
	switch val := v.(type) {
	case int64:
		return float64(val), true, nil
	case float64:
		return val, false, nil
	case int:
		return float64(val), true, nil
	case token.Token:
		f, err := strconv.ParseFloat(val.Value, 64)
		if err != nil {
			return 0, false, err
		}
		isInt := f == math.Trunc(f)
		return f, isInt, nil
	default:
		return 0, false, fmt.Errorf("runtime error: unknown operand type %T", v)
	}
}
