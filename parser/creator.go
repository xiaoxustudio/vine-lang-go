package parser

import (
	"fmt"
	"slices"
	"vine-lang/ast"
	"vine-lang/lexer"
	"vine-lang/token"
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

		id := &ast.Literal{Value: &idTk}
		p.expect(token.ASSIGN)

		value := p.parseExpression()

		return &ast.VariableDecl{
			Name:    id,
			Value:   value,
			IsConst: isConst,
		}
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
			return &ast.UseDecl{
				Specifiers: specifiers,
				Source:     source,
				Mode:       token.AS,
			}
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
						specifiers = append(specifiers, &ast.UseSpecifier{Remote: lit, Local: aliasLit})
					} else {
						panic(fmt.Sprintf("expected literal, got %s", remoteExpr.String()))
					}
					if p.peek().Type == token.COMMA {
						p.advance()
					}
				}
				p.expect(token.RPAREN)
				return &ast.UseDecl{
					Specifiers: specifiers,
					Source:     source,
					Mode:       token.PICK,
				}
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
					specifiers = append(specifiers, &ast.UseSpecifier{Remote: lit, Local: aliasLit})
				} else {
					panic(fmt.Sprintf("expected literal, got %s", remoteExpr.String()))
				}
				return &ast.UseDecl{
					Specifiers: specifiers,
					Source:     source,
					Mode:       token.PICK,
				}
			}
		} else {
			return &ast.UseDecl{
				Specifiers: specifiers,
				Source:     source,
				Mode:       token.USE,
			}
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
				return &ast.IfStmt{Test: condition, Consequent: &ast.BlockStmt{Body: body}, Alternate: p.CallStmtHandler(token.IF)}
			}
			return &ast.IfStmt{Test: condition, Consequent: &ast.BlockStmt{Body: body}, Alternate: p.parseBlockStatement()}
		} else {
			p.expect(token.END)
		}
		return &ast.IfStmt{Test: condition, Consequent: &ast.BlockStmt{Body: body}}
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
			return &ast.ForStmt{
				Body:  body,
				Init:  firstExpr,
				Range: iter,
			}
		}
		// for i := 0; i < 10; i++ :
		p.expect(token.SEMICOLON)
		secondExpr := p.parseCompareExpression()
		p.expect(token.SEMICOLON)
		thirdExpr := p.parseExpression()
		body = p.parseBlockStatement()
		return &ast.ForStmt{
			Body:   body,
			Init:   firstExpr,
			Value:  secondExpr,
			Update: thirdExpr,
		}
	})

	c.RegisterStmtHandler(token.FN, func(p *Parser) any {
		p.advance() // skip 'fn'
		id := p.expect(token.IDENT)
		p.expect(token.LPAREN)
		args := p.parseArgs()
		return &ast.FunctionDecl{
			ID:        p.createLiteral(id),
			Arguments: args,
			Body:      p.parseBlockStatement(),
		}
	})

	c.RegisterStmtHandler(token.EXPOSE, func(p *Parser) any {
		p.advance()

		switch p.peek().Type {
		case token.FN:
			decl := p.CallStmtHandler(token.FN)
			return &ast.ExposeStmt{Decl: decl}
		case token.LET, token.CST:
			decl := p.CallStmtHandler(p.peek().Type)
			return &ast.ExposeStmt{Decl: decl}
		case token.IDENT:
			idTk := p.expect(token.IDENT)
			name := p.createLiteral(idTk)
			if p.peek().Type == token.ASSIGN {
				p.advance()
				value := p.parseExpression()
				return &ast.ExposeStmt{Name: name, Value: value}
			}
			return &ast.ExposeStmt{Name: name}
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
		return &ast.SwitchStmt{Test: condition, Cases: cases}
	})

	c.RegisterStmtHandler(token.RETURN, func(p *Parser) any {
		p.advance() // skip 'return'
		return &ast.ReturnStmt{Value: p.parseExpression()}
	})

	c.RegisterStmtHandler(token.TASK, func(p *Parser) any {
		p.advance() // skip 'task'
		fn := p.parseStatement()
		return &ast.TaskStmt{
			Fn: fn.(*ast.FunctionDecl),
		}
	})

	c.RegisterStmtHandler(token.WAIT, func(p *Parser) any {
		p.advance() // skip 'wait'
		return &ast.WaitStmt{Async: p.parseExpression()}
	})

	return c
}
