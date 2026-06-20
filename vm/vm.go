package vm

import (
	"errors"
	"fmt"
	"vine-lang/bytecode"
	"vine-lang/env"
	"vine-lang/token"
	iface "vine-lang/vm/interface"
	"vine-lang/vm/types"
	"vine-lang/vm/vutils"
)

const (
	valueKindAny  uint8 = 0
	valueKindInt  uint8 = 1
	valueKindBool uint8 = 2
)

type VM struct {
	constants []any
	stack     []any
	stackMeta []uint8
	stackInt  []int64
	stackBool []bool
	sp        int // stack pointer

	frames     []*iface.Frame // 函数调用栈
	frameIndex int

	globals   []any // 全局变量
	locals    []any // 局部变量
	localMeta []uint8
	localInt  []int64
	localBool []bool
	handlers  [256]iface.VMFunc // 使用数组代替map，避免哈希查找开销
	env       *env.Environment  // 环境引用

	// 异步任务队列
	asyncTasks []*types.AsyncTask

	// 栈扩容策略
	stackCapacity int
	localCapacity int
}

func (v *VM) RegisterOpenCodeHandler(op bytecode.Opcode, handler iface.VMFunc) {
	v.handlers[op] = handler
}

func (v *VM) CallOpenCodeHandler(op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	handler := v.handlers[op]
	if handler == nil {
		return nil, errors.New("unknown opcode")
	}

	return handler(v, op, ins)
}

// GetConstants 获取常量池
func (v *VM) GetConstants() []any {
	return v.constants
}

// GetStack 获取栈
func (v *VM) GetStack() []any {
	return v.stack
}

// GetSP 获取栈指针
func (v *VM) GetSP() int {
	return v.sp
}

// SetSP 设置栈指针
func (v *VM) SetSP(sp int) {
	v.sp = sp
}

// GetFrameIndex 获取帧索引
func (v *VM) GetFrameIndex() int {
	return v.frameIndex
}

// SetFrameIndex 设置帧索引
func (v *VM) SetFrameIndex(index int) {
	v.frameIndex = index
}

// GetFrames 获取帧数组
func (v *VM) GetFrames() []*iface.Frame {
	return v.frames
}

// SetFrames 设置帧数组
func (v *VM) SetFrames(frames []*iface.Frame) {
	v.frames = frames
}

// GetLocals 获取局部变量
func (v *VM) GetLocals() []any {
	return v.locals
}

// SetLocals 设置局部变量
func (v *VM) SetLocals(locals []any) {
	v.locals = locals
	v.rebuildLocalMeta()
}

func (v *VM) expandLocals(minIndex int) {
	if minIndex >= len(v.locals) {
		newCap := v.localCapacity * 2
		if v.localCapacity == 0 {
			newCap = 256
		}
		if minIndex >= newCap {
			newCap = minIndex + 128
		}
		newLocals := make([]any, newCap)
		newLocalMeta := make([]uint8, newCap)
		newLocalInt := make([]int64, newCap)
		newLocalBool := make([]bool, newCap)
		copy(newLocals, v.locals)
		copy(newLocalMeta, v.localMeta)
		copy(newLocalInt, v.localInt)
		copy(newLocalBool, v.localBool)
		v.locals = newLocals
		v.localMeta = newLocalMeta
		v.localInt = newLocalInt
		v.localBool = newLocalBool
		v.localCapacity = newCap
	}
}

func (v *VM) LoadLocalToStack(index int) any {
	v.expandLocals(index)
	v.expandStack()
	switch v.localMeta[index] {
	case valueKindInt:
		val := v.localInt[index]
		v.stackMeta[v.sp] = valueKindInt
		v.stackInt[v.sp] = val
		v.sp++
		return val
	case valueKindBool:
		val := v.localBool[index]
		v.stackMeta[v.sp] = valueKindBool
		v.stackBool[v.sp] = val
		v.sp++
		return val
	default:
		value := v.locals[index]
		v.stackMeta[v.sp] = valueKindAny
		v.stack[v.sp] = value
		v.sp++
		return value
	}
}

func (v *VM) StoreLocalFromStack(index int) any {
	v.expandLocals(index)
	v.sp--
	switch v.stackMeta[v.sp] {
	case valueKindInt:
		val := v.stackInt[v.sp]
		v.localMeta[index] = valueKindInt
		v.localInt[index] = val
		return val
	case valueKindBool:
		val := v.stackBool[v.sp]
		v.localMeta[index] = valueKindBool
		v.localBool[index] = val
		return val
	default:
		value := v.stack[v.sp]
		v.localMeta[index] = valueKindAny
		v.locals[index] = value
		return value
	}
}

// GetEnv 获取环境
func (v *VM) GetEnv() *env.Environment {
	return v.env
}

// GetEnvVar 从环境获取变量
func (v *VM) GetEnvVar(name string) (any, bool) {
	nameToken := token.Token{Type: token.IDENT, Value: name}
	return v.env.Get(nameToken)
}

// SetEnvVar 设置环境变量
func (v *VM) SetEnvVar(name string, value any) error {
	nameToken := token.Token{Type: token.IDENT, Value: name}
	v.env.Set(nameToken, value)
	return nil
}

// DefineEnvVar 定义环境变量
func (v *VM) DefineEnvVar(name string, value any) error {
	nameToken := token.Token{Type: token.IDENT, Value: name}
	return v.env.Define(nameToken, value)
}

// DefineEnvConst 定义环境常量
func (v *VM) DefineEnvConst(name string, value any) error {
	nameToken := token.Token{Type: token.IDENT, Value: name}
	v.env.DefineConst(nameToken, value)
	return nil
}

// AddAsyncTask 添加异步任务
func (v *VM) AddAsyncTask(fn *bytecode.CompiledFunction, toFunctions []*bytecode.CompiledFunction, basePointer int) {
	asyncTask := &types.AsyncTask{
		Fn:          fn,
		ToFunctions: toFunctions,
		BasePointer: basePointer,
	}
	v.asyncTasks = append(v.asyncTasks, asyncTask)
}

func (v *VM) Run() (any, error) {
	var result any
	stack := v.stack
	stackMeta := v.stackMeta
	stackInt := v.stackInt
	stackBool := v.stackBool
	locals := v.locals
	localMeta := v.localMeta
	localInt := v.localInt
	localBool := v.localBool
	constants := v.constants
	sp := v.sp
	for v.frameIndex >= 0 {
		frame := v.frames[v.frameIndex]
		// 检查是否超出指令范围
		if frame.Ip >= len(frame.Fn.Instructions) {
			// 函数执行完毕，触发返回处理
			// 只有当不是主帧时才触发返回处理
			if v.frameIndex > 0 {
				v.materializeFastState(sp)
				v.sp = sp
				handler := v.handlers[bytecode.OpReturn]
				if handler != nil {
					r, err := handler(v, bytecode.OpReturn, frame.Fn.Instructions)
					if err != nil {
						return nil, err
					}
					result = r
					stack = v.stack
					stackMeta = v.stackMeta
					stackInt = v.stackInt
					stackBool = v.stackBool
					locals = v.locals
					localMeta = v.localMeta
					localInt = v.localInt
					localBool = v.localBool
					constants = v.constants
					sp = v.sp
					v.rebuildFastState(sp)
					continue
				}
			}
			break
		}
		ins := frame.Fn.Instructions
		ip := frame.Ip
		opcode := bytecode.Opcode(ins[ip])
		switch opcode {
		case bytecode.OpConstant:
			constIndex := int(ins[ip+1]) | int(ins[ip+2])<<8
			constant := constants[constIndex]
			if intVal, ok := constant.(int64); ok {
				stackMeta[sp] = valueKindInt
				stackInt[sp] = intVal
			} else if boolVal, ok := constant.(bool); ok {
				stackMeta[sp] = valueKindBool
				stackBool[sp] = boolVal
			} else {
				stackMeta[sp] = valueKindAny
				stack[sp] = constant
			}
			sp++
			frame.Ip = ip + 3
			continue
		case bytecode.OpGetLocal:
			localIndex := int(ins[ip+1]) | int(ins[ip+2])<<8
			switch localMeta[localIndex] {
			case valueKindInt:
				stackMeta[sp] = valueKindInt
				stackInt[sp] = localInt[localIndex]
			case valueKindBool:
				stackMeta[sp] = valueKindBool
				stackBool[sp] = localBool[localIndex]
			default:
				stackMeta[sp] = valueKindAny
				stack[sp] = locals[localIndex]
			}
			sp++
			frame.Ip = ip + 3
			continue
		case bytecode.OpSetLocal:
			localIndex := int(ins[ip+1]) | int(ins[ip+2])<<8
			sp--
			switch stackMeta[sp] {
			case valueKindInt:
				localMeta[localIndex] = valueKindInt
				localInt[localIndex] = stackInt[sp]
			case valueKindBool:
				localMeta[localIndex] = valueKindBool
				localBool[localIndex] = stackBool[sp]
			default:
				localMeta[localIndex] = valueKindAny
				locals[localIndex] = stack[sp]
			}
			frame.Ip = ip + 3
			continue
		case bytecode.OpPop:
			sp--
			frame.Ip = ip + 1
			continue
		case bytecode.OpDup:
			if sp > 0 {
				stackMeta[sp] = stackMeta[sp-1]
				switch stackMeta[sp-1] {
				case valueKindInt:
					stackInt[sp] = stackInt[sp-1]
				case valueKindBool:
					stackBool[sp] = stackBool[sp-1]
				default:
					stack[sp] = stack[sp-1]
				}
				sp++
			}
			frame.Ip = ip + 1
			continue
		case bytecode.OpJump:
			offset := int(ins[ip+1]) | int(ins[ip+2])<<8
			frame.Ip += offset
			continue
		case bytecode.OpLoop:
			offset := int(int16(uint16(ins[ip+1]) | uint16(ins[ip+2])<<8))
			frame.Ip += offset
			continue
		case bytecode.OpJumpIfFalse:
			sp--
			offset := int(ins[ip+1]) | int(ins[ip+2])<<8
			isTruthy := true
			switch stackMeta[sp] {
			case valueKindInt:
				isTruthy = stackInt[sp] != 0
			case valueKindBool:
				isTruthy = stackBool[sp]
			default:
				isTruthy = vutils.IsTruthy(stack[sp])
			}
			if !isTruthy {
				frame.Ip += offset
			} else {
				frame.Ip = ip + 3
			}
			continue
		case bytecode.OpIncrement:
			top := sp - 1
			if stackMeta[top] == valueKindInt {
				stackInt[top] = stackInt[top] + 1
				frame.Ip = ip + 1
				continue
			}
		case bytecode.OpDecrement:
			top := sp - 1
			if stackMeta[top] == valueKindInt {
				stackInt[top] = stackInt[top] - 1
				frame.Ip = ip + 1
				continue
			}
		case bytecode.OpMinus:
			rightIndex := sp - 1
			leftIndex := sp - 2
			if stackMeta[leftIndex] == valueKindInt && stackMeta[rightIndex] == valueKindInt {
				stackInt[leftIndex] = stackInt[leftIndex] - stackInt[rightIndex]
				stackMeta[leftIndex] = valueKindInt
				sp--
				frame.Ip = ip + 1
				continue
			}
		case bytecode.OpDiv:
			rightIndex := sp - 1
			leftIndex := sp - 2
			if stackMeta[leftIndex] == valueKindInt && stackMeta[rightIndex] == valueKindInt {
				leftVal := stackInt[leftIndex]
				rightVal := stackInt[rightIndex]
				if rightVal == 0 {
					return nil, errors.New("division by zero")
				}
				if leftVal%rightVal == 0 {
					stackInt[leftIndex] = leftVal / rightVal
					stackMeta[leftIndex] = valueKindInt
				} else {
					stack[leftIndex] = float64(leftVal) / float64(rightVal)
					stackMeta[leftIndex] = valueKindAny
				}
				sp--
				frame.Ip = ip + 1
				continue
			}
		case bytecode.OpLessThan:
			rightIndex := sp - 1
			leftIndex := sp - 2
			if stackMeta[leftIndex] == valueKindInt && stackMeta[rightIndex] == valueKindInt {
				resultVal := stackInt[leftIndex] < stackInt[rightIndex]
				stackMeta[leftIndex] = valueKindBool
				stackBool[leftIndex] = resultVal
				sp--
				frame.Ip = ip + 1
				continue
			}
		case bytecode.OpPlus:
			rightIndex := sp - 1
			leftIndex := sp - 2
			if stackMeta[leftIndex] == valueKindInt && stackMeta[rightIndex] == valueKindInt {
				stackInt[leftIndex] = stackInt[leftIndex] + stackInt[rightIndex]
				stackMeta[leftIndex] = valueKindInt
				sp--
				frame.Ip = ip + 1
				continue
			}
		case bytecode.OpMul:
			rightIndex := sp - 1
			leftIndex := sp - 2
			if stackMeta[leftIndex] == valueKindInt && stackMeta[rightIndex] == valueKindInt {
				stackInt[leftIndex] = stackInt[leftIndex] * stackInt[rightIndex]
				stackMeta[leftIndex] = valueKindInt
				sp--
				frame.Ip = ip + 1
				continue
			}
		}
		v.materializeFastState(sp)
		v.sp = sp
		handler := v.handlers[opcode]
		if handler == nil {
			return nil, fmt.Errorf("unknown opcode %d", opcode)
		}
		r, err := handler(v, opcode, ins)
		if err != nil {
			return nil, err
		}
		result = r
		v.rebuildFastState(v.sp)
		stack = v.stack
		stackMeta = v.stackMeta
		stackInt = v.stackInt
		stackBool = v.stackBool
		locals = v.locals
		localMeta = v.localMeta
		localInt = v.localInt
		localBool = v.localBool
		constants = v.constants
		sp = v.sp
	}

	// 主程序执行完毕后，处理异步任务队列
	v.materializeFastState(sp)
	v.sp = sp
	v.runAsyncTasks()
	return result, nil
}

// runAsyncTasks 执行异步任务队列中的所有任务
func (v *VM) runAsyncTasks() (any, error) {
	var result any

	for _, task := range v.asyncTasks {

		// 创建新的帧来执行异步任务
		newFrame := &iface.Frame{
			Fn:          task.Fn,
			Ip:          0,
			BasePointer: task.BasePointer,
			ToCount:     len(task.ToFunctions),
		}

		// 保存当前帧索引
		savedFrameIndex := v.frameIndex

		// 添加新帧
		v.frames = append(v.frames, newFrame)
		v.frameIndex = len(v.frames) - 1

		// 保存当前的栈指针（在压入to表达式之前）
		taskSP := v.sp

		// 将to表达式压入栈
		for _, fn := range task.ToFunctions {
			v.Push(fn)
		}

		// 更新新帧的BasePointer，指向to表达式在栈中的起始位置
		// 这样to表达式就位于 [BasePointer, BasePointer + ToCount) 的范围内
		newFrame.BasePointer = taskSP

		// 执行任务
		for v.frameIndex > savedFrameIndex {
			frame := v.frames[v.frameIndex]
			// 检查是否超出指令范围
			if frame.Ip >= len(frame.Fn.Instructions) {
				// 函数执行完毕，触发返回处理
				handler := v.handlers[bytecode.OpReturn]
				if handler != nil {
					r, err := handler(v, bytecode.OpReturn, frame.Fn.Instructions)
					if err != nil {
						return nil, err
					}
					result = r
					continue
				}
				break
			}
			opcode := bytecode.Opcode(frame.Fn.Instructions[frame.Ip])
			handler := v.handlers[opcode]
			if handler == nil {
				return nil, fmt.Errorf("unknown opcode %d", opcode)
			}
			r, err := handler(v, opcode, frame.Fn.Instructions)
			if err != nil {
				return nil, err
			}
			result = r
		}

		// 恢复帧索引
		v.frameIndex = savedFrameIndex

		// 恢复栈指针到任务执行前的位置
		v.sp = taskSP
	}

	// 清空异步任务队列
	v.asyncTasks = v.asyncTasks[:0]

	return result, nil
}

func (v *VM) PushFrame(fn *bytecode.CompiledFunction) {
	v.frameIndex++
	v.frames = append(v.frames, &iface.Frame{Fn: fn, Ip: 0, BasePointer: v.sp})
}

// CurrentFrame 获取当前帧，内联优化
func (v *VM) CurrentFrame() *iface.Frame {
	return v.frames[v.frameIndex]
}

func (v *VM) RunLine(op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	return v.CallOpenCodeHandler(op, ins)
}

func (v *VM) expandStack() {
	if v.sp >= len(v.stack) {
		newCap := v.stackCapacity * 2
		if v.stackCapacity == 0 {
			newCap = 256
		}
		newStack := make([]any, newCap)
		newStackMeta := make([]uint8, newCap)
		newStackInt := make([]int64, newCap)
		newStackBool := make([]bool, newCap)
		copy(newStack, v.stack)
		copy(newStackMeta, v.stackMeta)
		copy(newStackInt, v.stackInt)
		copy(newStackBool, v.stackBool)
		v.stack = newStack
		v.stackMeta = newStackMeta
		v.stackInt = newStackInt
		v.stackBool = newStackBool
		v.stackCapacity = newCap
	}
}

func (v *VM) Push(value any) {
	v.expandStack()
	v.stack[v.sp] = value
	switch val := value.(type) {
	case int64:
		v.stackMeta[v.sp] = valueKindInt
		v.stackInt[v.sp] = val
	case bool:
		v.stackMeta[v.sp] = valueKindBool
		v.stackBool[v.sp] = val
	default:
		v.stackMeta[v.sp] = valueKindAny
	}
	v.sp++
}

func (v *VM) Pop() any {
	v.sp--
	switch v.stackMeta[v.sp] {
	case valueKindInt:
		return v.stackInt[v.sp]
	case valueKindBool:
		return v.stackBool[v.sp]
	default:
		return v.stack[v.sp]
	}
}



func (v *VM) materializeFastState(sp int) {
	for i := 0; i < sp; i++ {
		switch v.stackMeta[i] {
		case valueKindInt:
			v.stack[i] = v.stackInt[i]
		case valueKindBool:
			v.stack[i] = v.stackBool[i]
		}
	}
	for i := 0; i < len(v.locals); i++ {
		switch v.localMeta[i] {
		case valueKindInt:
			v.locals[i] = v.localInt[i]
		case valueKindBool:
			v.locals[i] = v.localBool[i]
		}
	}
}

func (v *VM) rebuildFastState(sp int) {
	for i := 0; i < sp; i++ {
		switch val := v.stack[i].(type) {
		case int64:
			v.stackMeta[i] = valueKindInt
			v.stackInt[i] = val
		case bool:
			v.stackMeta[i] = valueKindBool
			v.stackBool[i] = val
		default:
			v.stackMeta[i] = valueKindAny
		}
	}
	v.rebuildLocalMeta()
}

func (v *VM) rebuildLocalMeta() {
	for i := 0; i < len(v.locals); i++ {
		switch val := v.locals[i].(type) {
		case int64:
			v.localMeta[i] = valueKindInt
			v.localInt[i] = val
		case bool:
			v.localMeta[i] = valueKindBool
			v.localBool[i] = val
		default:
			v.localMeta[i] = valueKindAny
		}
	}
}
