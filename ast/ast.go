package ast

import (
	"fmt"
	"strings"
	"vine-lang/token"
)

type Token = token.Token

// ================================== Node Type Definitions ==================================

// NodeType 定义所有 AST 节点的类型
type NodeType string

const (
	NodeProgram          NodeType = "Program"
	NodeLiteral          NodeType = "Literal"
	NodeProperty         NodeType = "Property"
	NodeCommentStatement NodeType = "CommentStatement"
	NodeEmptyLine        NodeType = "EmptyLine"
	/* Declaration */
	NodeLambdaFunctionDecl  NodeType = "LambdaFunctionDecl"
	NodeFunctionDeclaration NodeType = "FunctionDeclaration"
	NodeVariableDeclaration NodeType = "VariableDeclaration"
	NodeUseDeclaration      NodeType = "UseDeclaration"
	NodeUseSpecifier        NodeType = "UseSpecifier"
	NodeUseDefaultSpecifier NodeType = "UseDefaultSpecifier"
	NodeTemplateElement     NodeType = "TemplateElement"
	/* expr */
	NodeBinaryExpression     NodeType = "BinaryExpression"
	NodeArrayExpression      NodeType = "ArrayExpression"
	NodeObjectExpression     NodeType = "ObjectExpression"
	NodeMemberExpression     NodeType = "MemberExpression"
	NodeCallExpression       NodeType = "CallExpression"
	NodeAssignmentExpression NodeType = "AssignmentExpression"
	NodeCompareExpression    NodeType = "CompareExpression"
	NodeEqualExpression      NodeType = "EqualExpression"
	NodeTernayExpression     NodeType = "TernayExpression" // 保持原拼写
	NodeRangeExpression      NodeType = "RangeExpression"
	NodeIterableExpression   NodeType = "IterableExpression"
	NodeToExpression         NodeType = "ToExpression"
	NodeTemplateLiteral      NodeType = "TemplateLiteralExpression"
	/* stmt */
	NodeRunStatement        NodeType = "RunStatement"
	NodeWaitStatement       NodeType = "WaitStatement"
	NodeTaskStatement       NodeType = "TaskStatement"
	NodeBlockStatement      NodeType = "BlockStatement"
	NodeReturnStatement     NodeType = "ReturnStatement"
	NodeExpressionStatement NodeType = "ExpressionStatement"
	NodeIfStatement         NodeType = "IfStatement"
	NodeForStatement        NodeType = "ForStatement"
	NodeSwitchStmtement     NodeType = "SwitchStmtement" // 保持原拼写
	NodeCaseBlockStatement  NodeType = "CaseBlockStatement"
	NodeDefaultCaseBlock    NodeType = "DefaultCaseBlockStatement"
	NodeExposeStmtement     NodeType = "ExposeStmtement" // 保持原拼写
)

// ================================== Base Interfaces & Structs ==================================

// Node 所有节点的基础接口
type Node interface {
	GetType() NodeType
}

// Expr 表达式接口
type Expr interface {
	Node
	String() string
}

// Stmt 语句接口
type Stmt interface {
	Node
	String() string
}

// BaseNode 基础节点实现
type BaseNode struct {
	Type NodeType
	ID   *Token
}

func (n *BaseNode) GetType() NodeType {
	return n.Type
}

func (n *BaseNode) String() string {
	return n.ID.Value
}

// ================================== Helper Interfaces for Unions ==================================

// Specifier UseDeclaration 中 specifiers 的联合类型接口
type Specifier interface {
	Node
}

// SwitchCase Switch 中 cases 的联合类型接口
type SwitchCase interface {
	Node
}

// ================================== Common Nodes ==================================

type Literal struct {
	BaseNode
	Value Token
}

func (l *Literal) String() string {
	return l.Value.String()
}

type Property struct {
	BaseNode
	Key   *Literal
	Value Expr
}

type TemplateElement struct {
	BaseNode
	Value *Literal
}

type EmptyLineStmt struct {
	BaseNode
}

type CommentStmt struct {
	BaseNode
	Value Token
}

func (l *CommentStmt) String() string {
	return fmt.Sprintf("#CommentStmt(%s)", l.Value.String())
}

// ================================== Declarations ==================================

// UseSpecifier
type UseSpecifier struct {
	BaseNode
	Remote *Literal
	Local  *Literal
}

// UseDefaultSpecifier
type UseDefaultSpecifier struct {
	BaseNode
	Local *Literal
}

// UseDecl
type UseDecl struct {
	BaseNode
	Source     *Literal
	Specifiers []Specifier
}

func (u *UseDecl) String() string {
	return fmt.Sprintf("UseDecl(%s, %s)", u.Source.String(), u.Specifiers)
}

// FunctionDecl
type FunctionDecl struct {
	BaseNode
	PreID     Token // 对应 preId: Token
	ID        *Literal
	Arguments []Expr
	Body      *BlockStmt
}

// LambdaFunctionDecl
type LambdaFunctionDecl struct {
	BaseNode
	Arguments []Expr
	Body      *BlockStmt
}

// VariableDecl
type VariableDecl struct {
	BaseNode
	Name    *Literal
	Value   Expr
	IsConst bool
}

func (v *VariableDecl) String() string {
	var prefix string
	if v.IsConst {
		prefix = "const"
	} else {
		prefix = "let"
	}
	return fmt.Sprintf("%s %s = %s", prefix, v.Name.String(), v.Value)
}

// ================================== Expressions ==================================

// RangeExpr
type RangeExpr struct {
	BaseNode
	Start Expr
	End   Expr
	Step  Token
}

// EqualExpr
type EqualExpr struct {
	BaseNode
	Left     Expr
	Right    Expr
	Operator Token
}

// CompareExpr
type CompareExpr struct {
	BaseNode
	Left     Expr
	Right    Expr
	Operator Token
}

// ArgsExpr
type ArgsExpr struct {
	BaseNode
	Arguments []Expr
}

func (a *ArgsExpr) GetType() NodeType {
	return a.Type
}

func (a *ArgsExpr) String() string {
	var args = make([]string, len(a.Arguments))
	for i, arg := range a.Arguments {
		args[i] = arg.String()
	}
	return fmt.Sprintf("ArgsExpr(%s)", strings.Join(args, ", "))
}

// CallExpr
type CallExpr struct {
	BaseNode
	Callee *Literal
	Args   ArgsExpr
}

func (c *CallExpr) GetType() NodeType {
	return c.Type
}
func (c *CallExpr) String() string {
	return fmt.Sprintf("CallExpr(%s, %s)", c.Callee.String(), c.Args.String())
}

type AssignmentExpr struct {
	BaseNode
	Left     Expr
	Right    Expr
	Operator Token
}

func (a *AssignmentExpr) String() string {
	return fmt.Sprintf("AssignmentExpr(%s %s %s)", a.Left.String(), a.Operator.String(), a.Right.String())
}

// TernaryExpr (Type: TernayExpression)
type TernaryExpr struct {
	BaseNode
	Condition  Expr
	Consequent Expr
	Alternate  Expr
}

// ObjectExpr
type ObjectExpr struct {
	BaseNode
	Properties []*Property
}

// ArrayExpr (原 items 为 Property 数组，这里假设应该是 Expr 或类似结构，按原样定义)
type ArrayExpr struct {
	BaseNode
	Items []*Property // TypeScript 中定义为 Property 数组，此处保持一致
}

// MemberExpr
type MemberExpr struct {
	BaseNode
	Object   Expr
	Property Expr // 或 []Expr，视具体语法而定，通常单一属性
	Computed bool
}

// BinaryExpr
type BinaryExpr struct {
	BaseNode
	Left     Expr
	Right    Expr
	Operator Token
}

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("BinaryExpr(%s %s %s)", b.Left.String(), b.Operator.String(), b.Right.String())
}

// ToExpr
type ToExpr struct {
	BaseNode
	Body      *BlockStmt
	Arguments []Expr
}

// TemplateLiteralExpr
type TemplateLiteralExpr struct {
	BaseNode
	Quotes []Node // TemplateElement | Expr
}

// IterableExpr (在 NodeType 中存在但未定义接口，补充定义)
type IterableExpr struct {
	BaseNode
	// 根据实际语法补充字段
}

// ================================== Statements ==================================

// RunStmt
type RunStmt struct {
	BaseNode
	Callee *CallExpr
	To     []*ToExpr
}

// ExposeStmt
type ExposeStmt struct {
	BaseNode
	ID         Token
	Body       Expr
	Specifiers []Expr
}

// BlockStmt
type BlockStmt struct {
	BaseNode
	Body []Expr
}

// CaseBlockStmt
type CaseBlockStmt struct {
	BaseNode
	Body *BlockStmt
	Test Expr
}

// DefaultCaseBlockStmt
type DefaultCaseBlockStmt struct {
	BaseNode
	Body *BlockStmt
	Test Expr
}

// ReturnStmt
type ReturnStmt struct {
	BaseNode
	Value Expr
}

// ProgramStmt
type ProgramStmt struct {
	BaseNode
	Body []Stmt
}

func (p *ProgramStmt) Print() {
	for _, stmt := range p.Body {
		if stmt == nil {
			continue
		}
		fmt.Println(stmt.String())
	}
}

// IfStmt
type IfStmt struct {
	BaseNode
	Test       Expr
	Consequent *BlockStmt
	Alternate  *BlockStmt // 可选
}

// ForStmt
type ForStmt struct {
	BaseNode
	Init   Expr
	Value  Expr // 可选
	Range  *RangeExpr
	Update Expr
	Body   *BlockStmt
}

// SwitchStmt
type SwitchStmt struct {
	BaseNode
	Test  Expr
	Cases []SwitchCase
}

// ExpressionStmt
type ExpressionStmt struct {
	BaseNode
	Expression Expr
}

func (e *ExpressionStmt) String() string {
	return fmt.Sprintf("ExpressionStmt: %v", e.Expression)
}

// TaskStmt
type TaskStmt struct {
	BaseNode
	Fn *FunctionDecl
}

// WaitStmt
type WaitStmt struct {
	BaseNode
	Async *RunStmt
}
