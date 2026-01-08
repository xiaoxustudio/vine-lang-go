package token

import (
	"fmt"
	"strings"
	"vine-lang/verror"
)

type TokenType string

const (
	EOF TokenType = "EOF"

	NEWLINE    TokenType = "\n"
	WHITESPACE TokenType = "WHITESPACE"
	ILLEGAL    TokenType = "ILLEGAL"

	// Identifiers and literals
	IDENT  TokenType = "IDENT"
	NUMBER TokenType = "NUMBER"
	STRING TokenType = "STRING"

	// Operators
	ASSIGN TokenType = "="
	PLUS   TokenType = "+"
	MINUS  TokenType = "-"
	BANG   TokenType = "!"
	MUL    TokenType = "*"
	DIV    TokenType = "/"
	QUOTE  TokenType = "\""

	// Logical operators
	AND        TokenType = "AND"
	OR         TokenType = "OR"
	NOT        TokenType = "NOT"
	EQ         TokenType = "=="
	NOT_EQ     TokenType = "!="
	LESS_EQ    TokenType = "<="
	GREATER_EQ TokenType = ">="
	LESS       TokenType = "<"
	GREATER    TokenType = ">"
	INC        TokenType = "++"
	DEC        TokenType = "--"
	INC_EQ     TokenType = "+="
	DEC_EQ     TokenType = "-="
	MUL_EQ     TokenType = "*="
	DIV_EQ     TokenType = "/="

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	DOT       TokenType = "."
	COLON     TokenType = ":"
	QUESTION  TokenType = "?"

	LPAREN TokenType = "("
	RPAREN TokenType = ")"
	LBRACE TokenType = "{"
	RBRACE TokenType = "}"

	// Keywords
	FUNCTION TokenType = "FUNCTION"
	LET      TokenType = "LET"
	CST      TokenType = "CST" // const keywords
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	RETURN   TokenType = "RETURN"
	FOR      TokenType = "FOR"
	WHILE    TokenType = "WHILE"
	IN       TokenType = "IN"
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"
	USE      TokenType = "USE"
	AS       TokenType = "AS"
	task     TokenType = "TASK"
	EXPOSE   TokenType = "EXPOSE"
	TYPEOF   TokenType = "TYPEOF"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	NIL      TokenType = "NIL"
	PICK     TokenType = "PICK"

	/* Inside Tag */
	Module TokenType = "__Module_TAG__"
)

var Keywords = map[string]TokenType{
	"fn":       FUNCTION,
	"let":      LET,
	"cst":      CST,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"for":      FOR,
	"while":    WHILE,
	"in":       IN,
	"break":    BREAK,
	"continue": CONTINUE,
	"use":      USE,
	"as":       AS,
	"task":     task,
	"expose":   EXPOSE,
	"typeof":   TYPEOF,
	"true":     TRUE,
	"false":    FALSE,
	"nil":      NIL,
}

type Token struct {
	Type   TokenType
	Value  string // 字符串原始值
	Line   int    // 行号
	Column int    // 列号
}

func NewToken(t TokenType, v rune, col, line int) Token {
	return Token{
		Type:   t,
		Value:  string(v),
		Column: col,
		Line:   line,
	}
}

func (t Token) String() string {
	if t.Type == NEWLINE {
		return fmt.Sprintf("Token{NEWLINE(\\n), %d, %d}", t.Line, t.Column)
	}
	return fmt.Sprintf("Token{%s(%s), %d, %d}", t.Type, t.Value, t.Line, t.Column)
}

func (t Token) ToPosition(fname string) verror.Position {
	return verror.Position{Filename: fname, Line: t.Line, Column: t.Column}
}

func (t Token) IsEmpty() bool {
	return t == (Token{})
}

func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
