package main

import (
	"os"
	"vine-lang/env"
	"vine-lang/ipt"
	"vine-lang/lexer"
	"vine-lang/parser"
)

func main() {
	bytes, _ := os.ReadFile("./examples/001.vine")
	var lex = lexer.New("main.vine", string(bytes))
	lex.Parse()
	var p = parser.New(lex)
	var e = env.New()
	var i = ipt.New(p, e)
	i.EvalSafe()
	e.Print()
}
