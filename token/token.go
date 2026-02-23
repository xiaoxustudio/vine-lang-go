package token

import (
	"fmt"
	"strconv"
	"strings"
	"vine-lang/verror"
)

type TokenType string

const (
	EOF TokenType = "EOF"

	NEWLINE    TokenType = "\n"
	WHITESPACE TokenType = "WHITESPACE"
	ILLEGAL    TokenType = "ILLEGAL"
	COMMENT    TokenType = "COMMENT"

	// Identifiers and literals
	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	FLOAT  TokenType = "FLOAT"
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

	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LBRACE   TokenType = "{"
	RBRACE   TokenType = "}"
	LBRACKET TokenType = "["
	RBRACKET TokenType = "]"

	// Keywords
	FN       TokenType = "FN"
	END      TokenType = "END"
	LET      TokenType = "LET"
	CST      TokenType = "CST" // const keywords
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	RETURN   TokenType = "RETURN"
	FOR      TokenType = "FOR"
	IN       TokenType = "IN"
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"
	USE      TokenType = "USE"
	AS       TokenType = "AS"
	TASK     TokenType = "TASK"
	EXPOSE   TokenType = "EXPOSE"
	TYPEOF   TokenType = "TYPEOF"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	NIL      TokenType = "NIL"
	PICK     TokenType = "PICK"
	SWITCH   TokenType = "SWITCH"
	CASE     TokenType = "CASE"
	DEFAULT  TokenType = "DEFAULT"

	/* Inside Tag */
	Module TokenType = "__Module_TAG__"
)

var Keywords = map[string]TokenType{
	"fn":       FN,
	"let":      LET,
	"cst":      CST,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"for":      FOR,
	"in":       IN,
	"break":    BREAK,
	"continue": CONTINUE,
	"use":      USE,
	"as":       AS,
	"pick":     PICK,
	"task":     TASK,
	"expose":   EXPOSE,
	"typeof":   TYPEOF,
	"true":     TRUE,
	"false":    FALSE,
	"nil":      NIL,
	"end":      END,
	"switch":   SWITCH,
	"default":  DEFAULT,
	"case":     CASE,
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

func NewTokenDuplicated(t TokenType, v rune, col, line int, vv rune) Token {
	return Token{
		Type:   t,
		Value:  string(v) + string(vv),
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

func (t *Token) GetInt() (int64, error) {
	if t.Type != INT {
		return 0, fmt.Errorf("token is not INT type")
	}
	i, err := strconv.ParseInt(t.Value, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (t *Token) GetFloat() (float64, error) {
	if t.Type != FLOAT {
		return 0, fmt.Errorf("token is not FLOAT type")
	}
	f, err := strconv.ParseFloat(t.Value, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}

func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
