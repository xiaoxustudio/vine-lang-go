package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"vine-lang/libs"
	"vine-lang/object/store"
	"vine-lang/token"
	"vine-lang/types"
	LibsUtils "vine-lang/utils"
	"vine-lang/verror"
)

type Token = token.Token

type ExecuteCodeFunc func(filename string, code string, wk Workspace) (any, error)

var executeCode ExecuteCodeFunc

func SetExecuteCode(fn ExecuteCodeFunc) {
	executeCode = fn
}

func ExecuteCode(filename string, code string, wk Workspace) (any, error) {
	if executeCode == nil {
		return nil, errors.New("execute code handler is not set")
	}
	return executeCode(filename, code, wk)
}

type Environment struct {
	types.Scope
	parent     *Environment
	store      map[Token]any
	nameMap    map[string]Token
	consts     map[string]struct{}
	FileName   string
	MountScope types.Scope // 挂载的Scope，可能是对象什么的
	WorkSpace  Workspace
	Exports    *store.StoreObject
	isPassing  bool // 是否正在定义临时参数，将不查找父级
}

func New(workspace Workspace) *Environment {
	e := &Environment{
		parent:    nil,
		store:     make(map[Token]any),
		nameMap:   make(map[string]Token), // for faster lookup
		consts:    make(map[string]struct{}),
		WorkSpace: workspace,
		isPassing: false,
	}

	lc := store.NewStoreObject()
	lc.Define(token.Token{Type: token.IDENT, Value: "Test"}, "Test")
	e.Define(token.Token{Type: token.IDENT, Value: "GLOBAL"}, lc)
	return e
}

func (e *Environment) GetWorkSpace() Workspace {
	if !WorkSpace.IsEmpty(&e.WorkSpace) {
		return e.WorkSpace
	}
	return e.parent.WorkSpace
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

	if e.MountScope != nil {
		return e.MountScope.Get(name)
	}
	return nil, false
}

func (e *Environment) Lookup(name Token) (Environment, Token) {
	if tk, exists := e.nameMap[name.Value]; exists {
		return *e, tk
	}
	if e.parent != nil && !e.isPassing {
		return e.parent.Lookup(name)
	}

	return Environment{}, Token{}
}

func (e *Environment) Set(name Token, val any) {
	theEnv, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		if _, isConst := theEnv.consts[name.Value]; isConst {
			panic(verror.InterpreterVError{
				Position: name.ToPosition(e.FileName),
				Message:  fmt.Sprintf("constant %s cannot be reassigned", LibsUtils.TrasformPrintString(name.Value)),
			})
		}
		theEnv.store[tk] = val
	} else {
		if e.MountScope != nil {
			e.MountScope.Define(name, val)
		} else {
			panic(verror.InterpreterVError{
				Position: name.ToPosition(e.FileName),
				Message:  fmt.Sprintf("variable %s is not defined", LibsUtils.TrasformPrintString(name.Value)),
			})
		}
	}
}

func (e *Environment) Define(name Token, val any) error {
	_, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		return verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("variable %s is already declared", LibsUtils.TrasformPrintString(name.Value)),
		}
	} else {
		e.store[name] = val
		e.nameMap[name.Value] = name
	}
	return nil
}

// 定义临时参数
func (e *Environment) DefinePassing(name Token, val any) {
	e.isPassing = true
	e.Define(name, val)
	e.isPassing = false
}

func (e *Environment) DefineConst(name Token, val any) {
	_, tk := e.Lookup(name)
	if !tk.IsEmpty() {
		panic(verror.InterpreterVError{
			Position: name.ToPosition(e.FileName),
			Message:  fmt.Sprintf("variable %s is already declared", LibsUtils.TrasformPrintString(name.Value)),
		})
	} else {
		e.store[name] = val
		e.nameMap[name.Value] = name
		e.consts[name.Value] = struct{}{}
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

func (e *Environment) ImportModule(name string) (any, error) {
	tk := Token{Type: token.IDENT, Value: name}

	if existed, ok := e.Get(tk); ok {
		if _, isMod := existed.(types.LibsModule); isMod {
			return nil, verror.InterpreterVError{
				Position: Token{}.ToPosition(e.FileName),
				Message:  fmt.Sprintf("module %s is already imported", LibsUtils.TrasformPrintString(name)),
			}
		}
	}
	if v, ok := libs.LibsMap[types.LibsKeywords(name)]; ok {
		e.Define(tk, v)
		return nil, nil
	}

	wk := e.GetWorkSpace()
	fullPath := filepath.Join(wk.GetBasePath(), name)
	code, err := os.ReadFile(fullPath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, verror.InterpreterVError{
				Position: Token{}.ToPosition(e.FileName),
				Message:  fmt.Sprintf("module %s is not defined or file not found", LibsUtils.TrasformPrintString(name)),
			}
		}
		return nil, verror.InterpreterVError{
			Position: Token{}.ToPosition(e.FileName),
			Message:  fmt.Sprintf("failed to read module %s: %v", LibsUtils.TrasformPrintString(name), err),
		}
	}

	oldFileName := e.FileName
	oldBasePath := wk.GetBasePath()

	e.FileName = name
	wk.Cd(filepath.Dir(name))

	result, execErr := ExecuteCode(name, string(code), wk)

	e.FileName = oldFileName
	wk.Cd(oldBasePath)

	if execErr != nil {
		return nil, execErr
	}

	e.Define(tk, result)

	return result, nil
}
