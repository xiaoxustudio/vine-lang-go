package types

import (
	"errors"
	"vine-lang/token"
)

type Scope interface {
	Get(t token.Token) (any, bool)
	Set(t token.Token, val any)
	Print()
	ForEach(fn func(tk token.Token, val any))
	Define(t token.Token, val any)
}

type StoreObject struct {
	Scope
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

func (e *StoreObject) Set(name token.Token, val any) error {
	theEnv, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		theEnv.store[tk] = val
		return nil
	} else {
		return errors.New("variable not found")
	}
}

func (e *StoreObject) Define(name token.Token, val any) error {
	_, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		return errors.New("variable already defined")
	} else {
		e.store[name] = val
		e.nameMap[name.Value] = name
		return nil
	}
}

func (e *StoreObject) ForEach(fn func(tk token.Token, val any)) {
	for k, v := range e.store {
		fn(k, v)
	}
}

func (e *StoreObject) Print() {
	for k, v := range e.store {
		println(k.Value, v)
	}
}

func (e *StoreObject) IsEmpty() bool {
	return len(e.store) == 0
}
