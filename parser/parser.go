package parser

import (
	"fmt"
	"slices"
	"vine-lang/ast"
	"vine-lang/lexer"
)

type Token = lexer.Token
type ParseError struct {
	Line    int
	Column  int
	Message string
}
type Parser struct {
	lexer    *lexer.Lexer
	tokens   []Token
	position int
	errors   []ParseError // 收集所有错误
}

func (e ParseError) Error() string {
	return fmt.Sprintf("[Line %d, Column %d] Error: %s", e.Line, e.Column, e.Message)
}

func New(lex *lexer.Lexer) *Parser {
	p := &Parser{lexer: lex, tokens: []Token{}, position: 0, errors: []ParseError{}}
	// 移除不进行解析的token
	for _, token := range lex.Tokens() {
		if slices.Contains([]lexer.TokenType{lexer.WHITESPACE, lexer.NEWLINE}, token.Type) {
			continue
		}
		p.tokens = append(p.tokens, token)
	}
	return p
}

func (p *Parser) GetErrors() []ParseError {
	return p.errors
}

/* Tool */
func (p *Parser) peek() Token {
	if p.isEof() {
		return p.lexer.TheEof()
	}
	return p.tokens[p.position]
}

func (p *Parser) peekIndex(index int) Token {
	if p.isEof() {
		return p.lexer.TheEof()
	}
	return p.tokens[p.position+index]
}

func (p *Parser) advance() Token {
	current := p.peek()
	if !p.isEof() {
		p.position++
	}
	return current
}

func (p *Parser) advanceIndex(index int) Token {
	current := p.peekIndex(index)
	if !p.isEof() {
		p.position += index + 1
	}
	return current
}

func (p *Parser) isEof() bool {
	return p.position >= len(p.tokens)
}

func (p *Parser) errorf(token Token, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	err := ParseError{
		Line:    token.Line,
		Column:  token.Col,
		Message: msg,
	}
	panic(err)
}

func (p *Parser) expect(types ...lexer.TokenType) Token {
	if len(types) == 0 {
		current := p.peek()
		p.errorf(current, "Internal error: expect() called with no arguments")
	}

	var i = 0
	for !p.isEof() {
		current := p.peekIndex(i)
		if slices.Contains(types, current.Type) {
			return p.advanceIndex(i)
		}
		i++
	}

	current := p.peek()
	typeStr := fmt.Sprintf("%v", types)
	p.errorf(current, "expected next token to be %s, got %s instead", typeStr, current.Type)
	return Token{}
}

/* Creaters  */
func (p *Parser) createLiteral(val lexer.Token) *ast.Literal {
	return &ast.Literal{Value: val}
}

/* Parsers */
func (p *Parser) ParseProgram() *ast.ProgramStmt {
	program := &ast.ProgramStmt{}
	program.Body = []ast.Stmt{}

	for !p.isEof() {
		stmt := p.parseStatementSafe()
		if stmt != nil {
			program.Body = append(program.Body, stmt)
		}
	}

	return program
}

func (p *Parser) parseStatementSafe() ast.Stmt {
	defer func() {
		if r := recover(); r != nil {
			if parseErr, ok := r.(ParseError); ok {
				p.errors = append(p.errors, parseErr)
				fmt.Printf("Caught Error: %v\n", parseErr)
			} else {
				panic(r)
			}
			p.synchronize()
		}
	}()

	return p.parseStatement()
}

func (p *Parser) synchronize() {
	for !p.isEof() {
		tk := p.peek()
		if tk.Type == lexer.SEMICOLON || tk.Type == lexer.NEWLINE {
			p.advance()
			return
		}
		p.advance()
	}
}

func (p *Parser) parseStatement() ast.Stmt {
	tk := p.peek()
	switch tk.Type {
	case lexer.LET:
		return p.parseLetStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.VariableDecl {
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
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStmt {
	return &ast.ExpressionStmt{Expression: p.parseExpression()}
}

func (p *Parser) parseExpression() ast.Expr {
	return p.parseBinaryExpression()
}

func (p *Parser) parseBinaryExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	opMap := []lexer.TokenType{lexer.PLUS, lexer.MINUS, lexer.ASTERISK, lexer.SLASH}
	left := p.parsePrimaryExpression()
	if slices.Contains(opMap, p.peek().Type) {
		op := p.advance()
		right := p.parseBinaryExpression()
		return &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parsePrimaryExpression() ast.Expr {
	tk := p.peek()

	switch tk.Type {
	case lexer.IDENT, lexer.STRING, lexer.NUMBER:
		p.advance()
		return p.createLiteral(tk)
	case lexer.NEWLINE, lexer.WHITESPACE:
		p.advance()
		return p.parsePrimaryExpression()
	default:
		p.errorf(tk, "unexpected token: %s", tk.Value)
		return nil
	}
}
