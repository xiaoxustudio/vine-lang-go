package main

import (
	"fmt"
	"os"
	"vine-lang/env"
	"vine-lang/ipt"
	"vine-lang/lexer"
	"vine-lang/parser"
	"vine-lang/verror"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(verror.VError); ok {
				fmt.Println(err.Error())
				return
			}
			if err, ok := r.(verror.ParseVError); ok {
				fmt.Println(err.Error())
				return
			}
			if err, ok := r.(verror.InterpreterVError); ok {
				fmt.Println(err.Error())
				return
			}
			if err, ok := r.(verror.LexerVError); ok {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(r)
		}
	}()
	bytes, _ := os.ReadFile("./examples/001.vine")
	var lex = lexer.New("main.vine", string(bytes))
	lex.Parse()
	// lex.Print()
	var p = parser.CreateParser(lex)
	var e = env.New("main.vine")
	var i = ipt.New(p, e)
	i.EvalSafe()
	// e.Print()
}
