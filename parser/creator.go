package parser

import (
	"fmt"
	"slices"
	"vine-lang/ast"
	"vine-lang/lexer"
	"vine-lang/token"
	"vine-lang/verror"
)

func CreateParser(lex *lexer.Lexer) *Parser {
	c := New(lex)

	c.RegisterStmtHandler(token.COMMENT, func(p *Parser) any {
		return &ast.CommentStmt{
			Value: p.advance(),
		}
	})

	c.RegisterStmtHandlerWithKeyWords([]token.TokenType{token.LET, token.CST}, func(p *Parser) any {
		startToken := p.advance()
		isConst := startToken.Type == token.CST

		idTk := p.expect(token.IDENT)

		id := ast.NewLiteral(&idTk)
		p.expect(token.ASSIGN)

		value := p.parseExpression()

		return ast.NewVariableDecl(*id, value, isConst)
	})

	c.RegisterStmtHandler(token.USE, func(p *Parser) any {
		p.advance() // skip 'use'
		var source *ast.Literal
		likeSource := p.parsePrimaryExpression()
		if _, e := likeSource.(*ast.Literal); !e {
			panic(fmt.Sprintf("expected literal, got %s", likeSource.String()))
		}
		source = likeSource.(*ast.Literal)
		var specifiers []ast.Specifier
		// use "fmt" , use "fmt" as fmt, use "fmt" pick addr, use "fmt" pick (add,sub)
		if p.peek().Type == token.AS {
			p.advance() // skip 'as'
			alias := p.parsePrimaryExpression()
			specifiers = append(specifiers, alias)
			return ast.NewUseDecl(source, specifiers, token.USE)
		} else if p.peek().Type == token.PICK {
			p.advance() // skip 'pick'
			if p.peek().Type == token.LPAREN {
				p.advance() // skip '('
				for p.peek().Type != token.RPAREN {
					remoteExpr := p.parsePrimaryExpression()
					if lit, ok := remoteExpr.(*ast.Literal); ok {
						var aliasLit *ast.Literal
						if p.peek().Type == token.AS {
							p.advance()
							aliasExpr := p.parsePrimaryExpression()
							if al, ok := aliasExpr.(*ast.Literal); ok {
								aliasLit = al
							} else {
								panic(fmt.Sprintf("expected alias literal, got %s", aliasExpr.String()))
							}
						}
						specifiers = append(specifiers, ast.NewUseSpecifier(lit, aliasLit))
					} else {
						panic(fmt.Sprintf("expected literal, got %s", remoteExpr.String()))
					}
					if p.peek().Type == token.COMMA {
						p.advance()
					}
				}
				p.expect(token.RPAREN)
				return ast.NewUseDecl(source, specifiers, token.PICK)
			} else {
				remoteExpr := p.parsePrimaryExpression()
				if lit, ok := remoteExpr.(*ast.Literal); ok {
					var aliasLit *ast.Literal
					if p.peek().Type == token.AS {
						p.advance()
						aliasExpr := p.parsePrimaryExpression()
						if al, ok := aliasExpr.(*ast.Literal); ok {
							aliasLit = al
						} else {
							panic(fmt.Sprintf("expected alias literal, got %s", aliasExpr.String()))
						}
					}
					specifiers = append(specifiers, ast.NewUseSpecifier(lit, aliasLit))
				} else {
					panic(fmt.Sprintf("expected literal, got %s", remoteExpr.String()))
				}
				return ast.NewUseDecl(source, specifiers, token.PICK)
			}
		} else {
			return ast.NewUseDecl(source, specifiers, token.USE)
		}
	})

	c.RegisterStmtHandler(token.IF, func(p *Parser) any {
		p.advance() // skip 'if'
		condition := p.parseExpression()
		p.expect(token.COLON)
		var body []ast.Stmt
		for !p.isEof() && !slices.Contains([]token.TokenType{token.END, token.ELSE}, p.peek().Type) {
			stmt := p.parseStatement()
			if stmt != nil {
				body = append(body, stmt)
			}
		}
		if p.peek().Type == token.ELSE {
			p.advance() // skip 'else'
			if p.peek().Type == token.IF {
				return ast.NewIfStmt(condition, ast.NewBlockStmt(body), p.CallStmtHandler(token.IF))
			}
			return ast.NewIfStmt(condition, ast.NewBlockStmt(body), p.parseBlockStatement())
		} else {
			p.expect(token.END)
		}
		return ast.NewIfStmt(condition, ast.NewBlockStmt(body), nil)
	})

	c.RegisterStmtHandler(token.FOR, func(p *Parser) any {
		p.advance() // skip 'for'
		var firstExpr ast.Expr

		if p.peek().Type == token.LET {
			firstExpr = p.parseStatement()
		} else {
			firstExpr = p.parseExpression()
		}

		var body *ast.BlockStmt
		// for i in xxx
		if p.peek().Type == token.IN {
			p.advance() // skip 'in'
			iter := p.parseExpression()
			body = p.parseBlockStatement()
			return ast.NewForStmt(firstExpr, nil, nil, iter, *body)
		}
		// for i := 0; i < 10; i++ :
		p.expect(token.SEMICOLON)
		secondExpr := p.parseCompareExpression()
		p.expect(token.SEMICOLON)
		thirdExpr := p.parseExpression()
		body = p.parseBlockStatement()
		return ast.NewForStmt(firstExpr, secondExpr, thirdExpr, nil, *body)
	})

	c.RegisterStmtHandler(token.FN, func(p *Parser) any {
		p.advance() // skip 'fn'
		id := p.expect(token.IDENT)
		var args = ast.NewArgsExpr([]ast.Expr{})
		if p.peek().Type == token.LPAREN {
			p.expect(token.LPAREN)
			args = p.parseArgs()
			p.expect(token.RPAREN)
		}
		return ast.NewFunctionDecl(p.createLiteral(id), args, p.parseBlockStatement())
	})

	c.RegisterStmtHandler(token.EXPOSE, func(p *Parser) any {
		p.advance()

		switch p.peek().Type {
		case token.FN:
			decl := p.CallStmtHandler(token.FN)
			return ast.NewExposeStmt(decl, nil, nil)
		case token.LET, token.CST:
			decl := p.CallStmtHandler(p.peek().Type)
			return ast.NewExposeStmt(decl, nil, nil)
		case token.IDENT:
			idTk := p.expect(token.IDENT)
			name := p.createLiteral(idTk)
			if p.peek().Type == token.ASSIGN {
				p.advance()
				value := p.parseExpression()
				return ast.NewExposeStmt(nil, name, value)
			}
			return ast.NewExposeStmt(nil, name, nil)
		default:
			p.errorf(p.peek(), "unexpected token after expose: %s", p.peek().String())
			return nil
		}
	})

	c.RegisterStmtHandler(token.SWITCH, func(p *Parser) any {
		p.advance() // skip 'switch'
		condition := p.parseExpression()
		p.expect(token.COLON)
		var isDefinedDefault bool = false
		var cases []ast.Expr
		for !p.isEof() && p.peek().Type != token.END {
			var expr = p.parseSwitchCase()
			if isDefinedDefault {
				panic("default case already defined")
			}
			if expr != nil {
				cases = append(cases, expr)
				if expr.(*ast.SwitchCase).IsDefault {
					isDefinedDefault = true
				}
			}
		}
		p.expect(token.END)
		return ast.NewSwitchStmt(condition, cases)
	})

	c.RegisterStmtHandler(token.RETURN, func(p *Parser) any {
		p.advance() // skip 'return'
		return ast.NewReturnStmt(p.parseExpression())
	})

	c.RegisterStmtHandler(token.TASK, func(p *Parser) any {
		p.advance() // skip 'task'
		fn := p.parseStatement()
		if fn != nil {
			if f, ok := fn.(*ast.FunctionDecl); ok {
				return ast.NewTaskStmt(*f)
			}
		}
		panic(verror.ParseVError{
			Position: p.peek().ToPosition(""),
			Message:  "expected function declaration",
		})
	})

	c.RegisterStmtHandler(token.WAIT, func(p *Parser) any {
		p.advance() // skip 'wait'
		return ast.NewWaitStmt(p.parseExpression())
	})

	c.RegisterStmtHandler(token.BREAK, func(p *Parser) any {
		p.advance() // skip 'break'
		return ast.NewBreakStmt()
	})

	c.RegisterStmtHandler(token.CONTINUE, func(p *Parser) any {
		p.advance() // skip 'continue'
		return ast.NewContinueStmt()
	})

	return c
}
