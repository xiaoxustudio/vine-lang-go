package handlers

import (
	"vine-lang/ast"
	iface "vine-lang/compiler/interface"
)

// HandleProgramStmt 处理程序语句
func HandleProgramStmt(c iface.CompilerInterface, node ast.Node) (any, error) {
	n := node.(*ast.ProgramStmt)
	for _, s := range n.Body {
		// 跳过注释语句
		if _, ok := s.(*ast.CommentStmt); ok {
			continue
		}
		_, err := c.Compile(s)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
