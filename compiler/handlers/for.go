package handlers

import (
	"encoding/binary"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func HandleForStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.ForStmt)

	// 检查是否是 for in 循环
	if n.Range != nil {
		// for i in array 形式的循环
		// 编译数组表达式
		_, err := c.Compile(n.Range)
		if err != nil {
			return nil, err
		}

		// 保存循环开始位置
		loopStartPos := len(c.CurrentScope().Instructions)

		// 编译循环变量（如果存在）
		if n.Init != nil {
			_, err = c.Compile(n.Init)
			if err != nil {
				return nil, err
			}
		}

		// 发出循环跳转指令（占位）
		loopJumpPos := c.Emit(bytecode.OpLoop, 9999)

		// 为循环体创建新的作用域
		currentEnv := c.CurrentScope().Env
		c.EnterScope(currentEnv)

		// 编译循环体
		_, err = c.Compile(&n.Body)
		if err != nil {
			return nil, err
		}

		// 退出循环体作用域，并保存指令
		loopScope := c.LeaveScope()

		// 将循环体指令传递给父作用域
		if loopScope != nil && len(loopScope.Instructions) > 0 {
			parentScope := c.CurrentScope()
			if parentScope != nil {
				parentScope.Instructions = append(parentScope.Instructions, loopScope.Instructions...)
			}
		}

		// 发出跳转到循环开始的指令
		currentScope := c.CurrentScope()
		offset := uint16(loopStartPos - len(currentScope.Instructions))
		c.Emit(bytecode.OpLoop, int(offset))

		// 修改循环跳转指令，使其跳转到循环结束位置
		offset = uint16(len(currentScope.Instructions) - loopJumpPos)
		binary.LittleEndian.PutUint16(currentScope.Instructions[loopJumpPos+1:], offset)

		return nil, nil
	} else {
		// for i := 0; i < 10; i++ 形式的循环
		// 编译初始化表达式（在当前作用域中，不进入新作用域）
		if n.Init != nil {
			_, err := c.Compile(n.Init)
			if err != nil {
				return nil, err
			}
		}

		// 保存循环开始位置
		loopStartPos := len(c.CurrentScope().Instructions)

		// 声明条件跳转位置变量
		var jumpIfFalsePos int

		// 编译条件表达式
		if n.Value != nil {
			_, err := c.Compile(n.Value)
			if err != nil {
				return nil, err
			}

			// 发出条件跳转指令（如果条件为假，跳出循环）
			// 此时还不知道跳转位置，先发出一个占位指令
			jumpIfFalsePos = c.Emit(bytecode.OpJumpIfFalse, 9999)
		}

		// 保存父作用域的符号表大小
		parentScope := c.CurrentScope()
		parentSymbolCount := len(parentScope.SymbolTable)

		// 为循环体创建新的作用域
		currentEnv := c.CurrentScope().Env
		c.EnterScope(currentEnv)

		// 编译循环体
		_, err := c.Compile(&n.Body)
		if err != nil {
			return nil, err
		}

		// 退出循环体作用域，并保存指令
		loopScope := c.LeaveScope()

		// 将循环体指令传递给父作用域
		if loopScope != nil && len(loopScope.Instructions) > 0 {
			// 调整局部变量索引，加上父作用域的符号表大小
			for i := 0; i < len(loopScope.Instructions); i++ {
				op := bytecode.Opcode(loopScope.Instructions[i])
				if op == bytecode.OpGetLocal || op == bytecode.OpSetLocal {
					// 读取局部变量索引
					localIndex := int(binary.LittleEndian.Uint16(loopScope.Instructions[i+1:]))
					// 调整索引
					newIndex := localIndex + parentSymbolCount
					binary.LittleEndian.PutUint16(loopScope.Instructions[i+1:], uint16(newIndex))
				}
			}
			// 将调整后的指令添加到父作用域
			parentScope.Instructions = append(parentScope.Instructions, loopScope.Instructions...)
			// 将循环体的符号表添加到父作用域
			for name, index := range loopScope.SymbolTable {
				parentScope.SymbolTable[name] = index + parentSymbolCount
			}
		}

		// 编译更新表达式（在循环体之后，但在循环跳转之前）
		if n.Update != nil {
			_, err := c.Compile(n.Update)
			if err != nil {
				return nil, err
			}
		}

		// 发出跳转到循环开始的指令
		currentScope := c.CurrentScope()
		offset := uint16(loopStartPos - len(currentScope.Instructions))
		c.Emit(bytecode.OpLoop, int(offset))

		// 修改条件跳转指令，使其跳转到循环结束位置
		if n.Value != nil {
			offset = uint16(len(currentScope.Instructions) - jumpIfFalsePos)
			binary.LittleEndian.PutUint16(currentScope.Instructions[jumpIfFalsePos+1:], offset)
		}

		return nil, nil
	}
}
