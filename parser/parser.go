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
		if slices.Contains([]token.TokenType{token.WHITESPACE}, tk.Type) {
			continue
		}
		p.tokens = append(p.tokens, tk)
	}
	return p
}

func (p *Parser) Print() {
	for _, v := range p.ast.Body {
		fmt.Println(v.String())
	}
}

func (p *Parser) RegisterStmtHandler(kw token.TokenType, fn func(p *Parser) any) {
	if p.handlers[kw] == nil {
		p.handlers[kw] = make([]func(p *Parser) any, 0)
	}
	p.handlers[kw] = append(p.handlers[kw], fn)
}

func (p *Parser) RegisterStmtHandlerWithKeyWords(kw []token.TokenType, fns func(p *Parser) any) {
	for _, v := range kw {
		p.RegisterStmtHandler(v, fns)
	}
}

func (p *Parser) CallStmtHandler(tk token.TokenType) ast.Stmt {
	if handlers, ok := p.handlers[tk]; ok {
		for _, handler := range handlers {
			return handler(p).(ast.Stmt)
		}
	}
	return nil
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

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peek().Type == t
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
		} else {
			i++
		}
	}

	current := p.peek()
	typeStr := fmt.Sprintf("%v", types)
	p.errorf(current, "expected next token to be %s, got %s instead", typeStr, current.Type)
	return Token{}
}

/* Creaters  */
func (p *Parser) createLiteral(val token.Token) *ast.Literal {
	return &ast.Literal{Value: &val}
}

/* Parsers */
func (p *Parser) ParseProgram() *ast.ProgramStmt {
	p.ast = &ast.ProgramStmt{}
	p.ast.Body = []ast.Stmt{}

	for !p.isEof() {
		if p.peek().Type == token.EOF {
			break
		}
		if stmt := p.parseStatement(); stmt != nil {
			p.ast.Body = append(p.ast.Body, stmt)
		}
	}

	return p.ast
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

	if p.peekTokenIs(token.NEWLINE) || p.peekTokenIs(token.SEMICOLON) {
		p.advance()
		return nil
	}

	return p.parseExpressionStatement()
}

func (p *Parser) parseBlockStatement() *ast.BlockStmt {
	p.expect(token.COLON)
	var body []ast.Stmt
	for !p.isEof() && p.peek().Type != token.END {
		stmt := p.parseStatement()
		if stmt != nil {
			body = append(body, stmt)
		}
	}
	p.expect(token.END)
	return &ast.BlockStmt{Body: body}
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
	left := p.parseCompareExpression()
	if p.peek().Type == token.OR || p.peek().Type == token.AND {
		op := p.advance()
		right := p.parseLogicalExpression()
		return &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseCompareExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	left := p.parseBinaryExpression()
	if p.peek().Type == token.EQ || p.peek().Type == token.NOT_EQ || p.peek().Type == token.LESS_EQ || p.peek().Type == token.GREATER_EQ || p.peek().Type == token.LESS || p.peek().Type == token.GREATER {
		op := p.advance()
		right := p.parseCompareExpression()
		return &ast.CompareExpr{Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseBinaryExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	return p.parseAdditiveExpression()
}

func (p *Parser) parseAdditiveExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	left := p.parseMultiplicativeExpression()
	for p.peek().Type == token.PLUS || p.peek().Type == token.MINUS {
		op := p.advance()
		right := p.parseMultiplicativeExpression()
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseMultiplicativeExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	left := p.parseCallExpression()
	for p.peek().Type == token.MUL || p.peek().Type == token.DIV {
		op := p.advance()
		right := p.parseCallExpression()
		left = &ast.BinaryExpr{Left: left, Operator: op, Right: right}
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
	left := p.parseMemberExpression()
	for p.peek().Type == token.LPAREN {
		p.advance()
		args := p.parseArgs()
		p.expect(token.RPAREN)
		left = &ast.CallExpr{Callee: left, Args: *args}

		// 可能是换行
		if p.peek().Type == token.NEWLINE {
			p.advance()
		}

		var parentToStmt = &ast.ToExpr{Next: nil}
		var currentToStmt = parentToStmt
		if p.peek().Type == token.TO {
			for p.peek().Type == token.TO || p.peek().Type != token.CATCH {
				p.advance()
				toStmt := &ast.ToExpr{}
				if p.peek().Type == token.LPAREN {
					p.advance()
					args := p.parseArgs()
					p.expect(token.RPAREN)
					toStmt.Args = *args
				}
				p.expect(token.COLON)
				var block = &ast.BlockStmt{Body: []ast.Stmt{}}
				for !slices.Contains([]token.TokenType{token.TO, token.END, token.CATCH}, p.peek().Type) && !p.isEof() {
					stmt := p.parseStatement()
					if stmt != nil {
						block.Body = append(block.Body, stmt)
					}
				}
				toStmt.Body = *block
				currentToStmt.Next = toStmt
				if p.peek().Type == token.TO {
					currentToStmt = toStmt
				}
			}
			var catchStmt = &ast.LambdaFunctionDecl{Body: ast.BlockStmt{}}
			if p.peek().Type == token.CATCH {
				p.advance()
				p.expect(token.LPAREN)
				args := p.parseArgs()
				p.expect(token.RPAREN)
				p.expect(token.COLON)
				var blockStmt = &ast.BlockStmt{Body: []ast.Stmt{}}
				for !p.isEof() && p.peek().Type != token.END {
					stmt := p.parseStatement()
					if stmt != nil {
						blockStmt.Body = append(blockStmt.Body, stmt)
					}
				}
				catchStmt = &ast.LambdaFunctionDecl{Args: *args, Body: *blockStmt}
			}

			p.expect(token.END)

			target, ok := left.(*ast.CallExpr)
			if !ok {
				return left
			}

			return &ast.CallTaskFn{
				Target: *target,
				To:     *parentToStmt.Next,
				Catch:  catchStmt,
			}
		}
	}

	return left
}

func (p *Parser) parseLambda() *ast.LambdaFunctionDecl {
	if p.isEof() {
		return nil
	}
	p.expect(token.FN)
	var args = &ast.ArgsExpr{Arguments: []ast.Expr{}}
	if p.peek().Type == token.LPAREN {
		p.expect(token.LPAREN)
		args = p.parseArgs()
		p.expect(token.RPAREN)
	}
	body := p.parseBlockStatement()
	return &ast.LambdaFunctionDecl{Args: *args, Body: *body}
}

func (p *Parser) parsePropertyExpression() []*ast.Property {
	var properties = []*ast.Property{}
	if p.isEof() {
		return properties
	}
	var index = 0
	for p.peek().Type != token.RBRACE && p.peek().Type != token.RBRACKET {
		key := p.parseExpression()
		if p.peek().Type == token.COLON {
			p.advance()
			value := p.parseExpression()
			if p.peek().Type == token.COMMA {
				p.advance()
			}
			properties = append(properties, &ast.Property{Key: key.(*ast.Literal), Value: value})
		} else {
			if p.peek().Type == token.COMMA {
				p.advance()
			}
			properties = append(properties, &ast.Property{Key: p.createLiteral(token.Token{Type: token.INT, Value: fmt.Sprint(index)}), Value: key})
		}
		index++
	}
	return properties
}

func (p *Parser) parseArrayExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	args := p.parsePropertyExpression()
	arr := &ast.ArrayExpr{Items: args}
	return arr
}

func (p *Parser) parseObjectExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	args := p.parsePropertyExpression()
	obj := &ast.ObjectExpr{Properties: args}
	return obj
}

func (p *Parser) parseMemberExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	left := p.parseSuffixExpression()
	if p.peek().Type == token.DOT {
		p.advance()
		right := p.parseMemberExpression()
		return &ast.MemberExpr{Object: left, Property: right}
	} else if p.peek().Type == token.LBRACKET {
		p.advance()
		right := p.parseMemberExpression()
		p.expect(token.RBRACKET)
		return &ast.MemberExpr{Object: left, Property: right, Computed: true}
	}
	return left
}

func (p *Parser) parseSuffixExpression() ast.Expr {
	if p.isEof() {
		return nil
	}
	left := p.parsePrimaryExpression()
	if p.peek().Type == token.INC || p.peek().Type == token.DEC {
		op := p.advance()
		return &ast.UnaryExpr{Operator: op, Value: left, IsSuffix: true}
	}
	return left
}

func (p *Parser) parseSwitchCase() ast.Expr {
	if p.isEof() {
		return nil
	}
	var kw Token
	if p.peek().Type == token.CASE || p.peek().Type == token.DEFAULT {
		kw = p.advance()
	} else {
		p.errorf(p.peek(), "unexpected token(switch case): %s", p.peek().String())
	}
	var isDefault = kw.Type == token.DEFAULT
	var node = &ast.SwitchCase{Conds: nil, Body: &ast.BlockStmt{}, IsDefault: isDefault}
	var cond ast.Expr
	if !isDefault {
		for p.peek().Type != token.COLON {
			if p.peek().Type == token.COMMA {
				p.advance()
			}
			cond = p.parseExpression()
			node.Conds = append(node.Conds, cond)
		}
	}
	p.expect(token.COLON)

	/* 解析body */
	var body []ast.Stmt
	for !p.isEof() && !slices.Contains([]token.TokenType{token.DEFAULT, token.BREAK, token.CASE}, p.peek().Type) {
		stmt := p.parseStatement()
		if stmt != nil {
			body = append(body, stmt)
		}
	}
	if p.peek().Type == token.BREAK {
		p.advance()
	}
	node.Body = &ast.BlockStmt{Body: body}
	return node
}

func (p *Parser) parsePrimaryExpression() ast.Expr {
	tk := p.peek()

	switch tk.Type {
	case token.IDENT, token.STRING, token.INT, token.FLOAT, token.NIL, token.TRUE, token.FALSE:
		p.advance()
		return p.createLiteral(tk)
	case token.LPAREN:
		p.advance()
		expr := p.parseExpression()
		p.expect(token.RPAREN)
		return expr
	case token.LBRACKET:
		p.advance()
		expr := p.parseArrayExpression()
		p.expect(token.RBRACKET)
		return expr
	case token.LBRACE:
		p.advance()
		expr := p.parseObjectExpression()
		p.expect(token.RBRACE)
		return expr
	case token.NEWLINE, token.WHITESPACE, token.COMMENT:
		p.advance()
		return p.parsePrimaryExpression()
	case token.NOT, token.MINUS, token.DEC, token.INC:
		op := p.advance()
		right := p.parseSuffixExpression()
		return &ast.UnaryExpr{Operator: op, Value: right, IsSuffix: false}
	case token.WAIT:
		return p.CallStmtHandler(token.WAIT)
	case token.FN:
		return p.parseLambda()
	default:
		p.errorf(tk, "primary unexpected token: %s", tk.String())
		return nil
	}
}
