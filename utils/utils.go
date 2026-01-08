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
			return fmt.Sprintf("%s%d%s", Color.Green, current, "\033[0m")
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
