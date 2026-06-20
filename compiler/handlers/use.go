package handlers

import (
	"fmt"
	"vine-lang/ast"
	iface "vine-lang/compiler/interface"
	"vine-lang/object/store"
	"vine-lang/token"
	"vine-lang/types"
)

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
			for _, sp := range n.Specifiers {
				var fnName, localName string

				if lit, ok := sp.(*ast.Literal); ok {
					if lit.Value.Type != token.IDENT {
						return nil, fmt.Errorf("pick target must be an identifier")
					}
					fnName = lit.Value.Value
					localName = fnName
				} else if us, ok := sp.(*ast.UseSpecifier); ok {
					if us.Remote == nil || us.Remote.Value == nil || us.Remote.Value.Type != token.IDENT {
						return nil, fmt.Errorf("invalid pick specifier")
					}
					fnName = us.Remote.Value.Value
					if us.Local != nil && us.Local.Value != nil {
						if us.Local.Value.Type != token.IDENT {
							return nil, fmt.Errorf("alias must be an identifier")
						}
						localName = us.Local.Value.Value
					} else {
						localName = fnName
					}
				} else {
					return nil, fmt.Errorf("invalid pick specifier")
				}

				if val, ok := m.Get(token.Token{Type: token.IDENT, Value: fnName}); ok {
					env.DefineFast(localName, val)
				} else {
					return nil, fmt.Errorf("function %s not found in module %s", fnName, moduleName)
				}
			}
		case types.LibsModule:
			for _, sp := range n.Specifiers {
				var fnName, localName string

				if lit, ok := sp.(*ast.Literal); ok {
					if lit.Value.Type != token.IDENT {
						return nil, fmt.Errorf("pick target must be an identifier")
					}
					fnName = lit.Value.Value
					localName = fnName
				} else if us, ok := sp.(*ast.UseSpecifier); ok {
					if us.Remote == nil || us.Remote.Value == nil || us.Remote.Value.Type != token.IDENT {
						return nil, fmt.Errorf("invalid pick specifier")
					}
					fnName = us.Remote.Value.Value
					if us.Local != nil && us.Local.Value != nil {
						if us.Local.Value.Type != token.IDENT {
							return nil, fmt.Errorf("alias must be an identifier")
						}
						localName = us.Local.Value.Value
					} else {
						localName = fnName
					}
				} else {
					return nil, fmt.Errorf("invalid pick specifier")
				}

				if fn, ok := m.Get(token.Token{Type: token.IDENT, Value: fnName}); ok {
					env.DefineFast(localName, fn)
				} else {
					return nil, fmt.Errorf("function %s not found in module %s", fnName, moduleName)
				}
			}
		case *store.StoreObject:
			for _, sp := range n.Specifiers {
				var fnName, localName string

				if lit, ok := sp.(*ast.Literal); ok {
					if lit.Value.Type != token.IDENT {
						return nil, fmt.Errorf("pick target must be an identifier")
					}
					fnName = lit.Value.Value
					localName = fnName
				} else if us, ok := sp.(*ast.UseSpecifier); ok {
					if us.Remote == nil || us.Remote.Value == nil || us.Remote.Value.Type != token.IDENT {
						return nil, fmt.Errorf("invalid pick specifier")
					}
					fnName = us.Remote.Value.Value
					if us.Local != nil && us.Local.Value != nil {
						if us.Local.Value.Type != token.IDENT {
							return nil, fmt.Errorf("alias must be an identifier")
						}
						localName = us.Local.Value.Value
					} else {
						localName = fnName
					}
				} else {
					return nil, fmt.Errorf("invalid pick specifier")
				}

				if val, ok := m.Get(token.Token{Type: token.IDENT, Value: fnName}); ok {
					env.DefineFast(localName, val)
				} else {
					return nil, fmt.Errorf("function %s not found in module %s", fnName, moduleName)
				}
			}
		default:
			return nil, fmt.Errorf("invalid module type for pick: %T", mod)
		}

	default:
		return nil, fmt.Errorf("unknown use mode: %s", n.Mode)
	}

	return nil, nil
}
