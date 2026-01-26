package types

import (
	"reflect"
	"slices"
	"strings"
	"vine-lang/object/store"
	"vine-lang/token"
)

type Checker struct{}

type CheckerTypeEnum string

const (
	CheckerTypeAny    CheckerTypeEnum = "any"
	CheckerTypeString CheckerTypeEnum = "string"
	CheckerTypeInt    CheckerTypeEnum = "int"
	CheckerTypeFloat  CheckerTypeEnum = "float"
	CheckerTypeBool   CheckerTypeEnum = "bool"
	CheckerTypeNil    CheckerTypeEnum = "nil"
	CheckerTypeFunc   CheckerTypeEnum = "func"
	CheckerTypeObject CheckerTypeEnum = "object"
	CheckerTypeArray  CheckerTypeEnum = "array"
	CheckerTypeMap    CheckerTypeEnum = "map"
)

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) NormalizeTypeName(name string) string {
	return strings.ToLower(name)
}

func (c *Checker) IsKnownTypeName(name string) bool {
	return slices.Contains([]CheckerTypeEnum{CheckerTypeString, CheckerTypeInt, CheckerTypeFloat, CheckerTypeBool, CheckerTypeNil, CheckerTypeAny, CheckerTypeFunc, CheckerTypeObject, CheckerTypeArray, CheckerTypeMap}, CheckerTypeEnum(name))
}

func (c *Checker) ResolveTypeValue(val any, resolveIdent func(token.Token) (any, bool)) any {
	if resolveIdent == nil {
		return val
	}
	if tk, ok := val.(token.Token); ok {
		if tk.Type == token.IDENT {
			if v, exists := resolveIdent(tk); exists {
				return v
			}
		}
	}
	return val
}

func (c *Checker) RuntimeTypeName(val any, resolveIdent func(token.Token) (any, bool)) CheckerTypeEnum {
	v := c.ResolveTypeValue(val, resolveIdent)
	if v == nil {
		return CheckerTypeNil
	}
	switch v.(type) {
	case string:
		return CheckerTypeString
	case int, int32, int64:
		return CheckerTypeInt
	case float32, float64:
		return CheckerTypeFloat
	case bool:
		return CheckerTypeBool
	case *FunctionLikeValNode:
		return CheckerTypeFunc
	case *store.StoreObject:
		return CheckerTypeObject
	case LibsModule:
		return CheckerTypeObject
	}
	kind := reflect.ValueOf(v).Kind()
	switch kind {
	case reflect.Func:
		return CheckerTypeFunc
	case reflect.Map:
		return CheckerTypeMap
	case reflect.Array, reflect.Slice:
		return CheckerTypeArray
	case reflect.Struct:
		return CheckerTypeObject
	}
	return CheckerTypeAny
}

func (c *Checker) MatchType(typeName string, val any, resolveIdent func(token.Token) (any, bool)) bool {
	expected := CheckerTypeEnum(c.NormalizeTypeName(typeName))
	if expected == CheckerTypeAny {
		return true
	}
	actual := c.RuntimeTypeName(val, resolveIdent)
	return actual == expected
}
