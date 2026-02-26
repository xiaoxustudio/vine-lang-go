package utils

import "vine-lang/token"

func MapKeysToToken(store map[token.Token]any) map[string]token.Token {
	m := make(map[string]token.Token)
	for k := range store {
		m[k.Value] = k
	}
	return m
}
