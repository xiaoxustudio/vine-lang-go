package vm

import (
	"vine-lang/bytecode"
	"vine-lang/compiler"
	"vine-lang/env"
	"vine-lang/vm/handlers"
	iface "vine-lang/vm/interface"
)

func NewVM(c *compiler.Compiler, env *env.Environment) *VM {
	v := &VM{
		constants:  c.GetConstantRaw(),
		stack:      make([]any, 256),
		sp:         0,
		frames:     make([]*iface.Frame, 1),
		frameIndex: 0,
		globals:    make([]any, 256),
		locals:     make([]any, 256),
		handlers:   [256]iface.VMFunc{},
		env:        env,
	}

	// 初始化主帧
	mainFrame := &iface.Frame{
		Fn:          c.Bytecode(),
		Ip:          0,
		BasePointer: 0,
	}
	v.frames[0] = mainFrame

	v.RegisterOpenCodeHandler(bytecode.OpConstant, handlers.HandleConstant)
	v.RegisterOpenCodeHandler(bytecode.OpFalse, handlers.HandleFalse)
	v.RegisterOpenCodeHandler(bytecode.OpTrue, handlers.HandleTrue)
	v.RegisterOpenCodeHandler(bytecode.OpPlus, handlers.HandlePlus)
	v.RegisterOpenCodeHandler(bytecode.OpMinus, handlers.HandleMinus)
	v.RegisterOpenCodeHandler(bytecode.OpMul, handlers.HandleMul)
	v.RegisterOpenCodeHandler(bytecode.OpDiv, handlers.HandleDiv)
	v.RegisterOpenCodeHandler(bytecode.OpIncrement, handlers.HandleIncrement)
	v.RegisterOpenCodeHandler(bytecode.OpDecrement, handlers.HandleDecrement)
	v.RegisterOpenCodeHandler(bytecode.OpEqual, handlers.HandleEqual)
	v.RegisterOpenCodeHandler(bytecode.OpNotEqual, handlers.HandleNotEqual)
	v.RegisterOpenCodeHandler(bytecode.OpLessThan, handlers.HandleLessThan)
	v.RegisterOpenCodeHandler(bytecode.OpLessEqual, handlers.HandleLessEqual)
	v.RegisterOpenCodeHandler(bytecode.OpGreaterThan, handlers.HandleGreaterThan)
	v.RegisterOpenCodeHandler(bytecode.OpGreaterEqual, handlers.HandleGreaterEqual)
	v.RegisterOpenCodeHandler(bytecode.OpPop, handlers.HandlePop)
	v.RegisterOpenCodeHandler(bytecode.OpDup, handlers.HandleDup)
	v.RegisterOpenCodeHandler(bytecode.OpJump, handlers.HandleJump)
	v.RegisterOpenCodeHandler(bytecode.OpJumpIfFalse, handlers.HandleJumpIfFalse)
	v.RegisterOpenCodeHandler(bytecode.OpJumpIfTrue, handlers.HandleJumpIfTrue)
	v.RegisterOpenCodeHandler(bytecode.OpLoop, handlers.HandleLoop)
	v.RegisterOpenCodeHandler(bytecode.OpBreak, handlers.HandleBreak)
	v.RegisterOpenCodeHandler(bytecode.OpContinue, handlers.HandleContinue)
	v.RegisterOpenCodeHandler(bytecode.OpSetGlobal, handlers.HandleSetGlobal)
	v.RegisterOpenCodeHandler(bytecode.OpSetConst, handlers.HandleSetConst)
	v.RegisterOpenCodeHandler(bytecode.OpGetGlobal, handlers.HandleGetGlobal)
	v.RegisterOpenCodeHandler(bytecode.OpGetLocal, handlers.HandleGetLocal)
	v.RegisterOpenCodeHandler(bytecode.OpSetLocal, handlers.HandleSetLocal)
	v.RegisterOpenCodeHandler(bytecode.OpCall, handlers.HandleCall)
	v.RegisterOpenCodeHandler(bytecode.OpReturn, handlers.HandleReturn)
	v.RegisterOpenCodeHandler(bytecode.OpGetMember, handlers.HandleGetMember)
	v.RegisterOpenCodeHandler(bytecode.OpIndex, handlers.HandleIndex)
	v.RegisterOpenCodeHandler(bytecode.OpSetIndex, handlers.HandleSetIndex)
	v.RegisterOpenCodeHandler(bytecode.OpSetMember, handlers.HandleSetMember)
	v.RegisterOpenCodeHandler(bytecode.OpExpose, handlers.HandleExpose)
	v.RegisterOpenCodeHandler(bytecode.OpArray, handlers.HandleArray)
	v.RegisterOpenCodeHandler(bytecode.OpObject, handlers.HandleObject)
	v.RegisterOpenCodeHandler(bytecode.OpTask, handlers.HandleTask)
	v.RegisterOpenCodeHandler(bytecode.OpCallTask, handlers.HandleCallTask)
	v.RegisterOpenCodeHandler(bytecode.OpTo, handlers.HandleTo)

	return v
}
