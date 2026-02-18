package store

import (
	"encoding/json"
	"fmt"
	"go/types"
	"strconv"
	"vine-lang/token"
	LibsUtils "vine-lang/utils"
	"vine-lang/verror"
)

type StoreObject struct {
	types.Scope
	parent  *StoreObject
	store   map[token.Token]any
	nameMap map[string]token.Token
}

func NewStoreObject() *StoreObject {
	return &StoreObject{
		store:   make(map[token.Token]any),
		nameMap: make(map[string]token.Token),
		parent:  nil,
	}
}

func StoreObjectToMap(e *StoreObject) map[string]any {
	res := make(map[string]any)
	for k, v := range e.store {
		res[k.Value] = v
	}
	return res
}

func StoreObjectToReadableJSON(e *StoreObject) string {
	data := toJSONValue(e)
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprint(data)
	}
	return string(raw)
}

func toJSONValue(val any) any {
	switch v := val.(type) {
	case nil:
		return nil
	case *StoreObject:
		return storeObjectToJSONMap(v)
	case map[string]any:
		res := make(map[string]any, len(v))
		for k, item := range v {
			if k == "__proto__" {
				continue
			}
			res[k] = toJSONValue(item)
		}
		return res
	case []any:
		res := make([]any, len(v))
		for i, item := range v {
			res[i] = toJSONValue(item)
		}
		return res
	case token.Token:
		switch v.Type {
		case token.INT:
			if i, err := v.GetInt(); err == nil {
				return i
			}
		case token.FLOAT:
			if f, err := v.GetFloat(); err == nil {
				return f
			}
		case token.TRUE, token.FALSE:
			if b, err := strconv.ParseBool(v.Value); err == nil {
				return b
			}
		case token.NIL:
			return nil
		default:
			return v.Value
		}
		return v.Value
	case bool, int, int64, int32, float32, float64, string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

func storeObjectToJSONMap(e *StoreObject) map[string]any {
	res := make(map[string]any)
	for k, v := range e.store {
		if k.Value == "__proto__" {
			continue
		}
		res[k.Value] = toJSONValue(v)
	}
	return res
}

func (s *StoreObject) Get(name token.Token) (any, bool) {
	if tk, exists := s.nameMap[name.Value]; exists {
		return s.store[tk], true
	}
	if s.parent != nil {
		return s.parent.Get(name)
	}
	return nil, false
}

func (e *StoreObject) Lookup(name token.Token) (StoreObject, token.Token) {
	if tk, exists := e.nameMap[name.Value]; exists {
		return *e, tk
	}
	if e.parent != nil {
		return e.parent.Lookup(name)
	}
	return StoreObject{}, token.Token{}
}

func (e *StoreObject) Set(name token.Token, val any) {
	theEnv, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		theEnv.store[tk] = val
	} else {
		panic(verror.InterpreterVError{
			Position: name.ToPosition(""),
			Message:  fmt.Sprintf("variable %s is not defined", LibsUtils.TrasformPrintString(name.Value)),
		})
	}
}

func (e *StoreObject) Define(name token.Token, val any) error {
	_, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		return verror.InterpreterVError{
			Position: name.ToPosition(""),
			Message:  fmt.Sprintf("variable %s is already declared", LibsUtils.TrasformPrintString(name.Value)),
		}
	} else {
		e.store[name] = val
		e.nameMap[name.Value] = name
	}
	return nil
}

func (e *StoreObject) Print() {
	for k, v := range e.store {
		println(k.String(), LibsUtils.TrasformPrintString(v))
	}
}

func (e *StoreObject) ForEach(fn func(tk token.Token, val any)) {
	for k, v := range e.store {
		fn(k, v)
	}
}

func (e *StoreObject) IsEmpty() bool {
	return len(e.store) == 0
}
