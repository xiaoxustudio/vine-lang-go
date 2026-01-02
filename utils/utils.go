package utils

import (
	"regexp"
)

func IsIdentifier(s string) bool {
	r, err := regexp.MatchString("[\u4e00-\u9fa5a-zA-Z_]", string(s))
	return r && err == nil
}

func IsDigit(s string) bool {
	r, err := regexp.MatchString("[0-9.]+", string(s))
	return r && err == nil
}
