package handlers

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandleCall 处理函数调用
func HandleCall(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	stack := v.GetStack()
	sp := v.GetSP()

	// 从指令中读取参数数量
	argCount := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	// 从栈中弹出函数
	fn := stack[sp-1-argCount]

	switch fn := fn.(type) {
	case *bytecode.CompiledFunction:
		// 创建新的帧
		// 栈结构: [..., 函数对象, 参数1, 参数2, ...]
		// sp指向参数后的位置
		// 函数对象在 sp-argCount-1
		// 第一个参数在 sp-argCount
		newFrame := &iface.Frame{
			Fn:          fn,
			Ip:          0,
			BasePointer: sp - argCount - 1, // 指向函数对象的位置
		}

		frames := v.GetFrames()
		frames = append(frames, newFrame)
		v.SetFrames(frames)
		v.SetFrameIndex(v.GetFrameIndex() + 1)

		// 将参数从当前帧复制到新帧的栈中
		// 栈结构: [..., 函数对象, 参数1, 参数2, ...]
		// sp指向参数后的位置
		base := newFrame.BasePointer + 1
		// 直接使用copy复制参数
		for i := 0; i < argCount; i++ {
			stack[base+i] = stack[sp-argCount+i]
		}
		// 调整栈指针，指向参数之后的位置
		// 这样函数执行时可以从这个位置开始使用栈
		newSp := base + argCount
		v.SetSP(newSp)

		// 将参数复制到locals数组中
		locals := v.GetLocals()
		for i := 0; i < argCount; i++ {
			locals[i] = stack[base+i]
		}
		v.SetLocals(locals)

	case func(...any) (any, error):
		// 处理Go函数调用
		args := make([]any, argCount)
		copy(args, stack[sp-argCount:sp])
		result, err := fn(args...)
		if err != nil {
			return nil, err
		}
		// 清除栈上的函数和参数
		v.SetSP(sp - argCount - 1)
		// 将返回值压入栈
		if result != nil {
			v.Push(result)
		}
		// 更新指令指针
		frame.Ip += 3
		return result, nil
	default:
		// 尝试使用reflect调用函数
		if reflect.TypeOf(fn).Kind() != reflect.Func {
			return nil, fmt.Errorf("calling non-function: %T , callee : %v", fn, fn)
		}
		args := make([]reflect.Value, argCount)
		for i := 0; i < argCount; i++ {
			args[i] = reflect.ValueOf(stack[sp-argCount+i])
		}
		results := reflect.ValueOf(fn).Call(args)
		// 先清除栈上的函数和参数
		v.SetSP(sp - argCount - 1)
		// 再将返回值压入栈
		if len(results) > 0 {
			v.Push(results[0].Interface())
		}
		// 更新指令指针
		frame.Ip += 3
		return nil, nil
	}

	frame.Ip += 3
	return nil, nil
}

// HandleReturn 处理函数返回
func HandleReturn(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	sp := v.GetSP()

	// 获取返回值
	var returnValue any
	if sp > frame.BasePointer {
		returnValue = v.Pop()
	}

	// 弹出当前帧
	frameIndex := v.GetFrameIndex()
	if frameIndex-1 < 0 {
		// 返回到主函数
		return returnValue, nil
	}

	// 获取上一个帧
	prevFrame := v.GetFrames()[frameIndex-1]
	// 保存原始帧索引，用于后续更新prevFrame.ToCount
	originalFrameIndex := frameIndex
	v.SetFrameIndex(frameIndex - 1)

	// 检查是否有to表达式需要调用
	if frame.ToCount > 0 {
		// 栈结构: [..., to1, to2, ...]
		// 我们需要从栈中弹出to表达式，并调用它们
		// 首先将返回值压入栈，作为to表达式的参数
		if returnValue != nil {
			v.Push(returnValue)
		}

		// 获取当前的栈指针
		sp := v.GetSP()

		// 计算to表达式在栈中的起始位置
		// 在HandleCallTask中，to表达式被压入栈中，位置在任务对象之后
		// 所以to表达式的起始位置是 prevFrame.BasePointer
		toStart := prevFrame.BasePointer

		// 计算当前应该调用的to表达式索引
		currentToIndex := prevFrame.ToCount - frame.ToCount

		if currentToIndex >= 0 && toStart+currentToIndex < sp {
			// 从栈中获取当前to表达式
			toFn := v.GetStack()[toStart+currentToIndex]
			if compiledFn, ok := toFn.(*bytecode.CompiledFunction); ok {
				// 调用to表达式
				// 栈结构: [..., to1, to2, ..., 返回值]
				// 我们需要将返回值作为参数传递给to表达式
				returnValue := v.Pop()
				v.Push(returnValue) // 将返回值压入栈，作为to表达式的参数

				// 创建新帧来执行to表达式
				newFrame := &iface.Frame{
					Fn:          compiledFn,
					Ip:          0,
					BasePointer: v.GetSP() - 1,     // 返回值在栈顶
					ToCount:     frame.ToCount - 1, // 剩余的to表达式数量
				}

				// 将返回值复制到locals数组中
				locals := v.GetLocals()
				locals[0] = returnValue
				v.SetLocals(locals)

				// 将剩余的to表达式保存到栈中
				// 注意：我们需要按照正确的顺序保存to表达式
				// to表达式在栈中的顺序是to1, to2, to3，我们已经调用了to3，所以需要保存to1, to2
				// 但是，我们不应该保存已经执行过的to表达式
				// 所以我们需要从栈中移除已执行的to表达式
				// 这可以通过调整toStart来实现
				newToStart := toStart + currentToIndex + 1
				if newToStart < sp {
					// 将剩余的to表达式保存到栈中
					for j := newToStart; j < sp; j++ {
						toFn := v.GetStack()[j]
						v.Push(toFn)
					}
				}

				// 更新prevFrame的BasePointer，以便在下一次返回时正确地访问栈
				prevFrame.BasePointer = v.GetSP()
				// 更新prevFrame的ToCount，以便在下一次返回时正确计算currentToIndex
				frames := v.GetFrames()
				frames[originalFrameIndex-1].ToCount = frame.ToCount - 1
				frames[originalFrameIndex-1].BasePointer = v.GetSP()

				frames = append(frames, newFrame)
				v.SetFrames(frames)
				v.SetFrameIndex(len(frames) - 1) // 设置为新帧的索引
				// 不重置栈指针，保持当前状态
				// 只调用当前to表达式，其他的会在下一次返回时调用
				return returnValue, nil
			}
		}
		// 所有to表达式都已执行完毕，继续正常返回流程
	}

	v.SetSP(frame.BasePointer)

	// 将返回值压入上一个帧的栈
	if returnValue != nil {
		v.Push(returnValue)
	}

	return returnValue, nil
}
