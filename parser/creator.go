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

	c.RegisterStmtHandler(token.LET, func(p *Parser) any {
		startToken := p.advance()
		isConst := startToken.Type == token.CST

		idTk := p.expect(token.IDENT)

		id := p.createLiteral(idTk)
		var typeName *ast.Literal
		if p.peek().Type == token.IDENT && p.peek().Line == idTk.Line {
			typeTk := p.advance()
			typeName = &ast.Literal{Value: &typeTk}
		}
		var value ast.Expr = nil
		if p.peek().Type == token.ASSIGN {
			p.expect(token.ASSIGN)
			value = p.parseExpression()
		}

		return &ast.VariableDecl{
			Name:     id,
			TypeName: typeName,
			Value:    value,
			IsConst:  isConst,
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
		firstExpr := p.parseStatement()
		var body *ast.BlockStmt
		// for i := range xxx
		if p.peek().Type == token.COLON {
			body = p.parseBlockStatement()
			return &ast.ForStmt{
				Body:  body,
				Range: firstExpr,
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

	return c
}
