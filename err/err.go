package err

import (
	"fmt"
)

type Position struct {
	Filename string
	Line     int
	Column   int
}

type VError struct {
	Pos  Position
	Char string
	Msg  string
}

func (e *VError) Error() string {
	if e.Pos.Filename != "" {
		return fmt.Sprintf("%s:%d:%d: %s: ('%s')", e.Pos.Filename, e.Pos.Line, e.Pos.Column, e.Msg, e.Char)
	}
	return fmt.Sprintf("%d:%d: %s: ('%s')", e.Pos.Line, e.Pos.Column, e.Msg, e.Char)
}
