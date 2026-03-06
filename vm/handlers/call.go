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
	v.SetFrameIndex(frameIndex - 1)
	if frameIndex-1 < 0 {
		// 返回到主函数
		return returnValue, nil
	}

	v.SetSP(frame.BasePointer)

	// 将返回值压入上一个帧的栈
	if returnValue != nil {
		v.Push(returnValue)
	}

	return returnValue, nil
}
