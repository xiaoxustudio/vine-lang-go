package utils

import "vine-lang/token"

// MapKeysToToken 将map[string]any转换为map[string]token.Token
// 用于从字符串key创建Token映射
func MapKeysToToken(store map[string]any) map[string]token.Token {
	m := make(map[string]token.Token)
	for k := range store {
		m[k] = token.Token{Type: token.IDENT, Value: k}
	}
	return m
}
