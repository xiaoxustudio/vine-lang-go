package ipt

import (
	"fmt"
	"reflect"
	"slices"
	"vine-lang/ast"
	environment "vine-lang/env"
	"vine-lang/object/store"
	"vine-lang/parser"
	"vine-lang/token"
	"vine-lang/types"
	"vine-lang/utils"
	"vine-lang/verror"
)

type Interpreter struct {
	errors []verror.InterpreterVError
	p      *parser.Parser
	env    *environment.Environment
}

func New(p *parser.Parser, env *environment.Environment) *Interpreter {
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

func (i *Interpreter) Eval(node ast.Node, env *environment.Environment) (any, error) {
	if node == nil || env == nil {
		return nil, nil
	}

	switch n := node.(type) {
	case *ast.ProgramStmt:
		var lastResult any
		var err error
		for _, s := range n.Body {
			if _, ok := s.(*ast.CommentStmt); !ok {
				lastResult, err = i.Eval(s, env)
			}
		}
		return lastResult, err
	case *ast.BlockStmt:
		var lastResult any
		var err error
		for _, s := range n.Body {
			if _, ok := s.(*ast.CommentStmt); !ok {
				lastResult, err = i.Eval(s, env)
			}
		}
		return lastResult, err
	case *ast.UseDecl:
		var s token.Token
		if n.Source != nil && n.Source.Value != nil && (n.Source.Value.Type == token.IDENT || n.Source.Value.Type == token.STRING) {
			s = *n.Source.Value
		} else {
			return nil, i.Errorf(token.Token{}, "Invalid module name")
		}

		_, err := env.ImportModule(s.Value)
		if err != nil {
			return nil, err
		}

		if n.Mode == token.AS {
			if len(n.Specifiers) != 1 {
				return nil, i.Errorf(token.Token{}, "use as requires exactly one alias")
			}
			if aliasLit, ok := n.Specifiers[0].(*ast.Literal); ok {
				if aliasLit.Value.Type != token.IDENT {
					return nil, i.Errorf(*aliasLit.Value, "alias must be an identifier")
				}
				if mod, exists := env.Get(token.Token{Type: token.IDENT, Value: s.Value}); exists {
					env.Define(*aliasLit.Value, mod)
				} else {
					return nil, i.Errorf(token.Token{Type: token.IDENT, Value: s.Value}, "module not found after import")
				}
			} else {
				return nil, i.Errorf(token.Token{}, "invalid alias specifier")
			}
		} else if n.Mode == token.PICK {
			modAny, exists := env.Get(token.Token{Type: token.IDENT, Value: s.Value})
			if !exists {
				return nil, i.Errorf(token.Token{Type: token.IDENT, Value: s.Value}, "module not found after import")
			}
			if mod, ok := modAny.(types.LibsModule); ok {
				for _, sp := range n.Specifiers {
					if lit, ok := sp.(*ast.Literal); ok {
						if lit.Value.Type != token.IDENT {
							return nil, i.Errorf(*lit.Value, "pick target must be an identifier")
						}
						if fn, ok := mod.Get(*lit.Value); ok {
							env.Define(*lit.Value, fn)
						} else {
							return nil, i.Errorf(*lit.Value, fmt.Sprintf("function %s not found in module %s", lit.Value.Value, s.Value))
						}
						continue
					}
					if us, ok := sp.(*ast.UseSpecifier); ok {
						if us.Remote == nil || us.Remote.Value == nil || us.Remote.Value.Type != token.IDENT {
							return nil, i.Errorf(token.Token{}, "invalid pick specifier")
						}
						var local token.Token
						if us.Local != nil && us.Local.Value != nil {
							if us.Local.Value.Type != token.IDENT {
								return nil, i.Errorf(*us.Local.Value, "alias must be an identifier")
							}
							local = *us.Local.Value
						} else {
							local = *us.Remote.Value
						}
						if fn, ok := mod.Get(*us.Remote.Value); ok {
							env.Define(local, fn)
						} else {
							return nil, i.Errorf(*us.Remote.Value, fmt.Sprintf("function %s not found in module %s", us.Remote.Value.Value, s.Value))
						}
						continue
					}
					return nil, i.Errorf(token.Token{}, "invalid pick specifier")
				}
			} else {
				return nil, i.Errorf(token.Token{}, "invalid module type for pick")
			}
		} else if n.Mode == token.USE {
			if n.Source != nil && n.Source.Value != nil && n.Source.Value.Type == token.STRING {
				modAny, exists := env.Get(token.Token{Type: token.IDENT, Value: s.Value})
				if !exists {
					return nil, i.Errorf(token.Token{Type: token.IDENT, Value: s.Value}, "module not found after import")
				}
				if mod, ok := modAny.(types.LibsModule); ok {
					mod.ForEach(func(tk token.Token, val any) {
						env.Define(tk, val)
					})
				} else {
					return nil, i.Errorf(token.Token{}, "invalid module type for use")
				}
			}
		}
		return s, nil
	case *ast.ExpressionStmt:
		return i.Eval(n.Expression, env)
	case *ast.VariableDecl:
		val, err := i.Eval(n.Value, env)
		if n.IsConst {
			env.DefineConst(*n.Name.Value, val)
		} else {
			env.Define(*n.Name.Value, val)
		}
		return val, err
	case *ast.ExposeStmt:
		if env.Exports == nil {
			env.Exports = store.NewStoreObject()
		}

		if n.Decl != nil {
			if _, err := i.Eval(n.Decl, env); err != nil {
				return nil, err
			}

			switch decl := n.Decl.(type) {
			case *ast.FunctionDecl:
				if decl.ID == nil || decl.ID.Value == nil {
					return nil, i.Errorf(token.Token{}, "invalid expose function")
				}
				val, exists := env.Get(*decl.ID.Value)
				if !exists {
					return nil, i.Errorf(*decl.ID.Value, "expose target not found")
				}
				if err := env.Exports.Define(*decl.ID.Value, val); err != nil {
					return nil, err
				}
				return val, nil
			case *ast.VariableDecl:
				if decl.Name == nil || decl.Name.Value == nil {
					return nil, i.Errorf(token.Token{}, "invalid expose variable")
				}
				val, exists := env.Get(*decl.Name.Value)
				if !exists {
					return nil, i.Errorf(*decl.Name.Value, "expose target not found")
				}
				if err := env.Exports.Define(*decl.Name.Value, val); err != nil {
					return nil, err
				}
				return val, nil
			default:
				return nil, i.Errorf(token.Token{}, "invalid expose declaration")
			}
		}

		if n.Name != nil && n.Name.Value != nil {
			if n.Value != nil {
				val, err := i.Eval(n.Value, env)
				if err != nil {
					return nil, err
				}
				env.Define(*n.Name.Value, val)
			}

			val, exists := env.Get(*n.Name.Value)
			if !exists {
				return nil, i.Errorf(*n.Name.Value, "expose target not found")
			}
			if err := env.Exports.Define(*n.Name.Value, val); err != nil {
				return nil, err
			}
			return val, nil
		}

		return nil, i.Errorf(token.Token{}, "invalid expose statement")
	case *ast.ForStmt:
		loopEnv := environment.New(env.WorkSpace)
		loopEnv.Link(env)
		if n.Range != nil && n.Init != nil {
			name, ok := n.Init.(*ast.Literal)

			if !ok {
				return nil, i.Errorf(token.Token{}, "init must be a literal")
			}

			value, err := i.Eval(n.Range, loopEnv)
			if err != nil {
				return nil, err
			}

			if reflect.TypeOf(value).Kind() == reflect.Slice || reflect.TypeOf(value).Kind() == reflect.Array {
				for index := 0; index < reflect.ValueOf(value).Len(); index++ {
					bodyEnv := environment.NewPooled(env.FileName)
					bodyEnv.Define(*name.Value, reflect.ValueOf(value).Index(index).Interface())
					bodyEnv.Link(loopEnv)

					_, err := i.Eval(n.Body, bodyEnv)

					bodyEnv.Release()

					if err != nil {
						return nil, err
					}

				}
			}

			return nil, nil
		}

		if _, err := i.Eval(n.Init, loopEnv); err != nil {
			return nil, err
		}
		var result any

		for {
			if n.Value != nil {
				condVal, err := i.Eval(n.Value, loopEnv)
				if err != nil {
					return nil, err
				}

				if isTrue, ok := condVal.(bool); !ok || !isTrue {
					break
				}
			}

			bodyEnv := environment.NewPooled(env.FileName)
			bodyEnv.Link(loopEnv)

			res, err := i.Eval(n.Body, bodyEnv)

			bodyEnv.Release()

			if err != nil {
				return nil, err
			}

			if n.Update != nil {
				if _, err := i.Eval(n.Update, loopEnv); err != nil {
					return nil, err
				}
			}
			result = res
		}
		return result, nil
	case *ast.IfStmt:
		condVal, err := i.Eval(n.Test, env)
		if err != nil {
			return nil, err
		}

		if isTrue, ok := condVal.(bool); !ok || !isTrue {
			if n.Alternate != nil {
				return i.Eval(n.Alternate, env)
			}
			return nil, nil
		}

		return i.Eval(n.Consequent, env)
	case *ast.FunctionDecl:
		env.Define(*n.ID.Value, &types.FunctionLikeValNode{
			IsLamda:  false,
			IsModule: false,
			IsInside: false,
			Token:    n.ID.Value,
			Args:     n.Arguments,
			Body:     n.Body,
		})
		return nil, nil
	case *ast.SwitchStmt:
		condVal, err := i.Eval(n.Test, env)
		if err != nil {
			return nil, err
		}

		for _, c := range n.Cases {
			caseValue, ok := c.(*ast.SwitchCase)
			if !ok {
				return nil, i.Errorf(token.Token{}, "invalid switch case")
			}
			if caseValue != nil {
				for _, test := range caseValue.Conds {
					testVal, err := i.Eval(test, env)
					if err != nil {
						return nil, err
					}
					if testVal == nil {
						if caseValue.Body != nil {
							return i.Eval(caseValue.Body, env)
						}
					}
					ok, err := utils.CompareVal(testVal, token.EQ, condVal)
					if err != nil {
						return nil, err
					}
					if ok {
						return i.Eval(caseValue.Body, env)
					}
				}
			}
		}

		return nil, nil
	case *ast.AssignmentExpr:
		var err error
		operand, ok := n.Left.(*ast.Literal)
		if !ok {
			return nil, i.Errorf(n.Operator, "operand of assign must be a variable")
		}
		if operand.Value.Type != token.IDENT {
			return nil, i.Errorf(*operand.Value, fmt.Sprintf("cannot increment non-variable (type %s)", operand.Value.String()))
		}

		val, err := i.Eval(n.Right, env)
		env.Set(*operand.Value, val)
		return nil, err
	case *ast.CompareExpr:
		leftRaw, err := i.Eval(n.Left, env)
		if err != nil {
			return nil, i.Errorf(n.Operator, err.Error())
		}
		rightRaw, err := i.Eval(n.Right, env)
		if err != nil {
			return nil, i.Errorf(n.Operator, err.Error())
		}

		return utils.CompareVal(leftRaw, n.Operator.Type, rightRaw)
	case *ast.BinaryExpr:
		{
			leftRaw, err := i.Eval(n.Left, env)
			if err != nil {
				return nil, i.Errorf(n.Operator, err.Error())
			}
			rightRaw, err := i.Eval(n.Right, env)
			if err != nil {
				return nil, i.Errorf(n.Operator, err.Error())
			}

			result, err := utils.BinaryVal(leftRaw, n.Operator.Type, rightRaw)
			return result, err
		}
	case *ast.Property:
		return n, nil
	case *ast.ArrayExpr:
		var arr = make([]any, len(n.Items))
		for index, element := range n.Items {
			v, err := i.Eval(element.Value, env)
			if err != nil {
				return nil, err
			}
			arr[index] = v
		}
		return arr, nil
	case *ast.ObjectExpr:
		obj := store.NewStoreObject()
		obj.Define(token.Token{Type: token.IDENT, Value: "__proto__"}, store.NewStoreObject())
		for _, prop := range n.Properties {
			v, err := i.Eval(prop.Value, env)
			if err != nil {
				return nil, err
			}
			// 字符串和数字才可以当作key
			if slices.Contains([]token.TokenType{token.STRING, token.INT, token.FLOAT, token.IDENT}, prop.Key.Value.Type) {
				obj.Define(*prop.Key.Value, v)
			} else {
				return nil, i.Errorf(token.Token{}, "invalid key type")
			}
		}
		return obj, nil
	case *ast.MemberExpr:
		obj, err := i.Eval(n.Object, env)
		if obj == nil {
			return nil, i.Errorf(token.Token{}, "nil object")
		}
		if err != nil {
			return nil, err
		}

		var prop any
		if n.Computed {
			val, err := i.Eval(n.Property, env)
			if err != nil {
				return nil, err
			}

			switch v := val.(type) {
			case *token.Token:
				prop = *v
			case token.Token:
				prop = v
			default:
				prop = v
			}
		} else {
			if ident, ok := n.Property.(*ast.Literal); ok {
				prop = *ident.Value
			} else {
				if scope, ok := obj.(types.Scope); ok {
					env.MountScope = scope
					val, err := i.Eval(n.Property, env)
					env.MountScope = nil
					if err != nil {
						return nil, err
					}
					return val, nil
				} else {
					return nil, i.Errorf(token.Token{}, "invalid property structure")
				}
			}
		}

		/* 对象 */
		if m, ok := obj.(*store.StoreObject); ok {
			if p, ok := prop.(token.Token); ok {
				if v, ok := m.Get(p); ok {
					return v, nil
				}
			}
		}

		/* 模块 */
		if m, ok := obj.(types.LibsModule); ok {
			if v, ok := m.Get(prop.(token.Token)); ok {
				return v, nil
			}
		}

		/* Slice or Array */
		if m, ok := obj.([]any); ok {
			switch v := prop.(type) {
			case int64:
				if prop.(int64) >= int64(len(m)) {
					return nil, i.Errorf(token.Token{}, "index out of range")
				}
				return m[prop.(int64)], nil
			case token.Token:
				if v.Type != token.INT {
					return nil, i.Errorf(token.Token{}, "index must be an integer")
				}
				_i := prop.(token.Token)
				if _i.Type == token.INT {
					index, err := _i.GetInt()
					if err != nil {
						return nil, i.Errorf(token.Token{}, "index must be an integer")
					}
					if index >= int64(len(m)) {
						return nil, i.Errorf(token.Token{}, "index out of range")
					}
					return m[index], nil
				} else {
					return nil, i.Errorf(token.Token{}, "index must be an integer")
				}
			}

		}

		return nil, i.Errorf(token.Token{}, fmt.Sprintf("property %s not found", prop))
	case *ast.ArgsExpr:
		for index, arg := range n.Arguments {
			v, err := i.Eval(arg, env)
			if err != nil {
				return nil, err
			}
			n.Arguments[index] = v.(ast.Expr)
		}
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
			} else if reflect.ValueOf(function).Kind() == reflect.Func {
				return env.CallFuncObject(function, args)
			} else if fn, ok := function.(*types.FunctionLikeValNode); ok {
				newEnv := environment.New(env.WorkSpace)
				newEnv.Link(env)

				for index, arg := range fn.Args.Arguments {
					name, ok := arg.(*ast.Literal)
					if !ok {
						return nil, i.Errorf(token.Token{}, "Not a valid variable to bind")
					}
					if len(args) <= index {
						return nil, i.Errorf(token.Token{}, "Not enough arguments")
					}
					env.Define(*name.Value, args[index])
				}

				res, err := i.Eval(fn.Body, newEnv)
				if err != nil {
					return nil, err
				}
				return res, nil
			} else {
				return nil, i.Errorf(token.Token{}, "Not a function")
			}
		}
	case *ast.UnaryExpr:
		if n.Operator.Type == token.MINUS || n.Operator.Type == token.NOT {
			val, err := i.Eval(n.Value, env)
			if err != nil {
				return nil, err
			}

			if n.Operator.Type == token.MINUS {
				switch v := val.(type) {
				case int64:
					return -v, nil
				case float64:
					return -v, nil
				case int:
					return -v, nil
				default:
					return nil, i.Errorf(n.Operator, fmt.Sprintf("invalid operation: - (non-numeric type %T)", v))
				}
			} else {
				if b, ok := val.(bool); ok {
					return !b, nil
				}
				return nil, i.Errorf(n.Operator, fmt.Sprintf("invalid operation: ! (non-boolean type %T)", val))
			}
		}

		operand, ok := n.Value.(*ast.Literal)
		if !ok {
			return nil, i.Errorf(n.Operator, "operand of increment/decrement must be a variable")
		}

		if operand.Value.Type != token.IDENT {
			return nil, i.Errorf(*operand.Value, fmt.Sprintf("cannot increment non-variable (type %s)", operand.Value.String()))
		}

		oldVal, exists := env.Get(*operand.Value)
		if !exists {
			return nil, i.Errorf(*operand.Value, fmt.Sprintf("undefined variable: %s", operand.Value.Value))
		}

		var newVal any

		switch v := oldVal.(type) {
		case int64:
			if n.Operator.Type == token.INC {
				newVal = v + 1
			} else {
				newVal = v - 1
			}
		case float64:
			if n.Operator.Type == token.INC {
				newVal = v + 1
			} else {
				newVal = v - 1
			}
		case int:
			if n.Operator.Type == token.INC {
				newVal = v + 1
			} else {
				newVal = v - 1
			}
		default:
			return nil, i.Errorf(*operand.Value, fmt.Sprintf("invalid operation: %s (non-numeric type %T)", n.Operator.Value, v))
		}

		env.Set(*operand.Value, newVal)

		if n.IsSuffix {
			return oldVal, nil
		} else {
			return newVal, nil
		}
	case *ast.Literal:
		switch n.Value.Type {
		case token.INT:
			num, err := n.Value.GetInt()
			if err != nil {
				return nil, err
			}
			return num, nil
		case token.FLOAT:
			num, err := n.Value.GetFloat()
			if err != nil {
				return nil, err
			}
			return num, nil
		case token.STRING:
			return n.Value.Value, nil
		case token.TRUE:
			return true, nil
		case token.FALSE:
			return false, nil
		case token.NIL:
			return nil, nil
		case token.IDENT:
			ok := types.LibsKeywords(n.Value.Value)
			if v, isGet := env.Get(*n.Value); isGet {
				if reflect.ValueOf(v).Kind() == reflect.Func {
					return *n.Value, nil
				}
				return v, nil
			}
			if ok.IsValidLibsKeyword() {
				return *n.Value, nil
			}
			i.Errorf(*n.Value, fmt.Sprintf("Unknown identifier: %s", n.Value.Value))
		}
		i.Errorf(*n.Value, fmt.Sprintf("Unknown literal type: %s", n.Value.Type))
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
	// ast.Print()
	v, e := i.Eval(ast, i.env)
	if e != nil {
		panic(e)
	}
	if i.env.Exports != nil {
		return types.NewUserModule(i.env.FileName, i.env.Exports), nil
	}
	return v, e
}
