package ipt

import (
	"fmt"
	"reflect"
	"slices"
	"vine-lang/ast"
	environment "vine-lang/env"
	"vine-lang/object/store"
	"vine-lang/object/task"
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

func (i *Interpreter) EvalProgramStmt(program *ast.ProgramStmt, env *environment.Environment) (any, error) {
	var lastResult any
	var err error
	for _, s := range program.Body {
		if _, ok := s.(*ast.CommentStmt); !ok {
			lastResult, err = i.Eval(s, env)
		}
	}
	return lastResult, err
}

func (i *Interpreter) EvalBlockStmt(node *ast.BlockStmt, env *environment.Environment) (any, error) {
	var lastResult any
	var err error
	for _, s := range node.Body {
		if _, ok := s.(*ast.CommentStmt); !ok {
			lastResult, err = i.Eval(s, env)
		}
	}
	return lastResult, err
}

func (i *Interpreter) EvalUseDecl(n *ast.UseDecl, env *environment.Environment) (any, error) {
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
}

func (i *Interpreter) EvalVariableDecl(n *ast.VariableDecl, env *environment.Environment) (any, error) {
	val, err := i.Eval(n.Value, env)
	if n.IsConst {
		env.DefineConst(*n.Name.Value, val)
	} else {
		env.DefineFast(n.Name.Value.Value, val)
	}
	return val, err
}

func (i *Interpreter) EvalExposeStmt(n *ast.ExposeStmt, env *environment.Environment) (any, error) {
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
			if decl.Name.Value == nil {
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
}

func (i *Interpreter) EvalForStmt(n *ast.ForStmt, env *environment.Environment) (any, error) {
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
			valueOf := reflect.ValueOf(value)
			length := valueOf.Len()
			// 对于简单的循环体（不包含break/continue等），直接在loopEnv中执行
			// 只有在需要隔离作用域时才创建新环境
			simpleBody := isSimpleLoopBody(&n.Body)
			nameToken := *name.Value

			for index := 0; index < length; index++ {
				if simpleBody {
					// 简单循环体直接使用loopEnv，避免环境创建和释放
					// 使用SetFast方法，避免Lookup开销
					loopEnv.SetFast(nameToken.Value, valueOf.Index(index).Interface())
					_, err = i.Eval(&n.Body, loopEnv)
				} else {
					// 复杂循环体需要隔离作用域
					bodyEnv := environment.NewPooled(env.FileName)
					bodyEnv.Define(nameToken, valueOf.Index(index).Interface())
					bodyEnv.Link(loopEnv)
					_, err = i.Eval(&n.Body, bodyEnv)
					bodyEnv.Release()
				}

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

	// 判断是否需要隔离作用域
	simpleBody := isSimpleLoopBody(&n.Body)

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

		var err error
		var res any

		if simpleBody {
			// 简单循环体直接使用loopEnv
			res, err = i.Eval(&n.Body, loopEnv)
		} else {
			// 复杂循环体需要隔离作用域
			bodyEnv := environment.NewPooled(env.FileName)
			bodyEnv.Link(loopEnv)
			res, err = i.Eval(&n.Body, bodyEnv)
			bodyEnv.Release()
		}

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
}

// 判断循环体是否简单（不包含需要隔离作用域的语句）
func isSimpleLoopBody(body *ast.BlockStmt) bool {
	if body == nil {
		return true
	}
	for _, stmt := range body.Body {
		switch stmt.(type) {
		case *ast.VariableDecl:
			// 如果有变量声明，需要隔离作用域
			return false
		case *ast.ForStmt, *ast.IfStmt:
			// 嵌套控制结构通常需要隔离作用域
			return false
		case *ast.FunctionDecl, *ast.LambdaFunctionDecl:
			// 函数声明需要隔离作用域
			return false
		case *ast.SwitchStmt:
			// switch语句需要隔离作用域
			return false
		case *ast.TaskStmt:
			// task语句需要隔离作用域
			return false
		case *ast.CallTaskFn:
			// task调用需要隔离作用域
			return false
		case *ast.ToExpr:
			// to表达式需要隔离作用域
			return false
		}
		// 赋值语句、表达式语句、return语句等不需要隔离作用域
	}
	return true
}

func (i *Interpreter) EvalIfStmt(n *ast.IfStmt, env *environment.Environment) (any, error) {
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
}

func (i *Interpreter) EvalFunctionDecl(n *ast.FunctionDecl, env *environment.Environment) (any, error) {
	env.Define(*n.ID.Value, &types.FunctionLikeValNode{
		IsLamda:  false,
		IsModule: false,
		IsInside: false,
		Token:    n.ID.Value,
		Args:     n.Arguments,
		Body:     n.Body,
		IsTask:   false,
	})
	return nil, nil
}

func (i *Interpreter) EvalLambdaFunctionDecl(n *ast.LambdaFunctionDecl, env *environment.Environment) (any, error) {
	return &types.FunctionLikeValNode{
		IsLamda:  true,
		IsModule: false,
		IsInside: false,
		Token:    &token.Token{},
		Args:     &n.Args,
		Body:     &n.Body,
		IsTask:   false,
	}, nil
}

func (i *Interpreter) EvalSwitchStmt(n *ast.SwitchStmt, env *environment.Environment) (any, error) {
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
}

func (i *Interpreter) EvalTaskStmt(n *ast.TaskStmt, env *environment.Environment) (any, error) {
	env.Define(*n.Fn.ID.Value, &types.FunctionLikeValNode{
		IsLamda:  false,
		IsModule: false,
		IsInside: false,
		Token:    n.Fn.ID.Value,
		Args:     n.Fn.Arguments,
		Body:     n.Fn.Body,
		IsTask:   true,
	})
	return nil, nil
}

func (i *Interpreter) EvalWaitStmt(n *ast.WaitStmt, env *environment.Environment) (any, error) {
	target, err := i.Eval(n.Async, env)
	if err != nil {
		return nil, err
	}
	if taskObject, ok := target.(*task.TaskObject); ok {
		return taskObject.Wait(), nil
	}
	return nil, nil
}

func (i *Interpreter) EvalCallTaskFn(n *ast.CallTaskFn, env *environment.Environment) (any, error) {
	target, err := i.Eval(&n.Target, env)
	if err != nil {
		return nil, err
	}
	if taskFn, ok := target.(*task.TaskObject); ok {
		taskFn.Next(func(args ...[]any) any {
			parentTaskResult := taskFn.GetResult()
			currentToStmt := n.To
			newEnv := environment.New(env.WorkSpace)
			newEnv.Link(env)
			currentToStmtArgs := currentToStmt.Args.Arguments
			if len(currentToStmtArgs) > 0 {
				newEnv.Define(*currentToStmtArgs[0].(*ast.Literal).Value, parentTaskResult)
			}
			r, err := i.Eval(&n.To, newEnv)
			if err != nil {
				return err
			}
			TaskTo, ok := r.(*types.TaskToValNode)
			if !ok {
				return nil
			}

			var res any
			theTaskTo := TaskTo
			for {
				if theTaskTo.Env(nil) == nil {
					theTaskTo.Env(newEnv)
				}
				theTaskTo.Parnet = taskFn
				res = theTaskTo.Current()
				if theTaskTo.Next == nil {
					break
				}
				theTaskTo = theTaskTo.Next
			}
			return res
		}).Catch(func(catchErr any) any {
			currentCatchStmt := n.Catch
			if currentCatchStmt == nil {
				return nil
			}
			newEnv := environment.New(env.WorkSpace)
			newEnv.Link(env)
			currentCatchStmtArgs := currentCatchStmt.Args.Arguments
			catchErr, ok = catchErr.(verror.InterpreterVError)
			if !ok {
				return catchErr
			}
			// 包装错误值
			catchErrWrapper := types.CreateErrorValNode(catchErr)
			if len(currentCatchStmtArgs) > 0 {
				newEnv.Define(*currentCatchStmtArgs[0].(*ast.Literal).Value, catchErrWrapper)
			}
			r, err := i.Eval(n.Catch, newEnv)
			if err != nil {
				return err
			}
			TaskCatch, ok := r.(*types.FunctionLikeValNode)
			if !ok {
				return nil
			}
			// 执行catch 函数
			r, err = i.Eval(TaskCatch.Body, newEnv)
			if err != nil {
				return err
			}
			return r
		}).Run()
		return nil, nil
	}
	return nil, i.Errorf(token.Token{}, "not a task function")
}

func (i *Interpreter) EvalToExpr(n *ast.ToExpr, env *environment.Environment) (any, error) {
	var Next *types.TaskToValNode
	if n.Next != nil {
		// 将Expr转换为 TaskToValNode
		next, err := i.Eval(n.Next, env)
		if err != nil {
			return nil, err
		}
		Next = next.(*types.TaskToValNode)
	}

	// 创建一个新的 task To Node
	var toVal = types.TaskToValNode{
		Next: Next,
	}

	toVal.Current = func() any {
		_currentEnv := toVal.Env(nil)
		currentEnv := _currentEnv.(*environment.Environment)
		res, err := i.Eval(&n.Body, currentEnv)
		if err != nil {
			return nil
		}

		// 为next节点定义 result 参数
		if Next != nil {
			nextEnv := environment.New(env.WorkSpace)
			currentToStmt := n.Next
			nextEnv.Link(currentEnv)
			currentToStmtArgs := currentToStmt.Args.Arguments
			if len(currentToStmtArgs) > 0 {
				nextEnv.Define(*currentToStmtArgs[0].(*ast.Literal).Value, res)
			}
			Next.Env(nextEnv)
		}
		return res
	}
	return &toVal, nil
}

func (i *Interpreter) EvalAssignmentExpr(n *ast.AssignmentExpr, env *environment.Environment) (any, error) {
	operand, ok := n.Left.(*ast.Literal)
	if !ok {
		return nil, i.Errorf(n.Operator, "operand of assign must be a variable")
	}
	if operand.Value.Type != token.IDENT {
		return nil, i.Errorf(*operand.Value, fmt.Sprintf("cannot increment non-variable (type %s)", operand.Value.String()))
	}

	val, err := i.Eval(n.Right, env)
	if err != nil {
		return nil, err
	}
	env.SetFast(operand.Value.Value, val)
	return nil, nil
}

func (i *Interpreter) EvalCompareExpr(n *ast.CompareExpr, env *environment.Environment) (any, error) {
	leftRaw, err := i.Eval(n.Left, env)
	if err != nil {
		return nil, i.Errorf(n.Operator, err.Error())
	}
	rightRaw, err := i.Eval(n.Right, env)
	if err != nil {
		return nil, i.Errorf(n.Operator, err.Error())
	}

	// 快速路径处理常见的整数比较，避免类型解析开销
	if left, ok := leftRaw.(int64); ok {
		if right, ok := rightRaw.(int64); ok {
			switch n.Operator.Type {
			case token.EQ:
				return left == right, nil
			case token.NOT_EQ:
				return left != right, nil
			case token.LESS:
				return left < right, nil
			case token.LESS_EQ:
				return left <= right, nil
			case token.GREATER:
				return left > right, nil
			case token.GREATER_EQ:
				return left >= right, nil
			}
		}
	}

	// 快速路径处理常见的浮点数比较
	if left, ok := leftRaw.(float64); ok {
		if right, ok := rightRaw.(float64); ok {
			switch n.Operator.Type {
			case token.EQ:
				return left == right, nil
			case token.NOT_EQ:
				return left != right, nil
			case token.LESS:
				return left < right, nil
			case token.LESS_EQ:
				return left <= right, nil
			case token.GREATER:
				return left > right, nil
			case token.GREATER_EQ:
				return left >= right, nil
			}
		}
	}

	// 其他情况使用通用的CompareVal处理
	return utils.CompareVal(leftRaw, n.Operator.Type, rightRaw)
}

func (i *Interpreter) EvalBinaryExpr(n *ast.BinaryExpr, env *environment.Environment) (any, error) {
	leftRaw, err := i.Eval(n.Left, env)
	if err != nil {
		return nil, i.Errorf(n.Operator, err.Error())
	}
	rightRaw, err := i.Eval(n.Right, env)
	if err != nil {
		return nil, i.Errorf(n.Operator, err.Error())
	}

	// 快速路径处理常见的整数运算，避免类型解析开销
	if left, ok := leftRaw.(int64); ok {
		if right, ok := rightRaw.(int64); ok {
			switch n.Operator.Type {
			case token.PLUS:
				return left + right, nil
			case token.MINUS:
				return left - right, nil
			case token.MUL:
				return left * right, nil
			case token.DIV:
				if right == 0 {
					return nil, i.Errorf(n.Operator, "division by zero")
				}
				if left%right == 0 {
					return left / right, nil
				}
				return float64(left) / float64(right), nil
			}
		}
	}

	// 快速路径处理常见的浮点数运算
	if left, ok := leftRaw.(float64); ok {
		if right, ok := rightRaw.(float64); ok {
			switch n.Operator.Type {
			case token.PLUS:
				return left + right, nil
			case token.MINUS:
				return left - right, nil
			case token.MUL:
				return left * right, nil
			case token.DIV:
				if right == 0 {
					return nil, i.Errorf(n.Operator, "division by zero")
				}
				return left / right, nil
			}
		}
	}

	// 其他情况使用通用的BinaryVal处理
	result, err := utils.BinaryVal(leftRaw, n.Operator.Type, rightRaw)
	return result, err
}

func (i *Interpreter) EvalArrayExpr(n *ast.ArrayExpr, env *environment.Environment) (any, error) {
	var arr = make([]any, len(n.Items))
	for index, element := range n.Items {
		v, err := i.Eval(element.Value, env)
		if err != nil {
			return nil, err
		}
		arr[index] = v
	}
	return arr, nil
}

func (i *Interpreter) EvalObjectExpr(n *ast.ObjectExpr, env *environment.Environment) (any, error) {
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
}

func (i *Interpreter) EvalMemberExpr(n *ast.MemberExpr, env *environment.Environment) (any, error) {
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

	// 强制转换对象为 StoreObject
	m := store.NewStoreObjectWithGoStruct(obj)

	if v, ok := m.Get(prop.(token.Token)); ok {
		return v, nil
	}

	return nil, i.Errorf(token.Token{}, fmt.Sprintf("property %s not found", prop))
}

func (i *Interpreter) EvalArgsExpr(n *ast.ArgsExpr, env *environment.Environment) (any, error) {
	for index, arg := range n.Arguments {
		v, err := i.Eval(arg, env)
		if err != nil {
			return nil, err
		}
		n.Arguments[index] = v.(ast.Expr)
	}
	return n, nil
}

func (i *Interpreter) EvalCallExpr(n *ast.CallExpr, env *environment.Environment) (any, error) {
	function, _ := i.Eval(n.Callee, env)

	args := make([]any, len(n.Args.Arguments))

	for ind, arg := range n.Args.Arguments {
		args[ind], _ = i.Eval(arg, env)
	}

	if fn, ok := function.(token.Token); ok {
		return env.CallFunc(fn, args)
	} else if fn, ok := function.(*types.FunctionLikeValNode); ok {
		// 对于简单函数（不包含嵌套函数声明），使用池化的环境
		newEnv := environment.NewPooled(env.FileName)
		newEnv.Link(env) // 继承父环境

		for index, arg := range fn.Args.Arguments {
			name, ok := arg.(*ast.Literal)
			if !ok {
				newEnv.Release()
				return nil, i.Errorf(token.Token{}, "Not a valid variable to bind")
			}
			if len(args) <= index {
				newEnv.Release()
				return nil, i.Errorf(token.Token{}, "Not enough arguments")
			}
			newEnv.DefinePassing(*name.Value, args[index])
		}

		if fn.IsTask {
			tk := task.NewTaskObject(func(args ...[]any) any {
				res, err := i.Eval(fn.Body, newEnv)
				if err != nil {
					return err
				}
				return res
			})
			tk.Run()
			return tk, nil
		} else {
			res, err := i.Eval(fn.Body, newEnv)
			newEnv.Release() // 释放环境到池中
			if err != nil {
				return nil, err
			}
			return res, nil
		}
	} else if reflect.ValueOf(function).Kind() == reflect.Func {
		return env.CallFuncObject(function, args)
	} else {
		return nil, i.Errorf(token.Token{}, "Not a function")
	}
}

func (i *Interpreter) EvalUnaryExpr(n *ast.UnaryExpr, env *environment.Environment) (any, error) {
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

	env.SetFast(operand.Value.Value, newVal)

	if n.IsSuffix {
		return oldVal, nil
	} else {
		return newVal, nil
	}
}

func (i *Interpreter) EvalLiteral(n *ast.Literal, env *environment.Environment) (any, error) {
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
		if v, isGet := env.GetFast(n.Value.Value); isGet {
			if reflect.ValueOf(v).Kind() == reflect.Func {
				return *n.Value, nil
			}
			return v, nil
		}
		if v, isGet := env.Get(*n.Value); isGet {
			if reflect.ValueOf(v).Kind() == reflect.Func {
				return *n.Value, nil
			}
			return v, nil
		}
		ok := types.LibsKeywords(n.Value.Value)
		if ok.IsValidLibsKeyword() {
			return *n.Value, nil
		}

		return nil, i.Errorf(*n.Value, fmt.Sprintf("Unknown identifier: %s", n.Value.Value))
	}
	return nil, i.Errorf(*n.Value, fmt.Sprintf("Unknown literal type: %s", n.Value.Type))
}

func (i *Interpreter) Eval(node ast.Node, env *environment.Environment) (any, error) {

	switch node.NodeType() {
	case ast.NodeTypeProgramStmt:
		return i.EvalProgramStmt(node.(*ast.ProgramStmt), env)
	case ast.NodeTypeBlockStmt:
		return i.EvalBlockStmt(node.(*ast.BlockStmt), env)
	case ast.NodeTypeUseDecl:
		return i.EvalUseDecl(node.(*ast.UseDecl), env)
	case ast.NodeTypeExpressionStmt:
		return i.Eval(node.(*ast.ExpressionStmt).Expression, env)
	case ast.NodeTypeVariableDecl:
		return i.EvalVariableDecl(node.(*ast.VariableDecl), env)
	case ast.NodeTypeExposeStmt:
		return i.EvalExposeStmt(node.(*ast.ExposeStmt), env)
	case ast.NodeTypeForStmt:
		return i.EvalForStmt(node.(*ast.ForStmt), env)
	case ast.NodeTypeIfStmt:
		return i.EvalIfStmt(node.(*ast.IfStmt), env)
	case ast.NodeTypeFunctionDecl:
		return i.EvalFunctionDecl(node.(*ast.FunctionDecl), env)
	case ast.NodeTypeLambdaFunctionDecl:
		return i.EvalLambdaFunctionDecl(node.(*ast.LambdaFunctionDecl), env)
	case ast.NodeTypeReturnStmt:
		return i.Eval(node.(*ast.ReturnStmt).Value, env)
	case ast.NodeTypeSwitchStmt:
		return i.EvalSwitchStmt(node.(*ast.SwitchStmt), env)
	case ast.NodeTypeTaskStmt:
		return i.EvalTaskStmt(node.(*ast.TaskStmt), env)
	case ast.NodeTypeWaitStmt:
		return i.EvalWaitStmt(node.(*ast.WaitStmt), env)
	case ast.NodeTypeCallTaskFn:
		return i.EvalCallTaskFn(node.(*ast.CallTaskFn), env)
	case ast.NodeTypeToExpr:
		return i.EvalToExpr(node.(*ast.ToExpr), env)
	case ast.NodeTypeAssignmentExpr:
		return i.EvalAssignmentExpr(node.(*ast.AssignmentExpr), env)
	case ast.NodeTypeCompareExpr:
		return i.EvalCompareExpr(node.(*ast.CompareExpr), env)
	case ast.NodeTypeBinaryExpr:
		return i.EvalBinaryExpr(node.(*ast.BinaryExpr), env)
	case ast.NodeTypeProperty:
		return node, nil
	case ast.NodeTypeArrayExpr:
		return i.EvalArrayExpr(node.(*ast.ArrayExpr), env)
	case ast.NodeTypeObjectExpr:
		return i.EvalObjectExpr(node.(*ast.ObjectExpr), env)
	case ast.NodeTypeMemberExpr:
		return i.EvalMemberExpr(node.(*ast.MemberExpr), env)
	case ast.NodeTypeArgsExpr:
		return i.EvalArgsExpr(node.(*ast.ArgsExpr), env)
	case ast.NodeTypeCallExpr:
		return i.EvalCallExpr(node.(*ast.CallExpr), env)
	case ast.NodeTypeUnaryExpr:
		return i.EvalUnaryExpr(node.(*ast.UnaryExpr), env)
	case ast.NodeTypeLiteral:
		return i.EvalLiteral(node.(*ast.Literal), env)
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
		return nil, e
	}

	// 保证所有任务都执行完毕
	task.WaitAll()

	if i.env.Exports != nil {
		return types.NewUserModule(i.env.FileName, i.env.Exports), nil
	}
	return v, e
}
