package types

import "vine-lang/bytecode"

// AsyncTask 异步任务
type AsyncTask struct {
	Fn          *bytecode.CompiledFunction
	ToFunctions []*bytecode.CompiledFunction
	BasePointer int
}
