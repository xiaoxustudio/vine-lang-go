# Vine 语言作用域实现原理

## 目录
1. [概述](#概述)
2. [作用域数据结构](#作用域数据结构)
3. [作用域生命周期](#作用域生命周期)
4. [变量解析机制](#变量解析机制)
5. [函数作用域](#函数作用域)
6. [块作用域](#块作用域)
7. [指令传递机制](#指令传递机制)
8. [示例分析](#示例分析)

## 概述

Vine 语言采用静态作用域（词法作用域）机制，通过编译时的作用域栈来管理变量的可见性和生命周期。作用域系统是编译器的核心组件之一，负责：

- 管理变量的声明和访问
- 控制指令的生成和作用域
- 支持嵌套作用域和闭包
- 区分局部变量和全局变量

## 作用域数据结构

### CompilationScope 结构

```go
type CompilationScope struct {
    instructions bytecode.Instructions  // 作用域内的字节码指令
    lastIns      EmittedInstruction     // 最后一条发射的指令
    env          env.Environment         // 运行时环境
    symbolTable  map[string]int         // 局部变量符号表
    parent       *CompilationScope       // 父作用域指针
}
```

**字段说明：**

1. **instructions**: 存储该作用域内生成的所有字节码指令
2. **lastIns**: 记录最后一条发射的指令，用于回填跳转地址
3. **env**: 运行时环境，用于存储全局变量和函数
4. **symbolTable**: 局部变量符号表，映射变量名到索引
5. **parent**: 指向父作用域的指针，形成作用域链

### Compiler 结构中的作用域管理

```go
type Compiler struct {
    handlers     map[ast.NodeType]compileFunc
    constants    []any
    instructions []bytecode.Instructions

    // 作用域 / 符号表
    symbolTable *store.StoreObject
    scopes      []*CompilationScope  // 作用域栈
    scopeIndex  int                 // 当前作用域索引
}
```

**关键点：**

- `scopes` 是一个切片，用作作用域栈
- `scopeIndex` 指向当前活动的作用域
- 通过 `scopeIndex` 可以访问当前作用域及其所有父作用域

## 作用域生命周期

### 1. EnterScope - 进入作用域

```go
func (c *Compiler) EnterScope(e env.Environment) *CompilationScope {
    scope := &CompilationScope{
        instructions: bytecode.Instructions{},
        env:          e,
        symbolTable:  make(map[string]int),
    }

    // 如果有父作用域，设置父作用域
    if c.scopeIndex >= 0 {
        scope.parent = c.scopes[c.scopeIndex]
    }

    // 先将作用域添加到数组中
    c.scopes = append(c.scopes, scope)
    // 再更新索引，指向新添加的作用域
    c.scopeIndex = len(c.scopes) - 1
    return scope
}
```

**执行步骤：**

1. 创建新的 `CompilationScope` 对象
2. 初始化空的指令集和符号表
3. 设置父作用域指针（如果存在）
4. 将新作用域压入作用域栈
5. 更新 `scopeIndex` 指向新作用域

**调用时机：**
- 函数声明时
- 块语句（BlockStmt）开始时

### 2. LeaveScope - 退出作用域

```go
func (c *Compiler) LeaveScope() *CompilationScope {
    if c.scopeIndex < 0 {
        return nil
    }

    scope := c.scopes[c.scopeIndex]
    c.scopes = c.scopes[:c.scopeIndex]
    c.scopeIndex--
    return scope
}
```

**执行步骤：**

1. 检查是否有活动作用域
2. 获取当前作用域引用
3. 从作用域栈中移除当前作用域
4. 递减 `scopeIndex`
5. 返回被移除的作用域对象

**调用时机：**
- 函数体编译完成时
- 块语句编译完成时

## 变量解析机制

### DefineLocal - 定义局部变量

```go
func (c *Compiler) DefineLocal(name string) int {
    currentScope := c.scopes[c.scopeIndex]
    index := len(currentScope.symbolTable)
    currentScope.symbolTable[name] = index
    return index
}
```

**工作原理：**

1. 获取当前作用域
2. 计算新变量的索引（当前符号表长度）
3. 将变量名和索引存入符号表
4. 返回变量索引

**使用场景：**
- 函数参数声明
- 函数内局部变量声明

### ResolveVariable - 解析变量

```go
func (c *Compiler) ResolveVariable(name string) (isLocal bool, index int) {
    // 从当前作用域开始，向上查找
    for i := c.scopeIndex; i >= 0; i-- {
        scope := c.scopes[i]
        if idx, ok := scope.symbolTable[name]; ok {
            // 如果在当前作用域或其父作用域中找到，则是局部变量
            // 对于函数作用域，所有参数和局部变量都是局部的
            return true, idx
        }
    }

    // 没有在任何作用域中找到，是全局变量
    return false, -1
}
```

**查找策略：**

1. 从当前作用域开始向上查找
2. 遍历作用域链（从内到外）
3. 在第一个找到变量的作用域返回
4. 如果未找到，标记为全局变量

**作用域链示例：**

```
全局作用域 (scopeIndex=0)
  └── 函数作用域 (scopeIndex=1)
        └── 块作用域 (scopeIndex=2)
```

当在块作用域中解析变量时：
1. 先在块作用域 (scopeIndex=2) 中查找
2. 未找到则在函数作用域 (scopeIndex=1) 中查找
3. 未找到则在全局作用域 (scopeIndex=0) 中查找
4. 仍未找到则标记为全局变量

## 函数作用域

### 函数声明处理流程

```go
c.RegisterStmtHandler(ast.NodeTypeFunctionDecl, func(node ast.Node) (any, error) {
    n := node.(*ast.FunctionDecl)

    // 1. 进入新的编译作用域用于函数体
    currentEnv := c.scopes[c.scopeIndex].env
    c.EnterScope(currentEnv)

    // 2. 获取函数作用域
    funcScope := c.CurrentScope()

    // 3. 处理函数参数，将参数名添加到符号表
    for _, arg := range n.Arguments.Arguments {
        if lit, ok := arg.(*ast.Literal); ok && lit.Value.Type == token.IDENT {
            paramName := lit.Value.Value
            c.DefineLocal(paramName)
        }
    }

    // 4. 编译函数体
    var allInstructions []byte
    beforeCompileScopeIndex := c.scopeIndex
    _, err := c.Compile(n.Body)
    if err != nil {
        return nil, err
    }

    // 5. 收集函数体的所有指令
    if c.scopeIndex > beforeCompileScopeIndex {
        currentScope := c.CurrentScope()
        if len(currentScope.instructions) > 0 {
            allInstructions = append(allInstructions, currentScope.instructions...)
        }
    } else {
        currentScope := c.CurrentScope()
        allInstructions = currentScope.instructions
    }

    // 6. 退出所有嵌套的作用域
    for c.scopeIndex >= 0 && c.scopes[c.scopeIndex] != funcScope.parent {
        c.LeaveScope()
    }

    // 7. 创建函数对象
    fn := &bytecode.CompiledFunction{
        Instructions:  allInstructions,
        NumLocals:     len(funcScope.symbolTable),
        NumParameters: len(n.Arguments.Arguments),
    }

    // 8. 将函数对象添加到常量池
    fnIndex := c.AddConstant(fn)

    // 9. 生成函数定义指令
    c.Emit(bytecode.OpConstant, fnIndex)

    // 10. 将函数名添加到全局符号表
    c.symbolTable.Set(n.Name.Value, store.NewStoreObject(fn, store.StoreTypeFunction))

    return nil, nil
})
```

**详细步骤：**

1. **创建函数作用域**：
   - 使用当前环境创建新作用域
   - 建立父作用域链接

2. **注册函数参数**：
   - 遍历参数列表
   - 将每个参数名添加到函数作用域的符号表

3. **编译函数体**：
   - 保存编译前的 `scopeIndex`
   - 递归编译函数体（可能包含块语句）
   - 处理嵌套作用域的指令

4. **收集指令**：
   - 从当前作用域收集所有指令
   - 处理可能的嵌套作用域

5. **退出作用域**：
   - 退出函数作用域及其所有嵌套作用域
   - 返回到函数声明之前的作用域

6. **创建函数对象**：
   - 封装指令、局部变量数、参数数
   - 添加到编译器的常量池

7. **注册函数**：
   - 将函数对象存储到全局符号表
   - 生成相应的字节码指令

## 块作用域

### 块语句处理流程

```go
c.RegisterStmtHandler(ast.NodeTypeBlockStmt, func(node ast.Node) (any, error) {
    n := node.(*ast.BlockStmt)

    // 1. 进入新的作用域
    currentEnv := c.scopes[c.scopeIndex].env
    c.EnterScope(currentEnv)

    // 2. 编译块内的所有语句
    for _, s := range n.Body {
        if _, ok := s.(*ast.CommentStmt); ok {
            continue
        }
        _, err := c.Compile(s)
        if err != nil {
            return nil, err
        }
    }

    // 3. 退出作用域，并保存指令
    scope := c.LeaveScope()

    // 4. 将指令传递给父作用域
    if scope != nil && len(scope.instructions) > 0 {
        parentScope := c.CurrentScope()
        if parentScope != nil {
            parentScope.instructions = append(parentScope.instructions, scope.instructions...)
        }
    }

    return nil, nil
})
```

**关键点：**

1. **独立作用域**：
   - 每个块语句创建独立的作用域
   - 使用父作用域的环境

2. **指令传递**：
   - 块作用域的指令在退出时传递给父作用域
   - 确保指令不会丢失

3. **变量隔离**：
   - 块内定义的变量只在块内可见
   - 通过符号表隔离

## 指令传递机制

### 问题背景

在早期实现中，块作用域的指令会随着作用域的退出而丢失，导致函数体的指令不完整。这是因为 `LeaveScope` 会从作用域栈中移除作用域，导致指令无法访问。

### 解决方案

**块语句中的指令传递：**

```go
// 退出作用域，并保存指令
scope := c.LeaveScope()

// 将指令传递给父作用域
if scope != nil && len(scope.instructions) > 0 {
    parentScope := c.CurrentScope()
    if parentScope != nil {
        parentScope.instructions = append(parentScope.instructions, scope.instructions...)
    }
}
```

**工作原理：**

1. `LeaveScope` 返回被移除的作用域对象
2. 检查作用域是否有指令
3. 将指令追加到父作用域的指令集
4. 确保指令不会丢失

**指令流示例：**

```
函数作用域 (scopeIndex=1)
  ├── EnterScope
  ├── 编译参数
  ├── 编译函数体
  │     └── 块作用域 (scopeIndex=2)
  │           ├── EnterScope
  │           ├── 生成指令 A
  │           ├── LeaveScope (返回指令 A)
  │           └── 指令 A 传递给函数作用域
  ├── 收集所有指令 (包括 A)
  ├── LeaveScope
  └── 创建函数对象 (包含指令 A)
```

## 示例分析

### 示例代码

```vine
use glb pick print

fn b_fn(val):
    print("你给我的值为："+val)
    return val
end

b_fn("hello world")
```

### 编译过程分析

**1. 初始状态**
```
scopeIndex=0 (全局作用域)
scopes=[全局作用域]
```

**2. 编译函数声明 `fn b_fn(val):`**
```
EnterScope: scopeIndex=1
scopes=[全局作用域, 函数作用域]
```

- 创建函数作用域
- 注册参数 `val` 到符号表

**3. 编译函数体**
```
EnterScope: scopeIndex=2
scopes=[全局作用域, 函数作用域, 块作用域]
```

- 创建块作用域
- 生成函数体指令：
  - OpConstant (获取 "你给我的值为：")
  - OpGetLocal (获取 val)
  - OpAdd (拼接字符串)
  - OpCall (调用 print)
  - OpGetLocal (获取 val)
  - OpReturnValue (返回 val)

**4. 退出块作用域**
```
LeaveScope: scopeIndex=2
scopes=[全局作用域, 函数作用域]
```

- 块作用域指令传递给函数作用域
- 函数作用域现在包含所有函数体指令

**5. 退出函数作用域**
```
LeaveScope: scopeIndex=1
scopes=[全局作用域]
```

- 创建函数对象，包含所有指令
- 函数对象添加到常量池
- 函数名注册到全局符号表

**6. 编译函数调用 `b_fn("hello world")`**
```
scopeIndex=0 (全局作用域)
```

- 生成调用指令：
  - OpConstant (获取 "hello world")
  - OpGetGlobal (获取 b_fn)
  - OpCall (调用函数)

### 最终字节码

```
0000 0000 OpConstant(2) &{[13 0 0 1 0 14 0 0 3 16 1 0 14 0 0 17] 1 1}
0003 0000 OpSetGlobal [3]
0006 0000 OpGetGlobal b_fn
0009 0000 OpConstant(5) hello world
0012 0000 OpCall [1]
```

**指令说明：**

1. `OpConstant(2)`: 加载函数对象（包含函数体指令）
2. `OpSetGlobal [3]`: 将函数存储到全局变量索引 3
3. `OpGetGlobal b_fn`: 获取全局函数 b_fn
4. `OpConstant(5)`: 加载字符串 "hello world"
5. `OpCall [1]`: 调用函数，参数个数为 1

### 执行流程

1. **函数定义**：
   - 函数对象存储在全局变量中
   - 函数体指令包含在函数对象中

2. **函数调用**：
   - 加载函数对象
   - 加载参数 "hello world"
   - 创建新的执行帧
   - 执行函数体指令
   - 返回结果

## 总结

Vine 语言的作用域实现具有以下特点：

1. **静态作用域**：变量在编译时解析，作用域链在编译时确定
2. **作用域栈**：使用栈结构管理嵌套作用域
3. **指令隔离**：每个作用域维护独立的指令集
4. **指令传递**：子作用域的指令传递给父作用域
5. **变量解析**：从内到外查找变量，支持闭包

这种设计确保了：
- 变量的正确作用域
- 指令的完整性
- 高效的变量访问
- 良好的封装性
