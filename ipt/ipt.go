package ipt

import (
	"fmt"
	"reflect"
	"vine-lang/ast"
	"vine-lang/env"
	environment "vine-lang/env"
	"vine-lang/parser"
	"vine-lang/token"
	"vine-lang/types"
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
	panic(verror.InterpreterVError{
		Message:  format,
		Position: tk.ToPosition(i.env.FileName),
	})
}

func (i *Interpreter) Eval(node ast.Node, env *env.Environment) (any, error) {
	if node == nil || env == nil {
		return nil, nil
	}

	switch n := node.(type) {
	case *ast.ProgramStmt:
		var lastResult any
		var err error
		for _, s := range n.Body {
			lastResult, err = i.Eval(s, env)
		}
		return lastResult, err
	case *ast.BlockStmt:
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
		return val, err
	case *ast.ForStmt:
		var newEnv = environment.New(env.FileName)
		newEnv.Link(env)
		if n.Range != nil {
			return nil, nil
		} else {
			i.Eval(n.Init, newEnv)
			for {
				if res, err := i.Eval(n.Value, newEnv); err == nil {
					if val, ok := res.(bool); ok && val {
						i.Eval(n.Body, newEnv)
						i.Eval(n.Update, newEnv)
					} else {
						break
					}
				} else {
					break
				}
			}
			return nil, nil
		}
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
	case *ast.CompareExpr:
		{
			leftRaw, err := i.Eval(n.Left, env)
			rightRaw, err := i.Eval(n.Right, env)
			if err != nil {
				return nil, i.Errorf(n.Operator, err.Error())
			}
			return utils.CompareVal(leftRaw, n.Operator.Type, rightRaw)
		}
	case *ast.BinaryExpr:
		{
			var err error
			leftRaw, err := i.Eval(n.Left, env)
			rightRaw, err := i.Eval(n.Right, env)
			leftVal, leftEnum, err := utils.GetNumberAndType(leftRaw)
			isLeftInt := leftEnum == types.GNT_INT
			if err != nil {
				return nil, i.Errorf(n.Operator, err.Error())
			}

			/* 数字转换 */
			rightVal, rightEnum, err := utils.GetNumberAndType(rightRaw)
			isRightInt := rightEnum == types.GNT_INT
			if err != nil {
				return nil, i.Errorf(n.Operator, err.Error())
			}

			var result any

			if isLeftInt && isRightInt {
				lInt := leftVal.(int64)
				rInt := rightVal.(int64)

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
					if isLeftInt {
						if rightEnum == types.GNT_INT {
							result = leftVal.(int64) + rightVal.(int64)
						} else {
							intVal, ok := leftVal.(int64)
							if !ok {
								return nil, i.Errorf(n.Operator, "Invalid operand type")
							}
							result = float64(intVal) + rightVal.(float64)
						}
					} else if isRightInt {
						intVal, ok := rightVal.(int64)
						if !ok {
							return nil, i.Errorf(n.Operator, "Invalid operand type")
						}
						result = leftVal.(float64) + float64(intVal)
					} else {
						result = leftVal.(float64) + rightVal.(float64)
					}
				case token.MINUS:
					if isLeftInt {
						if rightEnum == types.GNT_INT {
							result = leftVal.(int64) - rightVal.(int64)
						} else {
							intVal, ok := leftVal.(int64)
							if !ok {
								return nil, i.Errorf(n.Operator, "Invalid operand type")
							}
							result = float64(intVal) - rightVal.(float64)
						}
					} else if isRightInt {
						intVal, ok := rightVal.(int64)
						if !ok {
							return nil, i.Errorf(n.Operator, "Invalid operand type")
						}
						result = leftVal.(float64) - float64(intVal)
					} else {
						result = leftVal.(float64) - rightVal.(float64)
					}
				case token.MUL:
					if isLeftInt {
						if rightEnum == types.GNT_INT {
							result = leftVal.(int64) * rightVal.(int64)
						} else {
							intVal, ok := leftVal.(int64)
							if !ok {
								return nil, i.Errorf(n.Operator, "Invalid operand type")
							}
							result = float64(intVal) * rightVal.(float64)
						}
					} else if isRightInt {
						intVal, ok := rightVal.(int64)
						if !ok {
							return nil, i.Errorf(n.Operator, "Invalid operand type")
						}
						result = leftVal.(float64) * float64(intVal)
					} else {
						result = leftVal.(float64) * rightVal.(float64)
					}
				case token.DIV:
					if rightVal == 0 {
						return nil, i.Errorf(n.Operator, "Divide by zero")
					}
					if isLeftInt {
						if rightEnum == types.GNT_INT {
							result = leftVal.(int64) / rightVal.(int64)
						} else {
							intVal, ok := leftVal.(int64)
							if !ok {
								return nil, i.Errorf(n.Operator, "Invalid operand type")
							}
							result = float64(intVal) / rightVal.(float64)
						}
					} else if isRightInt {
						intVal, ok := rightVal.(int64)
						if !ok {
							return nil, i.Errorf(n.Operator, "Invalid operand type")
						}
						result = leftVal.(float64) / float64(intVal)
					} else {
						result = leftVal.(float64) / rightVal.(float64)
					}
				}
			}

			return result, err
		}
	case *ast.ArgsExpr:
		return n, nil
	case *ast.CallExpr:
		{
			function, _ := i.Eval(n.Callee, env)
			args := make([]any, len(n.Args.Arguments))
			for ind, arg := range n.Args.Arguments {
				args[ind], _ = i.Eval(arg, env)
			}
			if fn, ok := function.(token.Token); ok {
				v, e := env.CallFunc(fn, args)
				return v, e
			} else {
				return nil, i.Errorf(*n.ID, "Not a function")
			}
		}
	case *ast.UnaryExpr:
		operand, ok := n.Value.(*ast.Literal)
		if !ok {
			return nil, i.Errorf(n.Operator, "operand of increment/decrement must be a variable")
		}

		if operand.Value.Type != token.IDENT {
			return nil, i.Errorf(operand.Value, fmt.Sprintf("cannot increment non-variable (type %s)", operand.Value.String()))
		}

		oldVal, exists := env.Get(operand.Value)
		if !exists {
			return nil, i.Errorf(operand.Value, fmt.Sprintf("undefined variable: %s", operand.Value.Value))
		}

		var newVal interface{}

		switch v := oldVal.(type) {
		case int:
			newVal = v + 1
		case int64:
			newVal = v + 1
		case float64:
			newVal = v + 1
		default:
			return nil, i.Errorf(operand.Value, fmt.Sprintf("invalid operation: ++ (non-numeric type %T)", v))
		}

		env.Set(operand.Value, newVal)

		if n.IsSuffix {
			return oldVal, nil
		} else {
			return newVal, nil
		}

	case *ast.Literal:
		if n.Value.Type == token.IDENT {
			ok := types.LibsKeywords(n.Value.Value)
			if ok.IsValidLibsKeyword() {
				return n.Value, nil
			}
			if v, isGet := env.Get(n.Value); isGet {
				if reflect.ValueOf(v).Kind() == reflect.Func {
					// 内置函数
					return n.Value, nil
				}
				return v, nil
			}
		}
		switch n.Value.Type {
		case token.NUMBER:
			v, _, err := utils.GetNumberAndType(n.Value)
			return v, err
		case token.STRING:
			return n.Value.Value, nil
		case token.TRUE:
			return true, nil
		case token.FALSE:
			return false, nil
		case token.NIL:
			return nil, nil
		}

		i.Errorf(n.Value, fmt.Sprintf("Unknown identifier: %s", n.Value))
	case *ast.CommentStmt:
		return nil, nil
	}
	return nil, i.Errorf(token.Token{}, fmt.Sprintf("Unknown AST node type: %T", node))
}

func (i *Interpreter) EvalSafe() (any, error) {
	defer func() {
		if r := recover(); r != nil {
			if parseErr, ok := r.(verror.InterpreterVError); ok {
				i.errors = append(i.errors, parseErr)
				fmt.Printf("Runtime Error: %v\n", parseErr)
			} else {
				panic(r)
			}
		}
	}()
	ast := i.p.ParseProgram()
	v, e := i.Eval(ast, i.env)
	if e != nil {
		panic(e)
	}
	return v, e
}
