package handlers

import (
	"vine-lang/ast"
	"vine-lang/bytecode"
	iface "vine-lang/compiler/interface"
)

func HandleSwitchCaseStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.SwitchCase)
	currentScope := c.CurrentScope()

	// 如果不是default case，需要比较条件
	if !n.IsDefault {
		// 对每个条件进行比较
		for _, cond := range n.Conds {
			// 复制switch的测试值到栈顶
			// 使用OpDup指令复制栈顶值
			c.Emit(bytecode.OpDup)
			// 编译case条件
			_, err := c.Compile(cond)
			if err != nil {
				return nil, err
			}
			// 比较测试值和case条件
			c.Emit(bytecode.OpEqual)
			// 如果相等，跳转到case体
			// 先占位，稍后填充跳转位置
			jumpIfTruePos := c.Emit(bytecode.OpJumpIfTrue, 9999)
			// 保存跳转位置和对应的case体位置
			if currentScope.JumpPositions == nil {
				currentScope.JumpPositions = make([]int, 0)
			}
			// 保存跳转位置和占位的case体位置（稍后修复）
			currentScope.JumpPositions = append(currentScope.JumpPositions, jumpIfTruePos)
			currentScope.JumpPositions = append(currentScope.JumpPositions, -1) // 占位，稍后修复为case体开始位置
		}

		// 所有条件都不匹配，跳过当前case
		// 先占位，稍后填充跳转位置
		skipCaseJumpPos := c.Emit(bytecode.OpJump, 9999)
		// 保存跳过当前case的跳转位置和特殊标记（-2表示这是跳过case的跳转）
		if currentScope.JumpPositions == nil {
			currentScope.JumpPositions = make([]int, 0)
		}
		currentScope.JumpPositions = append(currentScope.JumpPositions, skipCaseJumpPos)
		currentScope.JumpPositions = append(currentScope.JumpPositions, -2)
	} else {
		// default case，直接跳转到这里
		// 需要从switch语句中获取跳转位置
		if currentScope.DefaultCasePos == -1 {
			currentScope.DefaultCasePos = len(currentScope.Instructions)
			// default case需要弹出测试值，因为跳转时测试值还在栈上
			c.Emit(bytecode.OpPop)
		}
	}

	// 记录case体开始位置（在编译完所有条件之后）
	caseBodyPos := len(currentScope.Instructions)

	// 编译语句块
	_, err := c.Compile(n.Body)
	if err != nil {
		return nil, err
	}

	// case体执行完后，需要弹出所有测试值副本
	if !n.IsDefault {
		// 每个条件比较前复制了一次测试值，所以需要弹出 len(n.Conds) 个测试值副本
		for i := 0; i < len(n.Conds); i++ {
			c.Emit(bytecode.OpPop)
		}
		// 还需要弹出原始的测试值
		c.Emit(bytecode.OpPop)
	}

	// case体执行完后，跳转到switch结束位置
	// 先占位，稍后填充跳转位置
	jumpPos := c.Emit(bytecode.OpJump, 9999)
	// 保存跳转位置以便后续修复
	if currentScope.JumpPositions == nil {
		currentScope.JumpPositions = make([]int, 0)
	}
	// 保存跳转位置和特殊标记（-1表示这是case体结束的跳转）
	currentScope.JumpPositions = append(currentScope.JumpPositions, jumpPos)
	currentScope.JumpPositions = append(currentScope.JumpPositions, -1)

	if !n.IsDefault {
		// 找到所有属于这个case的条件跳转
		// 在编译这个case之前，jumpPositions的长度
		beforeCasePos := len(currentScope.JumpPositions) - 2*len(n.Conds) - 2
		// 遍历当前case的所有条件跳转
		for i := beforeCasePos; i < len(currentScope.JumpPositions)-2; i += 2 {
			jumpPos := currentScope.JumpPositions[i]
			// 检查这是否是条件跳转
			op := bytecode.Opcode(currentScope.Instructions[jumpPos])
			if op == bytecode.OpJumpIfTrue {
				// 更新跳转目标位置为case体开始位置
				currentScope.JumpPositions[i+1] = caseBodyPos
			}
		}
	}

	return nil, nil
}
