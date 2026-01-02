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
		// 如果是递归计算出来的 int64，直接使用
		return float64(val), true, nil
	case float64:
		// 如果是递归计算出来的 float64，直接使用
		return val, false, nil
	case int:
		// 兼容普通 int
		return float64(val), true, nil
	case lexer.Token:
		// 如果是原始 Token，解析字符串
		f, err := strconv.ParseFloat(val.Value, 64)
		if err != nil {
			return 0, false, err
		}
		// 判断字符串解析出来是否带小数
		isInt := f == math.Trunc(f)
		return f, isInt, nil
	default:
		return 0, false, fmt.Errorf("runtime error: unknown operand type %T", v)
	}
}
