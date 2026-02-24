package types

import (
	"reflect"
	"vine-lang/ast"
	"vine-lang/token"
)

type ValNodeType int

const (
	VNT_Unknown ValNodeType = iota
	VNT_IDENT
	VNT_INT
	VNT_FLOAT
	VNT_STRING
	VNT_BOOL
	VNT_FUNC
	VNT_OBJECT
	VNT_ARRAY
	VNT_MAP
	VNT_NIL
)

type Val interface {
	IsObject() bool
	Value() any
	Type() ValNodeType
}

// 运行值，用于存储运行时的值
type ValNode struct {
	Val
	Token *token.Token
	V     any
}

func CreateValNode(token *token.Token) *ValNode {
	return &ValNode{
		Token: token,
		V:     token.Value,
	}
}

func (v *ValNode) IsObject() bool {
	return false
}

func (v ValNode) Value() any {
	if v.V != nil {
		return v.V
	}
	return v.Token
}

func (v ValNode) Type() ValNodeType {
	switch v.Token.Type {
	case token.IDENT:
		return VNT_IDENT
	case token.INT:
		return VNT_INT
	case token.FLOAT:
		return VNT_FLOAT
	case token.STRING:
		return VNT_STRING
	case token.TRUE, token.FALSE:
		return VNT_BOOL
	case token.NIL:
		return VNT_NIL
	}

	kd := reflect.ValueOf(v.Value).Kind()
	switch kd {
	case reflect.Func:
		return VNT_FUNC
	case reflect.Map:
		return VNT_MAP
	case reflect.Array, reflect.Slice:
		return VNT_ARRAY
	case reflect.Struct:
		return VNT_OBJECT
	}
	return VNT_Unknown
}

type FunctionLikeValNode struct {
	Val
	Token    *token.Token
	Args     *ast.ArgsExpr
	Body     *ast.BlockStmt
	IsLamda  bool // 是否是匿名函数
	IsModule bool // 是否是模块
	IsInside bool // 是否是模块内部函数
	IsTask   bool // 是否是协程函数
}

// 任务
type TaskToValNode struct {
	env     any
	Current func() any
	Parnet  any
	Next    *TaskToValNode
}

func (v *TaskToValNode) Env(env any) any {
	if env == nil {
		return v.env
	}
	v.env = env
	return v.env
}
