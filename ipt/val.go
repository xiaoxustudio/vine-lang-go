package ipt

import (
	"strconv"
	"vine-lang/lexer"
)

type Token = lexer.Token

// 定义值节点

type Val interface {
	value() any
}
type ValNode struct {
	Val
	Token *Token
}

func (v *ValNode) value() any {
	switch v.Token.Type {
	case lexer.NUMBER:
		Val, _ := strconv.Atoi(v.Token.Value)
		return Val
	case lexer.STRING:
		return v.Token.Value
	case lexer.TRUE:
		return true
	case lexer.FALSE:
		return false
	case lexer.NULL:
		return nil
	default:
		return v.Val
	}
}
