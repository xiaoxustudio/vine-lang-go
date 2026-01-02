package verror

import (
	"fmt"
)

type Position struct {
	Filename string
	Line     int
	Column   int
}

type VError interface {
	GetPosition() Position
	Error() string
}

type LexerVError struct {
	Position
	Message string
}

func (pv *LexerVError) GetPosition() Position {
	return pv.Position
}

func (pv *LexerVError) Error() string {
	return fmt.Sprintf("[Line %d, Column %d] Lexer Error: %s", pv.Line, pv.Column, pv.Message)
}

type ParseVError struct {
	Position
	Message string
}

func (pv *ParseVError) GetPosition() Position {
	return pv.Position
}

func (pv *ParseVError) Error() string {
	return fmt.Sprintf("[Line %d, Column %d] Parser Error: %s", pv.Line, pv.Column, pv.Message)
}

type InterpreterVError struct {
	Position
	Message string
}

func (e *InterpreterVError) GetPosition() Position {
	return e.Position
}

func (pv *InterpreterVError) Error() string {
	return fmt.Sprintf("[Line %d, Column %d] Interpreter Error: %s", pv.Line, pv.Column, pv.Message)
}
