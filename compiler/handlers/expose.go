package handlers

import (
	"fmt"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
	"vine-lang/object/store"
)

func HandleExposeStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.ExposeStmt)

	// 获取主作用域的环境（第一个作用域）
	mainEnv := c.GetIndexScope(0).Env

	// 确保 Exports 对象存在
	if mainEnv.Exports == nil {
		mainEnv.Exports = store.NewStoreObject()
	}

	if n.Decl != nil {
		// 处理变量或函数声明
		switch decl := n.Decl.(type) {
		case *ast.FunctionDecl:
			// 编译函数定义，获取函数对象
			fn, err := c.Compile(n.Decl)
			if err != nil {
				return nil, err
			}

			// 将函数添加到导出列表
			if decl.ID != nil && decl.ID.Value != nil {
				if err := mainEnv.Exports.Define(*decl.ID.Value, fn); err != nil {
					return nil, err
				}
			}

		case *ast.VariableDecl:
			// 编译变量定义
			_, err := c.Compile(n.Decl)
			if err != nil {
				return nil, err
			}

			// 将变量添加到导出列表
			if decl.Name.Value != nil {
				val, exists := mainEnv.Get(*decl.Name.Value)
				if !exists {
					return nil, fmt.Errorf("expose target not found: %s", decl.Name.Value.Value)
				}
				if err := mainEnv.Exports.Define(*decl.Name.Value, val); err != nil {
					return nil, err
				}
			}

		default:
			return nil, fmt.Errorf("invalid expose declaration type")
		}
	} else if n.Name != nil && n.Name.Value != nil {
		// 处理直接导出值的情况: expose name = value
		if n.Value != nil {
			// 编译值表达式
			_, err := c.Compile(n.Value)
			if err != nil {
				return nil, err
			}

			// 将变量名添加到常量池
			namePos := c.AddConstant(n.Name.Value.Value)

			// 先将值定义为全局变量
			c.Emit(bytecode.OpSetGlobal, namePos)

			// 发出 OpExpose 指令，将值添加到导出列表
			c.Emit(bytecode.OpExpose, namePos)
		}
	}

	return nil, nil
}
