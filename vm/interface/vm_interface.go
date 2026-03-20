package iface

import (
	"vine-lang/bytecode"
	"vine-lang/env"
)

// VMInterface 虚拟机接口
type VMInterface interface {
	// Run 运行虚拟机
	Run() (any, error)
	// RunLine 运行单行指令
	RunLine(op bytecode.Opcode, ins bytecode.Instructions) (any, error)
	// Push 压入值到栈
	Push(value any)
	// Pop 从栈弹出值
	Pop() any
	// currentFrame 获取当前帧
	CurrentFrame() *Frame
	// pushFrame 压入帧
	PushFrame(fn *bytecode.CompiledFunction)
	// GetConstants 获取常量池
	GetConstants() []any
	// GetStack 获取栈
	GetStack() []any
	// GetSP 获取栈指针
	GetSP() int
	// SetSP 设置栈指针
	SetSP(sp int)
	// GetFrameIndex 获取帧索引
	GetFrameIndex() int
	// SetFrameIndex 设置帧索引
	SetFrameIndex(index int)
	// GetFrames 获取帧数组
	GetFrames() []*Frame
	// SetFrames 设置帧数组
	SetFrames(frames []*Frame)
	// GetLocals 获取局部变量
	GetLocals() []any
	// SetLocals 设置局部变量
	SetLocals(locals []any)
	// GetEnv 获取环境
	GetEnv() *env.Environment
	// GetEnvVar 从环境获取变量
	GetEnvVar(name string) (any, bool)
	// SetEnvVar 设置环境变量
	SetEnvVar(name string, value any) error
	// DefineEnvVar 定义环境变量
	DefineEnvVar(name string, value any) error
	// DefineEnvConst 定义环境常量
	DefineEnvConst(name string, value any) error
	// AddAsyncTask 添加异步任务
	AddAsyncTask(fn *bytecode.CompiledFunction, toFunctions []*bytecode.CompiledFunction, basePointer int)
}

// Frame 虚拟机帧
type Frame struct {
	Fn          *bytecode.CompiledFunction
	Ip          int // instruction pointer
	BasePointer int // 栈基址，用于局部变量
	ToCount     int // to表达式数量
}

// VMFunc 虚拟机处理函数
type VMFunc func(v VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error)
