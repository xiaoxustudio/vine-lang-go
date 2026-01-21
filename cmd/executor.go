package cmd

import (
	"fmt"
	"os"
	"vine-lang/env"
	"vine-lang/ipt"
	"vine-lang/lexer"
	"vine-lang/parser"
	"vine-lang/verror"
)

func executeVineFile(filepath string) error {
	defer func() {
		if r := recover(); r != nil {
			handleError(r)
		}
	}()

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("无法读取文件 %s: %v", filepath, err)
	}

	return executeCode(filepath, string(bytes))
}

func executeCode(filename string, code string) error {
	lex := lexer.New(filename, code)
	lex.Parse()

	p := parser.CreateParser(lex)

	e := env.New(filename)

	i := ipt.New(p, e)
	_, err := i.EvalSafe()

	return err
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
