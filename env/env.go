package env

import (
	"errors"
	"fmt"
	"reflect"
	"vine-lang/libs"
	"vine-lang/object/store"
	"vine-lang/token"
	"vine-lang/types"
	LibsUtils "vine-lang/utils"
	"vine-lang/verror"
)

type Token = token.Token

type Environment struct {
	types.Scope
	parent   *Environment
	store    map[Token]any
	nameMap  map[string]Token
	FileName string
}

func New(fileName string) *Environment {
	e := &Environment{
		parent:   nil,
		store:    make(map[Token]any),
		nameMap:  make(map[string]Token), // for faster lookup
		FileName: fileName,
	}

	lc := store.NewStoreObject()
	lc.Define(token.Token{Type: token.IDENT, Value: "Test"}, "Test")
	e.Define(token.Token{Type: token.IDENT, Value: "GLOBAL"}, lc)
	return e
}

func (e *Environment) Link(parent *Environment) {
	e.parent = parent
}

func (e *Environment) Get(name Token) (any, bool) {
	if tk, exists := e.nameMap[name.Value]; exists {
		return e.store[tk], true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Environment) Lookup(name Token) (Environment, Token) {
	if tk, exists := e.nameMap[name.Value]; exists {
		return *e, tk
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
		panic(verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("variable %s is not defined", LibsUtils.TrasformPrintString(name.Value)),
		})
	}
}

func (e *Environment) Define(name Token, val any) {
	_, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		panic(verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("variable %s is already declared", LibsUtils.TrasformPrintString(name.Value)),
		})
	} else {
		e.store[name] = val
		e.nameMap[name.Value] = name
	}
}

func (e *Environment) Delete(name Token) {
	delete(e.store, name)
	delete(e.nameMap, name.Value)
}

func (e *Environment) Print() {
	for k, v := range e.store {
		println(k.String(), LibsUtils.TrasformPrintString(v))
	}
}

/* 根据Token执行方法 */
func (e *Environment) CallFunc(name Token, args []any) (any, error) {
	funcVal, ok := e.Get(name)
	if ok && funcVal != nil {
		fnValue := reflect.ValueOf(funcVal)

		if fnValue.Kind() != reflect.Func {
			return nil, verror.InterpreterVError{
				Position: name.ToPosition(e.FileName),
				Message:  fmt.Sprintf("variable %s is not a function", LibsUtils.TrasformPrintString(name.Value)),
			}
		}

		reflectArgs := []reflect.Value{
			reflect.ValueOf(e),
		}

		for _, arg := range args {
			// reflect cannot handle nil
			if arg == nil {
				reflectArgs = append(reflectArgs, reflect.ValueOf(Token{
					Type:  token.NIL,
					Value: "nil",
				}))
			} else {
				reflectArgs = append(reflectArgs, reflect.ValueOf(arg))
			}
		}

		results := fnValue.Call(reflectArgs)
		if len(results) == 1 {
			return results[0].Interface(), nil
		}

		return nil, nil
	} else {
		return nil, verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("function %s is not defined", LibsUtils.TrasformPrintString(name.Value)),
		}
	}
}

/* 根据传入的方法object执行 */
func (e *Environment) CallFuncObject(fnObject any, args []any) (any, error) {
	fnObj := reflect.ValueOf(fnObject)
	if fnObj.Kind() == reflect.Func {
		reflectArgs := []reflect.Value{
			reflect.ValueOf(e),
		}

		for _, arg := range args {
			// reflect cannot handle nil
			if arg == nil {
				reflectArgs = append(reflectArgs, reflect.ValueOf(Token{
					Type:  token.NIL,
					Value: "nil",
				}))
			} else {
				reflectArgs = append(reflectArgs, reflect.ValueOf(arg))
			}
		}

		results := fnObj.Call(reflectArgs)
		if len(results) == 1 {
			return results[0].Interface(), nil
		}

		return nil, nil
	} else {
		return nil, errors.New("Not a function to Call")
	}
}

func (e *Environment) ImportModule(name string) {
	if v, ok := libs.LibsMap[types.LibsKeywords(name)]; ok {
		v.ForEach(func(tk token.Token, val any) {
			e.Define(tk, val)
		})
		e.Define(Token{Type: token.IDENT, Value: name}, v)
	} else {
		panic(verror.InterpreterVError{
			Position: Token{}.ToPosition(e.FileName),
			Message:  fmt.Sprintf("module %s is not defined", LibsUtils.TrasformPrintString(name)),
		})
	}
}
