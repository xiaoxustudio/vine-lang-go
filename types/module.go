package types

import "vine-lang/token"

type LibsModule interface {
	ID() LibsKeywords
	Name() string
	IsInside() bool
	Return() StoreObject
	Register(fnName string, fn any)
	Get(key token.Token) (any, bool)
	Set(key token.Token, value any) error
	ForEach(func(tk token.Token, val any))
}

type LibsModuleObject struct {
	LibsModule
	Path  func() string
	Store StoreObject
}

func (l *LibsModuleObject) ID() LibsKeywords {
	return Unknown
}

func (l *LibsModuleObject) Get(key token.Token) (any, bool) {
	return l.Store.Get(key)
}

func (l *LibsModuleObject) Set(key token.Token, value any) error {
	return l.Store.Define(key, value)
}

func (l *LibsModuleObject) Return() StoreObject {
	return l.Store
}

func (l *LibsModuleObject) IsInside() bool {
	return true
}

func (l *LibsModuleObject) Register(fnName string, fn any) {
	l.Set(token.Token{Type: token.IDENT, Value: fnName}, fn)
}

func (l *LibsModuleObject) ForEach(fn func(tk token.Token, val any)) {
	l.Store.ForEach(fn)
}
