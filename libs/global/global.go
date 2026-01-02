package global

import "fmt"

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
