package main

import (
	"os"
	"vine-lang/lexer"
)

func main() {
	bytes, _ := os.ReadFile("./examples/001.vine")
	var lex = lexer.New("main.vine", string(bytes))
	lex.Parse()
	lex.Print()
	// var p = parser.New(lex)
	// p.ParseProgram()
	// fmt.Println(p)
}
