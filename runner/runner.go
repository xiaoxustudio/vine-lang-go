package runner

import (
	"vine-lang/compiler"
	"vine-lang/env"
	"vine-lang/lexer"
	"vine-lang/parser"
	"vine-lang/types"
	"vine-lang/vm"
)

func ExecuteCode(filename string, code string, wk env.Workspace) (any, error) {
	lex := lexer.New(filename, code)
	if err := lex.Parse(); err != nil {
		return nil, err
	}

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
