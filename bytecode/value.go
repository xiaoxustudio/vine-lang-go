package bytecode

type Integer struct{ Value int64 }

type String struct{ Value string }

type CompiledFunction struct {
	Instructions  Instructions
	NumLocals     int
	NumParameters int
}
