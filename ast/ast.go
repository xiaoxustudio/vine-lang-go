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
	NodeTernayExpression     NodeType = "TernayExpression"
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
	NodeSwitchStmtement     NodeType = "SwitchStmtement"
	NodeCaseBlockStatement  NodeType = "CaseBlockStatement"
	NodeDefaultCaseBlock    NodeType = "DefaultCaseBlockStatement"
	NodeExposeStmtement     NodeType = "ExposeStmtement"
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

// ================================== Common Nodes ==================================

type Literal struct {
	BaseNode
	Value *Token
}

func (l *Literal) String() string {
	return fmt.Sprintf("Literal(%s)", l.Value.String())
}

type Property struct {
	BaseNode
	Key   *Literal
	Value Expr
}

func (p *Property) String() string {
	return fmt.Sprintf("%s: %s", p.Key.String(), p.Value.String())
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

// UseDecl
type UseDecl struct {
	BaseNode
	Source     *Literal
	Specifiers []Specifier
	Mode       token.TokenType
}

func (u *UseDecl) String() string {
	return fmt.Sprintf("UseDecl(%s, %s, %s)", u.Source.String(), u.Specifiers, u.Mode)
}

// FunctionDecl
type FunctionDecl struct {
	BaseNode
	PreID     Token // 对应 preId: Token
	ID        *Literal
	Arguments *ArgsExpr
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

// UnaryExpr
type UnaryExpr struct {
	BaseNode
	Value    Expr
	Operator Token
	IsSuffix bool // 后缀运算符
}

func (u *UnaryExpr) String() string {
	if u.IsSuffix {
		return fmt.Sprintf("UnaryExpr(%s %s)", u.Value.String(), u.Operator.String())
	}
	return fmt.Sprintf("UnaryExpr(%s %s)", u.Operator.String(), u.Value.String())
}

// CompareExpr
type CompareExpr struct {
	BaseNode
	Left     Expr
	Right    Expr
	Operator Token
}

func (ce *CompareExpr) String() string {
	return fmt.Sprintf("CompareExpr(%s %s %s)", ce.Left.String(), ce.Operator.String(), ce.Right.String())
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
	Callee Expr
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

func (o *ObjectExpr) String() string {
	var props = make([]string, len(o.Properties))
	for i, prop := range o.Properties {
		props[i] = prop.String()
	}
	return fmt.Sprintf("ObjectExpr(%s)", strings.Join(props, ", "))
}

// ArrayExpr
type ArrayExpr struct {
	BaseNode
	Items []*Property
}

func (arr *ArrayExpr) String() string {
	var items = make([]string, len(arr.Items))
	for i, v := range arr.Items {
		items[i] = v.String()
	}
	return fmt.Sprintf("ArrayExpr(%s)", strings.Join(items, ", "))
}

// MemberExpr
type MemberExpr struct {
	BaseNode
	Object   Expr
	Property Expr
	Computed bool // 是否为计算属性 xxx[xxx]
}

func (m *MemberExpr) String() string {
	if m.Computed {
		return fmt.Sprintf("*MemberExpr(%s.%s)", m.Object.String(), m.Property.String())
	}
	return fmt.Sprintf("MemberExpr(%s.%s)", m.Object.String(), m.Property.String())
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
	Decl  Stmt
	Name  *Literal
	Value Expr
}

// BlockStmt
type BlockStmt struct {
	BaseNode
	Body []Stmt
}

func (b *BlockStmt) String() string {
	var body = make([]string, len(b.Body))
	for i, stmt := range b.Body {
		body[i] = stmt.String()
	}
	return fmt.Sprintf("BlockStmt(%s)", strings.Join(body, ", "))
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
	Alternate  Stmt // 可选
}

func (ifs *IfStmt) String() string {
	return fmt.Sprintf("IfStmt(%s, %s, %s)", ifs.Test.String(), ifs.Consequent.String(), ifs.Alternate.String())
}

// ForStmt
type ForStmt struct {
	BaseNode
	Init   Expr
	Value  Expr
	Update Expr
	Range  Expr
	Body   *BlockStmt
}

func (ifs *ForStmt) String() string {
	return fmt.Sprintf("ForStmt(%s, %s, %s, %s, %s)", ifs.Init.String(), ifs.Value.String(), ifs.Update.String(), ifs.Range, ifs.Body.String())
}

type SwitchCase struct {
	BaseNode
	Cond      Expr
	Body      *BlockStmt
	IsDefault bool // 是否为默认情况
}

func (s *SwitchCase) String() string {
	return fmt.Sprintf("SwitchCase(%s, %s)", s.Cond.String(), s.Body.String())
}

// SwitchStmt
type SwitchStmt struct {
	BaseNode
	Test  Expr
	Cases []Expr
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
