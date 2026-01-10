package parser

import (
	"fmt"
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

		id := &ast.Literal{Value: idTk}
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
			}
		} else if p.peek().Type == token.PICK {
			p.advance() // skip 'pick'
			if p.peek().Type == token.LPAREN {
				for p.peek().Type != token.RPAREN {
					if p.peek().Type == token.COMMA {
						p.advance() // skip ','
					}
					alias := p.parsePrimaryExpression()
					specifiers = append(specifiers, alias)
				}
				p.expect(token.RPAREN)
				return &ast.UseDecl{
					Specifiers: specifiers,
					Source:     source,
				}
			} else {
				alias := p.parsePrimaryExpression()
				specifiers = append(specifiers, alias)
				return &ast.UseDecl{
					Specifiers: specifiers,
					Source:     source,
				}
			}
		} else {
			return &ast.UseDecl{
				Specifiers: specifiers,
				Source:     source,
			}
		}
	})

	return c
}
