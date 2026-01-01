package parser

import (
	"vine-lang/ast"
	"vine-lang/lexer"
)

type Parser struct {
	lexer        *lexer.Lexer
	currentToken lexer.Token
	peekToken    lexer.Token
}

func New(lexer *lexer.Lexer) *Parser {
	p := &Parser{lexer: lexer}
	// Read two tokens, so currentToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	res, err := p.lexer.GetToken()
	if err != nil {
		panic(err)
	}
	p.currentToken = p.peekToken
	p.peekToken = res
}

func (p *Parser) ParseProgram() *ast.ProgramStmt {
	program := &ast.ProgramStmt{}
	program.Body = []ast.Stmt{}
	for p.currentToken.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Body = append(program.Body, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Stmt {
	switch p.currentToken.Type {
	case lexer.LET:
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStmt {
	return &ast.ExpressionStmt{Expression: p.parseExpression()}
}

func (p *Parser) parseExpression() ast.Expr {
	return p.parsePrimaryExpression()
}

func (p *Parser) parsePrimaryExpression() ast.Expr {
	tk := p.currentToken
	switch tk.Type {
	case lexer.IDENT:
	case lexer.STRING:
	case lexer.NUMBER:
		return &ast.Literal{Value: tk}
	default:
		panic("unknown token type")
	}
	return nil
}
