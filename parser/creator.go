package parser

import (
	"vine-lang/ast"
	"vine-lang/lexer"
)

func CreateParser(lex *lexer.Lexer) *Parser {
	c := New(lex)

	c.RegisterStmtHandler(lexer.LET, func(p *Parser) any {
		startToken := p.advance()
		isConst := startToken.Type == lexer.CST

		idTk := p.expect(lexer.IDENT)

		id := &ast.Literal{Value: idTk}
		p.expect(lexer.ASSIGN)

		value := p.parseExpression()

		return &ast.VariableDecl{
			Name:    id,
			Value:   value,
			IsConst: isConst,
		}
	})

	return c
}
