package ast

import (
	"fmt"
	"strings"
	"vine-lang/token"
)

type NodeType uint16

const (
	NodeTypeProgramStmt NodeType = iota
	NodeTypeBlockStmt
	NodeTypeUseDecl
	NodeTypeExpressionStmt
	NodeTypeVariableDecl
	NodeTypeExposeStmt
	NodeTypeForStmt
	NodeTypeIfStmt
	NodeTypeFunctionDecl
	NodeTypeLambdaFunctionDecl
	NodeTypeReturnStmt
	NodeTypeSwitchStmt
	NodeTypeTaskStmt
	NodeTypeWaitStmt
	NodeTypeCallTaskFn
	NodeTypeToExpr
	NodeTypeAssignmentExpr
	NodeTypeCompareExpr
	NodeTypeBinaryExpr
	NodeTypeProperty
	NodeTypeArrayExpr
	NodeTypeObjectExpr
	NodeTypeMemberExpr
	NodeTypeArgsExpr
	NodeTypeCallExpr
	NodeTypeUnaryExpr
	NodeTypeLiteral
	NodeTypeUseSpecifier
	NodeTypeSwitchCase

	NodeTypeCommentStmt
	NodeTypeBaseNode
)

type Node interface {
	NodeType() NodeType
	String() string
}

type BaseNode struct {
	Type  NodeType
	Token *token.Token
}

func (n *BaseNode) String() string {
	return n.Token.Value
}

func (n *BaseNode) NodeType() NodeType {
	return n.Type
}

type Expr interface {
	Node
	String() string
}

// Stmt 语句接口
type Stmt interface {
	Node
	String() string
}

// ================================== Helper Interfaces for Unions ==================================

// Specifier UseDeclaration 中 specifiers 的联合类型接口
type Specifier interface {
	Node
}

// ================================== Common Nodes ==================================

type Literal struct {
	BaseNode
	Value *token.Token
}

func NewLiteral(value *token.Token) *Literal {
	return &Literal{
		BaseNode: BaseNode{Type: NodeTypeLiteral},
		Value:    value,
	}
}

func (l *Literal) String() string {
	return fmt.Sprintf("Literal(%s)", l.Value.String())
}

func (l *Literal) NodeType() NodeType {
	return l.Type
}

type Property struct {
	BaseNode
	Key   *Literal
	Value Expr
}

func NewProperty(key *Literal, value Expr) *Property {
	return &Property{
		BaseNode: BaseNode{Type: NodeTypeProperty},
		Key:      key,
		Value:    value,
	}
}

func (p *Property) String() string {
	return fmt.Sprintf("%s: %s", p.Key.String(), p.Value.String())
}

func (p *Property) NodeType() NodeType {
	return p.Type
}

// type TemplateElement struct {
// 	BaseNode
// 	Value *Literal
// }

// type EmptyLineStmt struct {
// 	BaseNode
// }

type CommentStmt struct {
	BaseNode
	Value token.Token
}

func NewCommentStmt(value token.Token) *CommentStmt {
	return &CommentStmt{
		BaseNode: BaseNode{Type: NodeTypeCommentStmt},
		Value:    value,
	}
}

func (l *CommentStmt) String() string {
	return fmt.Sprintf("#CommentStmt(%s)", l.Value.String())
}

func (l *CommentStmt) NodeType() NodeType {
	return l.Type
}

// ================================== Declarations ==================================

// UseSpecifier
type UseSpecifier struct {
	BaseNode
	Remote *Literal
	Local  *Literal
}

func NewUseSpecifier(remote, local *Literal) *UseSpecifier {
	return &UseSpecifier{
		BaseNode: BaseNode{Type: NodeTypeUseSpecifier},
		Remote:   remote,
		Local:    local,
	}
}

func (u *UseSpecifier) String() string {
	return fmt.Sprintf("UseSpecifier(%s, %s)", u.Remote.String(), u.Local.String())
}

func (u *UseSpecifier) NodeType() NodeType {
	return u.Type
}

// UseDecl
type UseDecl struct {
	BaseNode
	Source     *Literal
	Specifiers []Specifier
	Mode       token.TokenType
}

func NewUseDecl(source *Literal, specifiers []Specifier, mode token.TokenType) *UseDecl {
	return &UseDecl{
		BaseNode:   BaseNode{Type: NodeTypeUseDecl},
		Source:     source,
		Specifiers: specifiers,
		Mode:       mode,
	}
}

func (u *UseDecl) String() string {
	return fmt.Sprintf("UseDecl(%s, %s, %s)", u.Source.String(), u.Specifiers, u.Mode)
}

func (u *UseDecl) NodeType() NodeType {
	return u.Type
}

// FunctionDecl
type FunctionDecl struct {
	BaseNode
	ID        *Literal
	Arguments *ArgsExpr
	Body      *BlockStmt
}

func NewFunctionDecl(id *Literal, args *ArgsExpr, body *BlockStmt) *FunctionDecl {
	return &FunctionDecl{
		BaseNode:  BaseNode{Type: NodeTypeFunctionDecl},
		ID:        id,
		Arguments: args,
		Body:      body,
	}
}

func (f *FunctionDecl) String() string {
	return fmt.Sprintf("FunctionDecl( %s, %s)", f.ID.String(), f.Body.String())
}

func (f *FunctionDecl) NodeType() NodeType {
	return f.Type
}

// LambdaFunctionDecl
type LambdaFunctionDecl struct {
	BaseNode
	Args ArgsExpr
	Body BlockStmt
}

func NewLambdaFunctionDecl(args ArgsExpr, body BlockStmt) *LambdaFunctionDecl {
	return &LambdaFunctionDecl{
		BaseNode: BaseNode{Type: NodeTypeLambdaFunctionDecl},
		Args:     args,
		Body:     body,
	}
}

func (l *LambdaFunctionDecl) String() string {
	return fmt.Sprintf("LambdaFunctionDecl(%s, %s)", l.Args.String(), l.Body.String())
}

func (l *LambdaFunctionDecl) NodeType() NodeType {
	return l.Type
}

// VariableDecl
type VariableDecl struct {
	BaseNode
	Name    Literal
	Value   Expr
	IsConst bool
}

func NewVariableDecl(name Literal, value Expr, isConst bool) *VariableDecl {
	return &VariableDecl{
		BaseNode: BaseNode{Type: NodeTypeVariableDecl},
		Name:     name,
		Value:    value,
		IsConst:  isConst,
	}
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

func (v *VariableDecl) NodeType() NodeType {
	return v.Type
}

// ================================== Expressions ==================================

// // RangeExpr
// type RangeExpr struct {
// 	BaseNode
// 	Start Expr
// 	End   Expr
// 	Step  token.Token
// }

// UnaryExpr
type UnaryExpr struct {
	BaseNode
	Value    Expr
	Operator token.Token
	IsSuffix bool // 后缀运算符
}

func NewUnaryExpr(value Expr, operator token.Token, isSuffix bool) *UnaryExpr {
	return &UnaryExpr{
		BaseNode: BaseNode{Type: NodeTypeUnaryExpr},
		Value:    value,
		Operator: operator,
		IsSuffix: isSuffix,
	}
}

func (u *UnaryExpr) String() string {
	if u.IsSuffix {
		return fmt.Sprintf("UnaryExpr(%s %s)", u.Value.String(), u.Operator.String())
	}
	return fmt.Sprintf("UnaryExpr(%s %s)", u.Operator.String(), u.Value.String())
}

func (u *UnaryExpr) NodeType() NodeType {
	return u.Type
}

// CompareExpr
type CompareExpr struct {
	BaseNode
	Left     Expr
	Right    Expr
	Operator token.Token
}

func NewCompareExpr(left, right Expr, operator token.Token) *CompareExpr {
	return &CompareExpr{
		BaseNode: BaseNode{Type: NodeTypeCompareExpr},
		Left:     left,
		Right:    right,
		Operator: operator,
	}
}

func (ce *CompareExpr) String() string {
	return fmt.Sprintf("CompareExpr(%s %s %s)", ce.Left.String(), ce.Operator.String(), ce.Right.String())
}

func (ce *CompareExpr) NodeType() NodeType {
	return ce.Type
}

// ArgsExpr
type ArgsExpr struct {
	BaseNode
	Arguments []Expr
}

func NewArgsExpr(args []Expr) *ArgsExpr {
	return &ArgsExpr{
		BaseNode:  BaseNode{Type: NodeTypeArgsExpr},
		Arguments: args,
	}
}

func (a *ArgsExpr) NodeType() NodeType {
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

func NewCallExpr(callee Expr, args ArgsExpr) *CallExpr {
	return &CallExpr{
		BaseNode: BaseNode{Type: NodeTypeCallExpr},
		Callee:   callee,
		Args:     args,
	}
}

func (c *CallExpr) NodeType() NodeType {
	return c.Type
}

func (c *CallExpr) String() string {
	return fmt.Sprintf("CallExpr(%s, %s)", c.Callee.String(), c.Args.String())
}

type AssignmentExpr struct {
	BaseNode
	Left     Expr
	Right    Expr
	Operator token.Token
}

func NewAssignmentExpr(left, right Expr, operator token.Token) *AssignmentExpr {
	return &AssignmentExpr{
		BaseNode: BaseNode{Type: NodeTypeAssignmentExpr},
		Left:     left,
		Right:    right,
		Operator: operator,
	}
}

func (a *AssignmentExpr) String() string {
	return fmt.Sprintf("AssignmentExpr(%s %s %s)", a.Left.String(), a.Operator.String(), a.Right.String())
}

func (a *AssignmentExpr) NodeType() NodeType {
	return a.Type
}

// // TernaryExpr (Type: TernayExpression)
// type TernaryExpr struct {
// 	BaseNode
// 	Condition  Expr
// 	Consequent Expr
// 	Alternate  Expr
// }

// ObjectExpr
type ObjectExpr struct {
	BaseNode
	Properties []*Property
}

func NewObjectExpr(properties []*Property) *ObjectExpr {
	return &ObjectExpr{
		BaseNode:   BaseNode{Type: NodeTypeObjectExpr},
		Properties: properties,
	}
}

func (o *ObjectExpr) String() string {
	var props = make([]string, len(o.Properties))
	for i, prop := range o.Properties {
		props[i] = prop.String()
	}
	return fmt.Sprintf("ObjectExpr(%s)", strings.Join(props, ", "))
}

func (o *ObjectExpr) NodeType() NodeType {
	return o.Type
}

// ArrayExpr
type ArrayExpr struct {
	BaseNode
	Items []*Property
}

func NewArrayExpr(items []*Property) *ArrayExpr {
	return &ArrayExpr{
		BaseNode: BaseNode{Type: NodeTypeArrayExpr},
		Items:    items,
	}
}

func (arr *ArrayExpr) String() string {
	var items = make([]string, len(arr.Items))
	for i, v := range arr.Items {
		items[i] = v.String()
	}
	return fmt.Sprintf("ArrayExpr(%s)", strings.Join(items, ", "))
}

func (arr *ArrayExpr) NodeType() NodeType {
	return arr.Type
}

// MemberExpr
type MemberExpr struct {
	BaseNode
	Object   Expr
	Property Expr
	Computed bool // 是否为计算属性 xxx[xxx]
}

func NewMemberExpr(object, property Expr, computed bool) *MemberExpr {
	return &MemberExpr{
		BaseNode: BaseNode{Type: NodeTypeMemberExpr},
		Object:   object,
		Property: property,
		Computed: computed,
	}
}

func (m *MemberExpr) String() string {
	if m.Computed {
		return fmt.Sprintf("*MemberExpr(%s.%s)", m.Object.String(), m.Property.String())
	}
	return fmt.Sprintf("MemberExpr(%s.%s)", m.Object.String(), m.Property.String())
}

func (m *MemberExpr) NodeType() NodeType {
	return m.Type
}

// BinaryExpr
type BinaryExpr struct {
	BaseNode
	Left     Expr
	Right    Expr
	Operator token.Token
}

func NewBinaryExpr(left, right Expr, operator token.Token) *BinaryExpr {
	return &BinaryExpr{
		BaseNode: BaseNode{Type: NodeTypeBinaryExpr},
		Left:     left,
		Right:    right,
		Operator: operator,
	}
}

func (b *BinaryExpr) String() string {
	return fmt.Sprintf("BinaryExpr(%s %s %s)", b.Left.String(), b.Operator.String(), b.Right.String())
}

func (b *BinaryExpr) NodeType() NodeType {
	return b.Type
}

// // TemplateLiteralExpr
// type TemplateLiteralExpr struct {
// 	BaseNode
// 	Quotes []Node // TemplateElement | Expr
// }

// // IterableExpr (在 NodeType 中存在但未定义接口，补充定义)
// type IterableExpr struct {
// 	BaseNode
// 	// 根据实际语法补充字段
// }

// ================================== Statements ==================================

// ExposeStmt
type ExposeStmt struct {
	BaseNode
	Decl  Stmt
	Name  *Literal
	Value Expr
}

func NewExposeStmt(decl Stmt, name *Literal, value Expr) *ExposeStmt {
	return &ExposeStmt{
		BaseNode: BaseNode{Type: NodeTypeExposeStmt},
		Decl:     decl,
		Name:     name,
		Value:    value,
	}
}

func (e *ExposeStmt) String() string {
	return fmt.Sprintf("ExposeStmt(%s, %s, %s)", e.Decl.String(), e.Name.String(), e.Value.String())
}

func (e *ExposeStmt) NodeType() NodeType {
	return e.Type
}

// BlockStmt
type BlockStmt struct {
	BaseNode
	Body []Stmt
}

func NewBlockStmt(body []Stmt) *BlockStmt {
	return &BlockStmt{
		BaseNode: BaseNode{Type: NodeTypeBlockStmt},
		Body:     body,
	}
}

func (b *BlockStmt) String() string {
	var body = make([]string, len(b.Body))
	for i, stmt := range b.Body {
		body[i] = stmt.String()
	}
	return fmt.Sprintf("BlockStmt(%s)", strings.Join(body, ", "))
}

func (b *BlockStmt) NodeType() NodeType {
	return b.Type
}

// ReturnStmt
type ReturnStmt struct {
	BaseNode
	Value Expr
}

func NewReturnStmt(value Expr) *ReturnStmt {
	return &ReturnStmt{
		BaseNode: BaseNode{Type: NodeTypeReturnStmt},
		Value:    value,
	}
}

func (r *ReturnStmt) String() string {
	return fmt.Sprintf("ReturnStmt(%s)", r.Value.String())
}

func (r *ReturnStmt) NodeType() NodeType {
	return r.Type
}

// ProgramStmt
type ProgramStmt struct {
	BaseNode
	Body []Stmt
}

func NewProgramStmt(body []Stmt) *ProgramStmt {
	return &ProgramStmt{
		BaseNode: BaseNode{Type: NodeTypeProgramStmt},
		Body:     body,
	}
}

func (p *ProgramStmt) Print() {
	for _, stmt := range p.Body {
		if stmt == nil {
			continue
		}
		fmt.Println(stmt.String())
	}
}

func (p *ProgramStmt) NodeType() NodeType {
	return p.Type
}

func (p *ProgramStmt) String() string {
	var body = make([]string, len(p.Body))
	for i, stmt := range p.Body {
		body[i] = stmt.String()
	}
	return fmt.Sprintf("ProgramStmt(%s)", strings.Join(body, ", "))
}

// IfStmt
type IfStmt struct {
	BaseNode
	Test       Expr
	Consequent *BlockStmt
	Alternate  Stmt // 可选
}

func NewIfStmt(test Expr, consequent *BlockStmt, alternate Stmt) *IfStmt {
	return &IfStmt{
		BaseNode:   BaseNode{Type: NodeTypeIfStmt},
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

func (ifs *IfStmt) String() string {
	return fmt.Sprintf("IfStmt(%s, %s, %s)", ifs.Test.String(), ifs.Consequent.String(), ifs.Alternate.String())
}

func (ifs *IfStmt) NodeType() NodeType {
	return ifs.Type
}

// ForStmt
type ForStmt struct {
	BaseNode
	Init   Expr
	Value  Expr
	Update Expr
	Range  Expr
	Body   BlockStmt
}

func NewForStmt(init, value, update, range_ Expr, body BlockStmt) *ForStmt {
	return &ForStmt{
		BaseNode: BaseNode{Type: NodeTypeForStmt},
		Init:     init,
		Value:    value,
		Update:   update,
		Range:    range_,
		Body:     body,
	}
}

func (ifs *ForStmt) String() string {
	return fmt.Sprintf("ForStmt(%s, %s, %s, %s, %s)", ifs.Init.String(), ifs.Value.String(), ifs.Update.String(), ifs.Range, ifs.Body.String())
}

func (ifs *ForStmt) NodeType() NodeType {
	return ifs.Type
}

type SwitchCase struct {
	BaseNode
	Conds     []Expr
	Body      *BlockStmt
	IsDefault bool // 是否为默认情况
}

func NewSwitchCase(conds []Expr, body *BlockStmt, isDefault bool) *SwitchCase {
	return &SwitchCase{
		BaseNode:  BaseNode{Type: NodeTypeSwitchCase},
		Conds:     conds,
		Body:      body,
		IsDefault: isDefault,
	}
}

func (s *SwitchCase) NodeType() NodeType {
	return s.Type
}

func (s *SwitchCase) String() string {
	var conds = make([]string, len(s.Conds))
	for i, cond := range s.Conds {
		conds[i] = cond.String()
	}
	return fmt.Sprintf("SwitchCase(%s, %s)", strings.Join(conds, ","), s.Body.String())
}

// SwitchStmt
type SwitchStmt struct {
	BaseNode
	Test  Expr
	Cases []Expr
}

func NewSwitchStmt(test Expr, cases []Expr) *SwitchStmt {
	return &SwitchStmt{
		BaseNode: BaseNode{Type: NodeTypeSwitchStmt},
		Test:     test,
		Cases:    cases,
	}
}

func (s *SwitchStmt) NodeType() NodeType {
	return s.Type
}

func (s *SwitchStmt) String() string {
	var cases = make([]string, len(s.Cases))
	for i, case_ := range s.Cases {
		cases[i] = case_.String()
	}
	return fmt.Sprintf("SwitchStmt(%s, %s)", s.Test.String(), cases)
}

// ExpressionStmt
type ExpressionStmt struct {
	BaseNode
	Expression Expr
}

func NewExpressionStmt(expr Expr) *ExpressionStmt {
	return &ExpressionStmt{
		BaseNode:   BaseNode{Type: NodeTypeExpressionStmt},
		Expression: expr,
	}
}

func (e *ExpressionStmt) NodeType() NodeType {
	return e.Type
}

func (e *ExpressionStmt) String() string {
	return fmt.Sprintf("ExpressionStmt: %v", e.Expression)
}

// TaskStmt
type TaskStmt struct {
	BaseNode
	Fn FunctionDecl
}

func NewTaskStmt(fn FunctionDecl) *TaskStmt {
	return &TaskStmt{
		BaseNode: BaseNode{Type: NodeTypeTaskStmt},
		Fn:       fn,
	}
}

func (task *TaskStmt) NodeType() NodeType {
	return task.Type
}

func (task *TaskStmt) String() string {
	return fmt.Sprintf("TaskStmt(%v)", task.Fn)
}

// WaitStmt
type WaitStmt struct {
	BaseNode
	Async Expr
}

func NewWaitStmt(async Expr) *WaitStmt {
	return &WaitStmt{
		BaseNode: BaseNode{Type: NodeTypeWaitStmt},
		Async:    async,
	}
}

func (wait *WaitStmt) NodeType() NodeType {
	return wait.Type
}

func (wait *WaitStmt) String() string {
	return fmt.Sprintf("WaitStmt(%v)", wait.Async)
}

type CallTaskFn struct {
	BaseNode
	Target CallExpr
	To     ToExpr
	Catch  *LambdaFunctionDecl
}

func NewCallTaskFn(target CallExpr, to ToExpr, catch *LambdaFunctionDecl) *CallTaskFn {
	return &CallTaskFn{
		BaseNode: BaseNode{Type: NodeTypeCallTaskFn},
		Target:   target,
		To:       to,
		Catch:    catch,
	}
}

func (c *CallTaskFn) NodeType() NodeType {
	return c.Type
}

func (c *CallTaskFn) String() string {
	if c.Catch == nil {
		return fmt.Sprintf("CallTaskFn(%v, %v)", c.Target.String(), c.To.String())
	}
	return fmt.Sprintf("CallTaskFn(%v, %v, %v)", c.Target.String(), c.To.String(), c.Catch.String())
}

type ToExpr struct {
	BaseNode
	Body BlockStmt
	Args ArgsExpr
	Next *ToExpr
}

func NewToExpr(body BlockStmt, args ArgsExpr, next *ToExpr) *ToExpr {
	return &ToExpr{
		BaseNode: BaseNode{Type: NodeTypeToExpr},
		Body:     body,
		Args:     args,
		Next:     next,
	}
}

func (t *ToExpr) NodeType() NodeType {
	return t.Type
}

func (t *ToExpr) String() string {
	if t.Next == nil {
		return fmt.Sprintf("ToExpr(%s , %s)", t.Body.String(), t.Args.String())
	}
	return fmt.Sprintf("ToExpr(%s , %s, %s)", t.Body.String(), t.Args.String(), t.Next.String())
}
