package ipt

import (
	"fmt"
	"vine-lang/ast"
	"vine-lang/env"
	"vine-lang/lexer"
	"vine-lang/libs/global"
	"vine-lang/parser"
	"vine-lang/verror"
)

type Interpreter struct {
	errors []verror.InterpreterVError
	p      *parser.Parser
	env    *env.Environment
}

func New(p *parser.Parser, env *env.Environment) *Interpreter {
	return &Interpreter{
		errors: make([]verror.InterpreterVError, 0),
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
	case *ast.ExpressionStmt:
		return i.Eval(n.Expression, env)
	case *ast.VariableDecl:
		val := i.Eval(n.Value, env)
		env.Define(n.Name.Value, val)
		return nil
	case *ast.AssignmentExpr:
		name := i.Eval(n.Left, env)
		val := i.Eval(n.Right, env)
		env.Set(name.(Token), val)
		return nil
	case *ast.BinaryExpr:
		{
			leftRaw := i.Eval(n.Left, env)
			rightRaw := i.Eval(n.Right, env)
			leftVal, isLeftInt, err := global.GetNumberAndType(leftRaw)
			if err != nil {
				panic(verror.InterpreterVError{
					Message: err.Error(),
					Position: verror.Position{
						Column: n.Operator.Column,
						Line:   n.Operator.Line,
					},
				})
			}

			/* 数字转换 */
			rightVal, isRightInt, err := global.GetNumberAndType(rightRaw)
			if err != nil {
				panic(verror.InterpreterVError{
					Position: verror.Position{
						Column: n.Operator.Column,
						Line:   n.Operator.Line,
					},
					Message: err.Error(),
				})
			}

			var result any

			if isLeftInt && isRightInt {
				lInt := int64(leftVal)
				rInt := int64(rightVal)

				switch n.Operator.Type {
				case lexer.PLUS:
					result = lInt + rInt
				case lexer.MINUS:
					result = lInt - rInt
				case lexer.MUL:
					result = lInt * rInt
				case lexer.DIV:
					if rInt == 0 {
						panic(verror.InterpreterVError{
							Position: verror.Position{
								Column: n.Operator.Column,
								Line:   n.Operator.Line,
							},
							Message: "Divide by zero",
						})
					}
					result = lInt / rInt
				}
			} else {
				switch n.Operator.Type {
				case lexer.PLUS:
					result = leftVal + rightVal
				case lexer.MINUS:
					result = leftVal - rightVal
				case lexer.MUL:
					result = leftVal * rightVal
				case lexer.DIV:
					if rightVal == 0 {
						panic(verror.InterpreterVError{
							Position: verror.Position{
								Column: n.Operator.Column,
								Line:   n.Operator.Line,
							},
							Message: "Divide by zero",
						})
					}
					result = leftVal / rightVal
				}
			}

			return result
		}
	case *ast.Literal:
		return n.Value
	}
	return nil
}

func (i *Interpreter) EvalSafe() any {
	defer func() {
		if r := recover(); r != nil {
			if parseErr, ok := r.(verror.InterpreterVError); ok {
				i.errors = append(i.errors, parseErr)
				fmt.Printf("Caught Error: %v\n", parseErr)
			} else {
				panic(r)
			}
		}
	}()

	return i.Eval(i.p.ParseProgram(), i.env)
}
