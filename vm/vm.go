package vm

import (
	"errors"
	"fmt"
	"vine-lang/bytecode"
	"vine-lang/env"
	"vine-lang/token"
	iface "vine-lang/vm/interface"
	"vine-lang/vm/types"
)

type VM struct {
	constants []any
	stack     []any
	sp        int // stack pointer

	frames     []*iface.Frame // 函数调用栈
	frameIndex int

	globals  []any             // 全局变量
	locals   []any             // 局部变量
	handlers [256]iface.VMFunc // 使用数组代替map，避免哈希查找开销
	env      *env.Environment  // 环境引用

	// 异步任务队列
	asyncTasks []*types.AsyncTask
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
	for v.frameIndex >= 0 {
		frame := v.frames[v.frameIndex]
		// 检查是否超出指令范围
		if frame.Ip >= len(frame.Fn.Instructions) {
			// 函数执行完毕，触发返回处理
			// 只有当不是主帧时才触发返回处理
			if v.frameIndex > 0 {
				handler := v.handlers[bytecode.OpReturn]
				if handler != nil {
					r, err := handler(v, bytecode.OpReturn, frame.Fn.Instructions)
					if err != nil {
						return nil, err
					}
					result = r
					continue
				}
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

	// 主程序执行完毕后，处理异步任务队列
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

func (v *VM) Push(value any) {
	v.stack[v.sp] = value
	v.sp++
}

func (v *VM) Pop() any {
	v.sp--
	value := v.stack[v.sp]
	return value
}
