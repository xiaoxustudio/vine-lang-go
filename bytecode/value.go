package bytecode

import "fmt"

type Integer struct{ Value int64 }

type String struct{ Value string }

type CompiledFunction struct {
	Instructions  Instructions
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) String() string {
	return fmt.Sprintf("<fn %p>", cf)
}
