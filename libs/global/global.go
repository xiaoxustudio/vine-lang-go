package global

import (
	"fmt"
	"math"
	"strconv"
	"vine-lang/lexer"
)

/* 转换字符 */
func TrasformPrintString(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	switch args[0].(type) {
	case string:
		return args[0].(string)
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
	case lexer.Token:
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
