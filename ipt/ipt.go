package ipt

import (
	"fmt"
	"reflect"
	"vine-lang/ast"
	"vine-lang/env"
	"vine-lang/parser"
	"vine-lang/token"
	"vine-lang/utils"
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

func (i *Interpreter) Errorf(tk token.Token, format string) verror.InterpreterVError {
	return verror.InterpreterVError{
		Message:  format,
		Position: tk.ToPosition(i.env.FileName),
	}
}

func (i *Interpreter) Eval(node ast.Node, env *env.Environment) (any, error) {

	switch n := node.(type) {
	case *ast.ProgramStmt:
		var lastResult any
		var err error
		for _, s := range n.Body {
			lastResult, err = i.Eval(s, env)
		}
		return lastResult, err
	case *ast.UseDecl:
		source, err := i.Eval(n.Source, env)
		if err != nil {
			return nil, err
		}
		s, ok := source.(token.Token)
		if !ok {
			return nil, i.Errorf(token.Token{}, "Invalid module name")
		}
		env.ImportModule(s.Value)
		for _, s := range n.Specifiers {
			i.Eval(s, env)
		}
		return source, err
	case *ast.ExpressionStmt:
		return i.Eval(n.Expression, env)
	case *ast.VariableDecl:
		val, err := i.Eval(n.Value, env)
		env.Define(n.Name.Value, val)
		return nil, err
	case *ast.AssignmentExpr:
		var err error
		name, err := i.Eval(n.Left, env)
		val, err := i.Eval(n.Right, env)
		nameKey, ok := name.(token.Token)
		if !ok {
			return nil, i.Errorf(*n.ID, "Invalid assignment target")
		}
		env.Set(nameKey, val)
		return nil, err
	case *ast.BinaryExpr:
		{
			var err error
			leftRaw, err := i.Eval(n.Left, env)
			rightRaw, err := i.Eval(n.Right, env)
			leftVal, isLeftInt, err := utils.GetNumberAndType(leftRaw)
			if err != nil {
				return nil, i.Errorf(n.Operator, err.Error())
			}

			/* 数字转换 */
			rightVal, isRightInt, err := utils.GetNumberAndType(rightRaw)
			if err != nil {
				return nil, i.Errorf(n.Operator, err.Error())
			}

			var result any

			if isLeftInt && isRightInt {
				lInt := int64(leftVal)
				rInt := int64(rightVal)

				switch n.Operator.Type {
				case token.PLUS:
					result = lInt + rInt
				case token.MINUS:
					result = lInt - rInt
				case token.MUL:
					result = lInt * rInt
				case token.DIV:
					if rInt == 0 {
						return nil, i.Errorf(n.Operator, "Divide by zero")
					}
					result = lInt / rInt
				}
			} else {
				switch n.Operator.Type {
				case token.PLUS:
					result = leftVal + rightVal
				case token.MINUS:
					result = leftVal - rightVal
				case token.MUL:
					result = leftVal * rightVal
				case token.DIV:
					if rightVal == 0 {
						return nil, i.Errorf(n.Operator, "Divide by zero")
					}
					result = leftVal / rightVal
				}
			}

			return result, err
		}
	case *ast.ArgsExpr:
		return n, nil
	case *ast.CallExpr:
		{
			var err error
			function, err := i.Eval(n.Callee, env)
			args := make([]any, len(n.Args.Arguments))
			for ind, arg := range n.Args.Arguments {
				args[ind], err = i.Eval(arg, env)
			}
			if fn, ok := function.(token.Token); ok {
				env.CallFunc(fn, args)
			} else {
				return nil, i.Errorf(*n.ID, "Not a function")
			}
			return nil, err
		}
	case *ast.Literal:
		if v, isGet := env.Get(n.Value); isGet {
			if reflect.ValueOf(v).Kind() == reflect.Func {
				// 内置函数
				return n.Value, nil
			}
			return v, nil
		} else {
			return n.Value, nil
		}
	}
	return nil, i.Errorf(token.Token{}, "Unknown node type")
}

func (i *Interpreter) EvalSafe() (any, error) {
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
