package handlers

import (
	"fmt"
	"vine-lang/ast"
	iface "vine-lang/compiler/interface"
	"vine-lang/env"
	"vine-lang/object/store"
	"vine-lang/token"
	"vine-lang/types"
)

type moduleAccessor interface {
	Get(tok token.Token) (any, bool)
}

func pickSpecifiers(env *env.Environment, moduleName string, specifiers []ast.Specifier, m moduleAccessor, getFn func(m moduleAccessor, fnName string) (any, bool)) error {
	for _, sp := range specifiers {
		var fnName, localName string
		if lit, ok := sp.(*ast.Literal); ok {
			if lit.Value.Type != token.IDENT {
				return fmt.Errorf("pick target must be an identifier")
			}
			fnName = lit.Value.Value
			localName = fnName
		} else if us, ok := sp.(*ast.UseSpecifier); ok {
			if us.Remote == nil || us.Remote.Value == nil || us.Remote.Value.Type != token.IDENT {
				return fmt.Errorf("invalid pick specifier")
			}
			fnName = us.Remote.Value.Value
			if us.Local != nil && us.Local.Value != nil {
				if us.Local.Value.Type != token.IDENT {
					return fmt.Errorf("alias must be an identifier")
				}
				localName = us.Local.Value.Value
			} else {
				localName = fnName
			}
		} else {
			return fmt.Errorf("invalid pick specifier")
		}
		if val, ok := getFn(m, fnName); ok {
			env.DefineFast(localName, val)
		} else {
			return fmt.Errorf("function %s not found in module %s", fnName, moduleName)
		}
	}
	return nil
}

func pickFromTokenGetter(m moduleAccessor, fnName string) (any, bool) {
	return m.Get(token.Token{Type: token.IDENT, Value: fnName})
}

func HandleUseDeclStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.UseDecl)

	moduleName := n.Source.Value.Value

	currentScope := c.CurrentScope()
	env := currentScope.Env

	mod, err := env.ImportModule(moduleName)
	if err != nil {
		return nil, err
	}

	switch n.Mode {
	case token.USE:
		env.DefineFast(moduleName, mod)

	case token.AS:
		if len(n.Specifiers) != 1 {
			return nil, fmt.Errorf("use as requires exactly one alias")
		}
		var alias string
		if aliasLit, ok := n.Specifiers[0].(*ast.Literal); ok {
			if aliasLit.Value.Type != token.IDENT {
				return nil, fmt.Errorf("alias must be an identifier")
			}
			alias = aliasLit.Value.Value
		} else {
			return nil, fmt.Errorf("invalid alias specifier")
		}
		env.DefineFast(alias, mod)

	case token.PICK:
		switch m := mod.(type) {
		case *types.UserModule:
			if err := pickSpecifiers(&env, moduleName, n.Specifiers, m, pickFromTokenGetter); err != nil {
				return nil, err
			}
		case types.LibsModule:
			if err := pickSpecifiers(&env, moduleName, n.Specifiers, m, pickFromTokenGetter); err != nil {
				return nil, err
			}
		case *store.StoreObject:
			if err := pickSpecifiers(&env, moduleName, n.Specifiers, m, pickFromTokenGetter); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("invalid module type for pick: %T", mod)
		}

	default:
		return nil, fmt.Errorf("unknown use mode: %s", n.Mode)
	}

	return nil, nil
}
