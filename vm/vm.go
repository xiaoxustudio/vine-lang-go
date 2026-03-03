package vm

import (
	"errors"
	"vine-lang/bytecode"
	"vine-lang/env"
)

type VMFunc func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error)

type VM struct {
	constants []any
	stack     []any
	sp        int // stack pointer

	frames     []*Frame // 函数调用栈
	frameIndex int

	globals  []any            // 全局变量
	locals   []any            // 局部变量
	handlers [256]VMFunc      // 使用数组代替map，避免哈希查找开销
	env      *env.Environment // 环境引用
}

type Frame struct {
	fn          *bytecode.CompiledFunction
	ip          int // instruction pointer
	basePointer int // 栈基址，用于局部变量
}

func (v *VM) RegisterOpenCodeHandler(op bytecode.Opcode, handler VMFunc) {
	v.handlers[op] = handler
}

func (v *VM) CallOpenCodeHandler(op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	handler := v.handlers[op]
	if handler == nil {
		return nil, errors.New("unknown opcode")
	}

	return handler(v, op, ins)
}
func (v *VM) Run() (any, error) {
	var result any
	for v.frameIndex >= 0 {
		frame := v.currentFrame()
		// 检查是否超出指令范围
		if frame.ip >= len(frame.fn.Instructions) {
			break
		}
		opcode := bytecode.Opcode(frame.fn.Instructions[frame.ip])
		handler := v.handlers[opcode]
		if handler == nil {
			return nil, errors.New("unknown opcode")
		}
		r, err := handler(v, opcode, frame.fn.Instructions)
		if err != nil {
			return nil, err
		}
		result = r
	}
	return result, nil
}

func (v *VM) pushFrame(fn *bytecode.CompiledFunction) {
	v.frameIndex++
	v.frames = append(v.frames, &Frame{fn: fn, ip: 0, basePointer: v.sp})
}

// currentFrame 获取当前帧，内联优化
func (v *VM) currentFrame() *Frame {
	return v.frames[v.frameIndex]
}

func (v *VM) RunLine(op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	return v.CallOpenCodeHandler(op, ins)
}

func (v *VM) push(value any) {
	v.stack[v.sp] = value
	v.sp++
}

func (v *VM) pop() any {
	v.sp--
	value := v.stack[v.sp]
	return value
}
