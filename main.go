package main

import (
	"os"
	"vine-lang/lexer"
	"vine-lang/parser"
)

func main() {
	bytes, _ := os.ReadFile("./examples/001.vine")
	var lex = lexer.New("main.vine", string(bytes))
	lex.Parse()
	var p = parser.New(lex)
	p.ParseProgram().Print()
}
