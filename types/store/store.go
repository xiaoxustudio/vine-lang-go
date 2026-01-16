package store

import (
	"fmt"
	"vine-lang/token"
	"vine-lang/types"
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

func (e *StoreObject) Define(name token.Token, val any) {
	_, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		panic(verror.InterpreterVError{
			Position: name.ToPosition(""),
			Message:  fmt.Sprintf("variable %s is already declared", LibsUtils.TrasformPrintString(name.Value)),
		})
	} else {
		e.store[name] = val
		e.nameMap[name.Value] = name
	}
}

func (e *StoreObject) Print() {
	for k, v := range e.store {
		println(k.String(), LibsUtils.TrasformPrintString(v))
	}
}
