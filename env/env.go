package env

import (
	"vine-lang/lexer"
	"vine-lang/libs/global"
)

type Token = lexer.Token

type Environment struct {
	store map[Token]any
}

func New() *Environment {
	return &Environment{store: make(map[Token]any)}
}

func (e *Environment) Get(name Token) (any, bool) {
	val, ok := e.store[name]
	return val, ok
}

func (e *Environment) Set(name Token, val any) {
	e.store[name] = val
}

func (e *Environment) Delete(name Token) {
	delete(e.store, name)
}

func (e *Environment) Print() {
	for k, v := range e.store {
		println(k.String(), global.TrasformPrintString(v))
	}
}
