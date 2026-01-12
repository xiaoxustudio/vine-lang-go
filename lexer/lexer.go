package lexer

import (
	"fmt"
	"unicode/utf8"
	"vine-lang/token"
	"vine-lang/utils"
	"vine-lang/verror"
)

type Lexer struct {
	input    string
	position int // current position in input (points to current char)
	column   int // current column in input
	line     int // current line in input
	ch       rune
	chWidth  int           // width of current rune in bytes
	tokens   []token.Token // list of tokens
	filename string
}

func New(filename string, input string) *Lexer {
	l := &Lexer{input: input, filename: filename, line: 1, position: 0}
	l.readChar() // Initialize l.ch to the first character
	return l
}

func (l *Lexer) isEof() bool {
	return l.position >= len(l.input)
}

func (l *Lexer) readChar() {
	if l.chWidth > 0 {
		l.position += l.chWidth
	}

	if l.position >= len(l.input) {
		l.ch = 0
		l.chWidth = 0
		return
	}

	l.ch, l.chWidth = utf8.DecodeRuneInString(l.input[l.position:])

	if l.ch != '\n' {
		l.column += 1
	}
}

func (l *Lexer) peekRune() rune {
	nextPos := l.position + l.chWidth
	if nextPos >= len(l.input) {
		return 0
	}

	// 只解码，不移动指针
	r, _ := utf8.DecodeRuneInString(l.input[nextPos:])
	return r
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for utils.IsIdentifier(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	hasDecimal := false
	for utils.IsDigit(l.ch) {
		if l.ch == '.' {
			if hasDecimal {
				break
			}
			hasDecimal = true
		}
		l.readChar()
	}
	numStr := l.input[position:l.position]
	return numStr
}

func (l *Lexer) GetToken() (token.Token, error) {
	var tok token.Token
	switch l.ch {
	case '#':
		l.readChar()
		pos := l.position
		for l.ch != '\n' && !l.isEof() {
			l.readChar()
		}
		tok.Value = l.input[pos:l.position]
		tok.Column = l.column
		tok.Line = l.line
		tok.Type = token.COMMENT
		return tok, nil
	case ',':
		tok = token.NewToken(token.COMMA, l.ch, l.column, l.line)
	case ':':
		tok = token.NewToken(token.COLON, l.ch, l.column, l.line)
	case '.':
		tok = token.NewToken(token.DOT, l.ch, l.column, l.line)
	case '?':
		tok = token.NewToken(token.QUESTION, l.ch, l.column, l.line)
	case '+':
		peek := l.peekRune()
		switch peek {
		case '+':
			tok = token.NewTokenDuplicated(token.INC, l.ch, l.column, l.line, peek)
			l.readChar()
		case '=':
			tok = token.NewTokenDuplicated(token.INC_EQ, l.ch, l.column, l.line, peek)
			l.readChar()
		default:
			tok = token.NewToken(token.PLUS, l.ch, l.column, l.line)
		}
	case '-':
		peek := l.peekRune()
		switch peek {
		case '-':
			tok = token.NewTokenDuplicated(token.DEC, l.ch, l.column, l.line, peek)
			l.readChar()
		case '=':
			tok = token.NewTokenDuplicated(token.DEC_EQ, l.ch, l.column, l.line, peek)
			l.readChar()
		default:
			if utils.IsDigit(peek) {
				l.readChar()
				num := l.readNumber()
				tok.Value = "-" + num
				tok.Column = l.column
				tok.Line = l.line
				tok.Type = token.NUMBER
			} else {
				tok = token.NewToken(token.MINUS, l.ch, l.column, l.line)
			}
		}
	case '*':
		peek := l.peekRune()
		switch peek {
		case '=':
			tok = token.NewTokenDuplicated(token.MUL_EQ, l.ch, l.column, l.line, peek)
			l.readChar()
		default:
			tok = token.NewToken(token.MUL, l.ch, l.column, l.line)
		}
	case '/':
		peek := l.peekRune()
		switch peek {
		case '=':
			tok = token.NewTokenDuplicated(token.DIV_EQ, l.ch, l.column, l.line, peek)
			l.readChar()
		default:
			tok = token.NewToken(token.DIV, l.ch, l.column, l.line)
		}
	case '=':
		peek := l.peekRune()
		if peek == '=' {
			tok = token.NewTokenDuplicated(token.EQ, l.ch, l.column, l.line, peek)
			l.readChar()
		} else {
			tok = token.NewToken(token.ASSIGN, l.ch, l.column, l.line)
		}
	case '!':
		peek := l.peekRune()
		if peek == '=' {
			tok = token.NewTokenDuplicated(token.NOT_EQ, l.ch, l.column, l.line, peek)
			l.readChar()
		} else {
			tok = token.NewToken(token.BANG, l.ch, l.column, l.line)
		}
	case '<':
		peek := l.peekRune()
		if peek == '=' {
			tok = token.NewTokenDuplicated(token.LESS_EQ, l.ch, l.column, l.line, peek)
			l.readChar()
		} else {
			tok = token.NewToken(token.LESS, l.ch, l.column, l.line)
		}
	case '>':
		peek := l.peekRune()
		if peek == '=' {
			tok = token.NewTokenDuplicated(token.GREATER_EQ, l.ch, l.column, l.line, peek)
			l.readChar()
		} else {
			tok = token.NewToken(token.GREATER, l.ch, l.column, l.line)
		}
	case '(':
		tok = token.NewToken(token.LPAREN, l.ch, l.column, l.line)
	case ')':
		tok = token.NewToken(token.RPAREN, l.ch, l.column, l.line)
	case '{':
		tok = token.NewToken(token.LBRACE, l.ch, l.column, l.line)
	case '}':
		tok = token.NewToken(token.RBRACE, l.ch, l.column, l.line)
	case ';':
		tok = token.NewToken(token.SEMICOLON, l.ch, l.column, l.line)
	case ' ':
		tok = token.NewToken(token.WHITESPACE, l.ch, l.column, l.line)
	case '\t':
		tok = token.NewToken(token.WHITESPACE, l.ch, l.column, l.line)
	case '\r':
		peek := l.peekRune()
		if peek == '\n' {
			tok = token.NewToken(token.NEWLINE, l.ch, l.column, l.line)
			l.readChar()
		} else {
			// unknown \r token
			return token.NewToken(token.ILLEGAL, l.ch, l.column, l.line), &verror.LexerVError{
				Position: verror.Position{
					Filename: l.filename,
					Line:     l.line,
					Column:   l.column,
				},
				Message: "the Lexer parse with expected token",
			}
		}
	case '\n':
		tok = token.NewToken(token.NEWLINE, l.ch, l.column, l.line)
		l.line += 1
		l.column = 0
	case '"':
		l.readChar()
		pos := l.position
		for l.ch != '"' {
			l.readChar()
		}
		tok.Value = l.input[pos:l.position]
		tok.Column = l.column
		tok.Line = l.line
		tok.Type = token.STRING
	default:
		if utils.IsIdentifier(l.ch) {
			tok.Value = l.readIdentifier()
			switch tok.Value {
			case "true":
				tok.Type = token.TRUE
			case "false":
				tok.Type = token.FALSE
			}
			tok.Type = token.LookupIdent(tok.Value)
			tok.Column = l.column
			tok.Line = l.line
			return tok, nil
		} else if utils.IsDigit(l.ch) {
			tok.Value = l.readNumber()
			tok.Type = token.NUMBER
			tok.Column = l.column
			tok.Line = l.line
			return tok, nil
		}
		return token.NewToken(token.ILLEGAL, l.ch, l.column, l.line), &verror.LexerVError{
			Position: verror.Position{
				Filename: l.filename,
				Line:     l.line,
				Column:   l.column,
			},
			Message: fmt.Sprintf("the Lexer parse with unexpected token: %q", l.ch),
		}
	}
	l.readChar()
	return tok, nil
}

func (l *Lexer) Parse() {
	for !l.isEof() {
		tok, err := l.GetToken()
		if err != nil {
			panic(err)
		}
		l.tokens = append(l.tokens, tok)
	}
}

func (l *Lexer) Tokens() []token.Token {
	return l.tokens
}

func (l *Lexer) TheEof() token.Token {
	var lastTk = l.tokens[len(l.tokens)-1]
	return token.Token{
		Type:   token.EOF,
		Value:  string(token.EOF),
		Column: lastTk.Column,
		Line:   lastTk.Line,
	}
}

func (l *Lexer) Print() {
	for i := range l.tokens {
		switch l.tokens[i].Type {
		case token.ILLEGAL:
			continue
		case token.NEWLINE:
			fmt.Println()
		case token.WHITESPACE:
			fmt.Print(" ")
		default:
			fmt.Print(l.tokens[i].String())
		}
	}
}
