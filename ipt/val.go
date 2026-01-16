package ipt

import (
	"vine-lang/ast"
	"vine-lang/token"
)

// 定义值节点

type Val interface {
	value() any
}
type ValNode struct {
	Val
	Token  *token.Token
	Object any
}

func (v *ValNode) value() any {
	switch v.Token.Type {
	case token.INT:
		val, _ := v.Token.GetInt()
		return val
	case token.FLOAT:
		val, _ := v.Token.GetFloat()
		return val
	case token.STRING:
		return v.Token.Value
	case token.TRUE:
		return true
	case token.FALSE:
		return false
	case token.NIL:
		return nil
	default:
		return v.Val
	}
}

type FunctionLikeValNode struct {
	Val
	Token     *token.Token
	Arguments []ast.Expr
	Body      *ast.BlockStmt
	isLamda   bool // 是否是匿名函数
	isModule  bool // 是否是模块
	isInside  bool // 是否是模块内部函数
}
