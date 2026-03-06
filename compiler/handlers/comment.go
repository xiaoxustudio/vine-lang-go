package handlers

import (
	"vine-lang/ast"
)

// HandleCommentStmt 处理注释语句
// 注释语句在编译时被忽略，不生成任何字节码
func HandleCommentStmt(node ast.Node) (any, error) {
	return nil, nil
}
