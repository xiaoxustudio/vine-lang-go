package env

import (
	"fmt"
	"vine-lang/lexer"
	"vine-lang/libs/global"
	"vine-lang/verror"
)

type Token = lexer.Token

type Environment struct {
	parent   *Environment
	store    map[Token]any
	FileName string
}

func New(fileName string) *Environment {
	return &Environment{parent: nil, store: make(map[Token]any), FileName: fileName}
}

func (e *Environment) Lookup(name Token) (any, bool) {
	return e.Get(name)
}

func (e *Environment) Get(name Token) (any, bool) {
	for k := range e.store {
		if k.Value == name.Value {
			return e.store[k], true
		}
	}
	if e.parent != nil {
		return e.parent.Get(name)
	} else {
		return nil, false
	}
}

func (e *Environment) GetKey(name Token) (Token, bool) {
	for k := range e.store {
		if k.Value == name.Value {
			return k, true
		}
	}
	if e.parent != nil {
		return e.parent.GetKey(name)
	} else {
		return Token{}, false
	}
}

func (e *Environment) Set(name Token, val any) {
	if _, ok := e.GetKey(name); ok {
		e.store[name] = val
	} else {
		panic(&verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("variable %s is not defined", global.TrasformPrintString(name.Value)),
		})
	}
}

func (e *Environment) Define(name Token, val any) {
	if _, ok := e.GetKey(name); ok {
		panic(&verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("variable %s is already declared", global.TrasformPrintString(name.Value)),
		})
	} else {
		e.store[name] = val
	}
}

func (e *Environment) Delete(name Token) {
	delete(e.store, name)
}

func (e *Environment) Print() {
	for k, v := range e.store {
		println(k.String(), global.TrasformPrintString(v))
	}
}
