package handlers

import (
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandleCallTask 处理任务调用
// OpCallTask操作码用于调用一个任务，并处理与之关联的to表达式
// 现在实现为异步调用，类似 JavaScript Promise
func HandleCallTask(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	stack := v.GetStack()
	sp := v.GetSP()

	// 检查栈是否为空
	if sp <= frame.BasePointer {
		return nil, nil
	}

	// 从栈中弹出to表达式（如果有）
	// 栈结构: [..., task, to1, to2, ...]
	var toFunctions []*bytecode.CompiledFunction
	for sp > frame.BasePointer {
		// 检查栈顶元素是否是编译好的函数（to表达式）
		if fn, ok := stack[sp-1].(*bytecode.CompiledFunction); ok {
			// 使用prepend方式添加，保持原始顺序
			toFunctions = append([]*bytecode.CompiledFunction{fn}, toFunctions...)
			sp--
		} else {
			// 如果不是函数，则停止弹出
			break
		}
	}

	// 检查是否还有任务对象
	if sp <= frame.BasePointer {
		return nil, nil
	}

	// 从栈中弹出任务对象或函数
	sp--
	var fn *bytecode.CompiledFunction
	if task, ok := stack[sp].(*Task); ok {
		// 如果是任务对象，获取其中的函数
		fn = task.Fn
	} else if compiledFn, ok := stack[sp].(*bytecode.CompiledFunction); ok {
		// 如果是编译好的函数，直接使用
		fn = compiledFn
	} else {
		// 如果不是任务也不是函数，直接返回
		return nil, nil
	}

	// 更新栈指针
	v.SetSP(sp)

	// 将异步任务添加到队列
	// 注意：这里使用 sp 作为 basePointer，因为任务执行时需要从正确的位置开始
	v.AddAsyncTask(fn, toFunctions, sp)

	// 更新指令指针
	// OpCallTask指令有2字节的操作数
	frame.Ip += 3

	// 返回nil，不阻塞后续代码执行
	return nil, nil
}
