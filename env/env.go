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

func (e *Environment) Lookup(name Token) (Environment, Token) {
	for k := range e.store {
		if k.Value == name.Value {
			return *e, k
		}
	}
	if e.parent != nil {
		return e.parent.Lookup(name)
	}
	return Environment{}, Token{}
}

func (e *Environment) Set(name Token, val any) {
	theEnv, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		theEnv.store[tk] = val
	} else {
		panic(&verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("variable %s is not defined", global.TrasformPrintString(name.Value)),
		})
	}
}

func (e *Environment) Define(name Token, val any) {
	_, tk := e.Lookup(name)
	if !tk.IsEmpty() {
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
