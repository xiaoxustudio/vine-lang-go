package ipt

import (
	"fmt"
	"strconv"
	"vine-lang/ast"
	"vine-lang/env"
	"vine-lang/lexer"
	"vine-lang/parser"
)

type InterpreterError struct {
	Line    int
	Column  int
	Message string
}

type Interpreter struct {
	errors []InterpreterError
	p      *parser.Parser
	env    *env.Environment
}

func New(p *parser.Parser, env *env.Environment) *Interpreter {
	return &Interpreter{
		errors: make([]InterpreterError, 0),
		p:      p,
		env:    env,
	}
}

func (i *Interpreter) Eval(node ast.Node, env *env.Environment) any {
	switch n := node.(type) {
	case *ast.ProgramStmt:
		var lastResult any
		for _, s := range n.Body {
			lastResult = i.Eval(s, env)
		}
		return lastResult
	case *ast.VariableDecl:
		val := i.Eval(n.Value, env)
		env.Set(n.Name.Value, val)
		return nil
	case *ast.Literal:
		return n.Value
	case *ast.BinaryExpr:
		{
			left := i.Eval(n.Left, env).(Token)
			right := i.Eval(n.Right, env).(Token)
			if left.Type != lexer.NUMBER || right.Type != lexer.NUMBER {
				panic(InterpreterError{
					Line:    n.Operator.Line,
					Column:  n.Operator.Column,
					Message: "Both sides of the operator must be numbers",
				})
			}

			leftValue, err := strconv.ParseFloat(left.Value, 64)
			if err != nil {
				panic(err)
			}

			rightValue, err := strconv.ParseFloat(right.Value, 64)
			if err != nil {
				panic(err)
			}

			switch n.Operator.Type {
			case lexer.PLUS:
				return leftValue + rightValue
			case lexer.MINUS:
				return leftValue - rightValue
			case lexer.MUL:
				return leftValue * rightValue
			case lexer.DIV:
				return leftValue / rightValue
			}
		}
	}
	return nil
}

func (i *Interpreter) EvalSafe() any {
	defer func() {
		if r := recover(); r != nil {
			if parseErr, ok := r.(InterpreterError); ok {
				i.errors = append(i.errors, parseErr)
				fmt.Printf("Caught Error: %v\n", parseErr)
			} else {
				panic(r)
			}
		}
	}()

	return i.Eval(i.p.ParseProgram(), i.env)
}
