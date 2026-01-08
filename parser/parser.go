package parser

import (
	"fmt"
	"slices"
	"vine-lang/ast"
	"vine-lang/lexer"
	"vine-lang/token"
	"vine-lang/verror"
)

type Token = token.Token

type Parser struct {
	lexer    *lexer.Lexer
	tokens   []Token
	position int
	errors   []verror.ParseVError // 收集所有错误
	ast      *ast.ProgramStmt
	handlers map[token.TokenType][]func(p *Parser) any
}

func New(lex *lexer.Lexer) *Parser {
	p := &Parser{lexer: lex, tokens: []Token{}, position: 0, errors: []verror.ParseVError{}, handlers: make(map[token.TokenType][]func(p *Parser) any)}
	// 移除不进行解析的token
	for _, tk := range lex.Tokens() {
		if slices.Contains([]token.TokenType{token.WHITESPACE, token.NEWLINE}, tk.Type) {
			continue
		}
		p.tokens = append(p.tokens, tk)
	}
	return p
}

func (p *Parser) RegisterStmtHandler(kw token.TokenType, fn func(p *Parser) any) {
	if p.handlers[kw] == nil {
		p.handlers[kw] = make([]func(p *Parser) any, 0)
	}
	p.handlers[kw] = append(p.handlers[kw], fn)
}

func (p *Parser) GetErrors() []verror.ParseVError {
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
	err := verror.ParseVError{
		Position: verror.Position{
			Line:   token.Line,
			Column: token.Column},
		Message: msg,
	}
	panic(err)
}

func (p *Parser) expect(types ...token.TokenType) Token {
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
func (p *Parser) createLiteral(val token.Token) *ast.Literal {
	return &ast.Literal{Value: val}
}

/* Parsers */
func (p *Parser) ParseProgram() *ast.ProgramStmt {
	program := &ast.ProgramStmt{}
	program.Body = []ast.Stmt{}

	for !p.isEof() {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Body = append(program.Body, stmt)
		}
	}
	p.ast = program
	return program
}

func (p *Parser) parseStatement() ast.Stmt {
	tk := p.peek()

	if handlers, ok := p.handlers[tk.Type]; ok {
		for _, handler := range handlers {
			res := handler(p)
			if stmt, ok := res.(ast.Stmt); ok {
				return stmt
			} else if _, ok := res.(error); ok {
				break
			}
		}
	}

	return p.parseExpressionStatement()
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStmt {
	expr := p.parseExpression()
	return &ast.ExpressionStmt{Expression: expr}
}

/* Expr */
func (p *Parser) parseExpression() ast.Expr {
	return p.parseAssignmentExpression()
}

func (p *Parser) parseAssignmentExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	left := p.parseLogicalExpression()
	if p.peek().Type == token.ASSIGN {
		op := p.expect(token.ASSIGN)
		right := p.parseAssignmentExpression()
		return &ast.AssignmentExpr{Left: left, Right: right, Operator: op}
	}
	return left
}

func (p *Parser) parseLogicalExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	left := p.parseBinaryExpression()
	if p.peek().Type == token.OR || p.peek().Type == token.AND {
		op := p.advance()
		right := p.parseLogicalExpression()
		return &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseBinaryExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	opMap := []token.TokenType{token.PLUS, token.MINUS, token.DIV, token.MUL}
	left := p.parseCallExpression()
	if slices.Contains(opMap, p.peek().Type) {
		op := p.advance()
		right := p.parseBinaryExpression()
		return &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseArgs() *ast.ArgsExpr {
	if p.isEof() {
		return nil
	}
	var node = &ast.ArgsExpr{Arguments: []ast.Expr{}}
	for !p.isEof() && p.peek().Type != token.RPAREN {
		expr := p.parseExpression()
		if expr == nil {
			break
		}
		if p.peek().Type == token.COMMA {
			p.advance()
		}
		node.Arguments = append(node.Arguments, expr)
	}
	return node
}

func (p *Parser) parseCallExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	left := p.parsePrimaryExpression()
	if p.peek().Type == token.LPAREN {
		p.advance()
		args := p.parseArgs()
		p.expect(token.RPAREN)
		return &ast.CallExpr{Callee: left.(*ast.Literal), Args: *args}
	}
	return left
}

func (p *Parser) parsePrimaryExpression() ast.Expr {
	tk := p.peek()

	switch tk.Type {
	case token.IDENT, token.STRING, token.NUMBER:
		p.advance()
		return p.createLiteral(tk)
	case token.LPAREN:
		p.advance()
		expr := p.parseExpression()
		p.expect(token.RPAREN)
		return expr
	case token.NEWLINE, token.WHITESPACE:
		p.advance()
		return p.parsePrimaryExpression()
	default:
		p.errorf(tk, "unexpected token: %s", tk.String())
		return nil
	}
}
