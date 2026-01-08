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
	error
	GetPosition() Position
	Error() string
}

type LexerVError struct {
	VError
	Position
	Message string
}

func (pv *LexerVError) GetPosition() Position {
	return pv.Position
}

func (pv LexerVError) Error() string {
	return fmt.Sprintf("[Line %d, Column %d] Lexer Error: %s", pv.Line, pv.Column, pv.Message)
}

type ParseVError struct {
	VError
	Position
	Message string
}

func (pv *ParseVError) GetPosition() Position {
	return pv.Position
}

func (pv ParseVError) Error() string {
	return fmt.Sprintf("[Line %d, Column %d] Parser Error: %s", pv.Line, pv.Column, pv.Message)
}

type InterpreterVError struct {
	VError
	Position
	Message string
}

func (e InterpreterVError) Error() string {
	return fmt.Sprintf("[Line %d, Column %d] Interpreter Error: %s", e.Line, e.Column, e.Message)
}

func (e *InterpreterVError) GetPosition() Position {
	return e.Position
}
