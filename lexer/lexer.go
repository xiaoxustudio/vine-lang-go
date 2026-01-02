package lexer

import (
	"fmt"
	"vine-lang/utils"
	"vine-lang/verror"
)

type Lexer struct {
	input    string
	position int // current position in input (points to current char)
	column   int // current column in input
	line     int // current line in input
	ch       string
	tokens   []Token // list of tokens
	filename string
}

func New(filename string, input string) *Lexer {
	l := &Lexer{input: input + "\n", filename: filename, line: 1, position: -1}
	l.readChar()
	return l
}

func (l *Lexer) isEof() bool {
	return l.position >= len(l.input)-1
}

func (l *Lexer) readChar() {
	if l.isEof() {
		l.ch = string(EOF)
	} else {
		if l.ch != "\n" {
			// 不是换行符，则列数加1
			l.column += 1
		}
		l.position += 1
		l.ch = string(l.input[l.position])
	}
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
	for utils.IsDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) GetToken() (Token, error) {
	var tok Token
	var peek string = ""
	if !l.isEof() {
		peek = string(l.input[l.position+1])
	}

	switch l.ch {
	case ",":
		tok = NewToken(COMMA, l.ch, l.column, l.line)
	case ":":
		tok = NewToken(COLON, l.ch, l.column, l.line)
	case ".":
		tok = NewToken(DOT, l.ch, l.column, l.line)
	case "?":
		tok = NewToken(QUESTION, l.ch, l.column, l.line)
	case "+":
		switch peek {
		case "+":
			tok = NewToken(INC, l.ch+peek, l.column, l.line)
			l.readChar()
		case "=":
			tok = NewToken(INC_EQ, l.ch+peek, l.column, l.line)
			l.readChar()
		default:
			tok = NewToken(PLUS, l.ch, l.column, l.line)
		}
	case "-":
		switch peek {
		case "-":
			tok = NewToken(DEC, l.ch+peek, l.column, l.line)
			l.readChar()
		case "=":
			tok = NewToken(DEC_EQ, l.ch+peek, l.column, l.line)
			l.readChar()
		default:
			if utils.IsDigit(peek) {
				l.readChar()
				num := l.readNumber()
				tok = NewToken(NUMBER, "-"+num, l.column, l.line)
			} else {
				tok = NewToken(MINUS, l.ch, l.column, l.line)
			}
		}
	case "*":
		switch peek {
		case "=":
			tok = NewToken(MUL_EQ, l.ch+peek, l.column, l.line)
			l.readChar()
		default:
			tok = NewToken(MUL, l.ch, l.column, l.line)
		}
	case "/":
		switch peek {
		case "=":
			tok = NewToken(DIV_EQ, l.ch+peek, l.column, l.line)
			l.readChar()
		default:
			tok = NewToken(DIV, l.ch, l.column, l.line)
		}
	case "=":
		if peek == "=" {
			tok = NewToken(EQ, l.ch+peek, l.column, l.line)
			l.readChar()
		} else {
			tok = NewToken(ASSIGN, l.ch, l.column, l.line)
		}
	case "!":
		if peek == "=" {
			tok = NewToken(NOT_EQ, l.ch+peek, l.column, l.line)
			l.readChar()
		} else {
			tok = NewToken(BANG, l.ch, l.column, l.line)
		}
	case "<":
		if peek == "=" {
			tok = NewToken(LESS_EQ, l.ch+peek, l.column, l.line)
			l.readChar()
		} else {
			tok = NewToken(LESS, l.ch, l.column, l.line)
		}
	case ">":
		if peek == "=" {
			tok = NewToken(GREATER_EQ, l.ch+peek, l.column, l.line)
			l.readChar()
		} else {
			tok = NewToken(GREATER, l.ch, l.column, l.line)
		}
	case "(":
		tok = NewToken(LPAREN, l.ch, l.column, l.line)
	case ")":
		tok = NewToken(RPAREN, l.ch, l.column, l.line)
	case "{":
		tok = NewToken(LBRACE, l.ch, l.column, l.line)
	case "}":
		tok = NewToken(RBRACE, l.ch, l.column, l.line)
	case ";":
		tok = NewToken(SEMICOLON, l.ch, l.column, l.line)
	case " ":
		tok = NewToken(WHITESPACE, l.ch, l.column, l.line)
	case "\t":
		tok = NewToken(WHITESPACE, l.ch, l.column, l.line)
	case "\r":
		if peek == "\n" {
			tok = NewToken(NEWLINE, l.ch, l.column, l.line)
			l.line += 1
			l.column = 0
			l.readChar()
		} else {
			// unknown \r token
			return NewToken(ILLEGAL, l.ch, l.column, l.line), &verror.LexerVError{
				Position: verror.Position{
					Filename: l.filename,
					Line:     l.line,
					Column:   l.column,
				},
				Message: "the Lexer parse with expected token",
			}
		}
	case "\n":
		tok = NewToken(NEWLINE, l.ch, l.column, l.line)
		l.line += 1
		l.column = 0
	case "\"":
		pos := l.position
		l.readChar()
		for l.ch != "\"" {
			l.readChar()
		}
		tok = NewToken(STRING, l.input[pos:l.position], l.column, l.line)
	default:
		if utils.IsIdentifier(l.ch) {
			tok.Value = l.readIdentifier()
			switch tok.Value {
			case "true":
				tok.Type = TRUE
			case "false":
				tok.Type = FALSE
			}
			tok.Type = LookupIdent(tok.Value)
			tok.Column = l.column
			tok.Line = l.line
			return tok, nil
		} else if utils.IsDigit(l.ch) {
			tok.Value = l.readNumber()
			tok.Type = NUMBER
			tok.Column = l.column
			tok.Line = l.line
			return tok, nil
		}
		return NewToken(ILLEGAL, l.ch, l.column, l.line), &verror.LexerVError{
			Position: verror.Position{
				Filename: l.filename,
				Line:     l.line,
				Column:   l.column,
			},
			Message: "the Lexer parse with unknown token",
		}
	}
	l.readChar() // 移动到下一个字符
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

func (l *Lexer) Tokens() []Token {
	return l.tokens
}

func (l *Lexer) TheEof() Token {
	var lastTk = l.tokens[len(l.tokens)-1]
	return Token{
		Type:   EOF,
		Value:  string(EOF),
		Column: lastTk.Column,
		Line:   lastTk.Line,
	}
}

func (l *Lexer) Print() {
	for i := range l.tokens {
		switch l.tokens[i].Type {
		case ILLEGAL:
			continue
		case NEWLINE:
			fmt.Println()
		case WHITESPACE:
			fmt.Print(" ")
		default:
			fmt.Print(l.tokens[i].String())
		}
	}
}
