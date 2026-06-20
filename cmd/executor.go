package cmd

import (
	"fmt"
	"os"
	"vine-lang/compiler"
	"vine-lang/env"
	"vine-lang/lexer"
	"vine-lang/parser"
	"vine-lang/types"
	"vine-lang/verror"
	"vine-lang/vm"
)

func init() {
	env.SetExecuteCode(executeCode)
}

func executeVineFile(filepath string, wk env.Workspace) error {
	defer func() {
		if r := recover(); r != nil {
			handleError(r)
		}
	}()

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("无法读取文件 %s: %v", filepath, err)
	}

	_, err = executeCode(filepath, string(bytes), wk)

	return err
}

func executeCode(filename string, code string, wk env.Workspace) (any, error) {
	lex := lexer.New(filename, code)
	lex.Parse()

	p := parser.CreateParser(lex)

	e := env.New(wk)
	e.FileName = filename

	c := compiler.NewCompiler(e)
	ast := p.ParseProgram()
	_, err := c.Compile(ast)
	if err != nil {
		return nil, err
	}

	v := vm.NewVM(c, e)
	result, err := v.Run()
	if err != nil {
		return nil, err
	}

	if e.Exports != nil {
		return types.NewUserModule(filename, e.Exports), nil
	}

	return result, nil
}

func handleError(r any) {
	switch err := r.(type) {
	case verror.VError:
		fmt.Fprintln(os.Stderr, err.Error())
	case verror.ParseVError:
		fmt.Fprintln(os.Stderr, err.Error())
	case verror.InterpreterVError:
		fmt.Fprintln(os.Stderr, err.Error())
	case verror.LexerVError:
		fmt.Fprintln(os.Stderr, err.Error())
	default:
		fmt.Fprintln(os.Stderr, r)
	}
}
