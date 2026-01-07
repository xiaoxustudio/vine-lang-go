package env

import (
	"fmt"
	"reflect"
	"vine-lang/lexer"
	"vine-lang/libs"
	"vine-lang/token"
	"vine-lang/types"
	LibsUtils "vine-lang/utils"
	"vine-lang/verror"
)

type Token = lexer.Token

type Environment struct {
	parent   *Environment
	store    map[Token]any
	FileName string
}

func New(fileName string) *Environment {
	e := &Environment{parent: nil, store: make(map[Token]any), FileName: fileName}
	return e
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
			Message:  fmt.Sprintf("variable %s is not defined", LibsUtils.TrasformPrintString(name.Value)),
		})
	}
}

func (e *Environment) Define(name Token, val any) {
	_, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		panic(&verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("variable %s is already declared", LibsUtils.TrasformPrintString(name.Value)),
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
		println(k.String(), LibsUtils.TrasformPrintString(v))
	}
}

/* Function */
func (e *Environment) CallFunc(name Token, args []any) (any, error) {
	funcVal, ok := e.Get(name)
	if ok && funcVal != nil {
		fnValue := reflect.ValueOf(funcVal)

		if fnValue.Kind() != reflect.Func {
			return nil, &verror.InterpreterVError{
				Position: name.ToPosition(e.FileName),
				Message:  fmt.Sprintf("variable %s is not a function", LibsUtils.TrasformPrintString(name.Value)),
			}
		}

		args := []reflect.Value{
			reflect.ValueOf(e),
			reflect.ValueOf(args),
		}

		results := fnValue.Call(args)

		return results, nil
	} else {
		return nil, &verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("function %s is not defined", LibsUtils.TrasformPrintString(name.Value)),
		}
	}
}

func (e *Environment) ImportModule(name string) {
	if v, ok := libs.LibsMap[types.LibsKeywords(name)]; ok {
		for k, v := range v.Return() {
			e.Define(Token{Type: token.IDENT, Value: k, Line: 0, Column: 0}, v)
		}
	} else {
		panic(&verror.InterpreterVError{
			Position: Token{}.ToPosition(e.FileName),
			Message:  fmt.Sprintf("module %s is not defined", LibsUtils.TrasformPrintString(name)),
		})
	}
}
