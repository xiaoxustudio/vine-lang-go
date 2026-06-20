package handlers

import (
	"encoding/binary"
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func HandleSwitchStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.SwitchStmt)
	currentScope := c.CurrentScope()

	// 保存当前的跳转位置列表和default case位置
	oldJumpPositions := currentScope.JumpPositions
	oldDefaultCasePos := currentScope.DefaultCasePos
	currentScope.JumpPositions = make([]int, 0)
	currentScope.DefaultCasePos = -1

	// 编译条件表达式，结果压入栈顶
	_, err := c.Compile(n.Test)
	if err != nil {
		return nil, err
	}

	// 编译每个case
	for _, sc := range n.Cases {
		_, err := c.Compile(&sc)
		if err != nil {
			return nil, err
		}
	}

	// 如果有default case，添加跳转到default case的指令
	var defaultJumpPos int = -1
	if currentScope.DefaultCasePos != 0 {
		// 所有case都不匹配时，跳转到default case
		// 先占位，稍后填充跳转位置
		defaultJumpPos = c.Emit(bytecode.OpJump, 9999)
		// 保存跳转位置和default case位置
		currentScope.JumpPositions = append(currentScope.JumpPositions, defaultJumpPos)
		currentScope.JumpPositions = append(currentScope.JumpPositions, currentScope.DefaultCasePos)
	} else {
		// 没有default case，弹出测试值
		c.Emit(bytecode.OpPop)
	}

	// switch结束位置
	switchEndPos := len(currentScope.Instructions)

	// jumpPositions数组格式: [jumpPos1, targetPos1, jumpPos2, targetPos2, ...]
	for i := 0; i < len(currentScope.JumpPositions); i += 2 {
		jumpPos := currentScope.JumpPositions[i]
		targetPos := currentScope.JumpPositions[i+1]

		// 检查这是哪种跳转
		op := bytecode.Opcode(currentScope.Instructions[jumpPos])
		switch op {
		case bytecode.OpJumpIfTrue:
			// 条件跳转，跳转到对应的case体开始位置
			offset := uint16(targetPos - jumpPos)
			binary.LittleEndian.PutUint16(currentScope.Instructions[jumpPos+1:], offset)
		case bytecode.OpJump:
			// 无条件跳转
			if targetPos == -1 {
				// case体结束的跳转，跳转到switch结束位置
				offset := uint16(switchEndPos - jumpPos)
				binary.LittleEndian.PutUint16(currentScope.Instructions[jumpPos+1:], offset)
			} else if targetPos == -2 {
				// 跳过当前case的跳转，跳转到下一个case的开始位置
				// 找到下一个case的开始位置
				nextCasePos := switchEndPos
				for j := i + 2; j < len(currentScope.JumpPositions); j += 2 {
					if currentScope.JumpPositions[j+1] != -1 && currentScope.JumpPositions[j+1] != -2 {
						nextCasePos = currentScope.JumpPositions[j+1]
						break
					}
				}
				offset := uint16(nextCasePos - jumpPos)
				binary.LittleEndian.PutUint16(currentScope.Instructions[jumpPos+1:], offset)
			} else {
				// 跳转到default case
				offset := uint16(targetPos - jumpPos)
				binary.LittleEndian.PutUint16(currentScope.Instructions[jumpPos+1:], offset)
			}
		}
	}

	// 恢复之前的跳转位置列表和default case位置
	currentScope.JumpPositions = oldJumpPositions
	currentScope.DefaultCasePos = oldDefaultCasePos

	return nil, nil
}
