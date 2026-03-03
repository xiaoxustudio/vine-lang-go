# Vine Lang 作用域实现原理及步骤

## 目录
1. [概述](#概述)
2. [编译器的作用域实现](#编译器的作用域实现)
3. [虚拟机的作用域实现](#虚拟机的作用域实现)
4. [完整示例](#完整示例)

## 概述

Vine Lang 的作用域系统负责管理变量和函数的可见性及生命周期。作用域的实现分为两个部分：

1. **编译器（Compiler）**：负责在编译时构建作用域层次结构，并将变量和函数的定义编译到正确的作用域中
2. **虚拟机（VM）**：负责在运行时维护作用域层次结构，并正确地访问和修改变量

## 编译器的作用域实现

### 1. 作用域数据结构

编译器使用 `CompilerScope` 结构来表示一个作用域：

```go
type CompilerScope struct {
    instructions bytecode.Instructions  // 该作用域编译生成的指令
    constants    []any                 // 该作用域的常量池
    lastOp       bytecode.Opcode       // 最后一条指令的操作码
}
```

编译器维护一个作用域栈：

```go
type Compiler struct {
    scopes      []*CompilerScope  // 作用域栈
    scopeIndex  int               // 当前作用域索引
    constants   []any             // 常量池
    // ... 其他字段
}
```

### 2. 作用域生命周期

#### 2.1 初始化主作用域

```go
func NewCompiler() *Compiler {
    c := &Compiler{
        scopes:     make([]*CompilerScope, 1),
        scopeIndex: 0,
        constants:  make([]any, 0),
    }
    // 初始化主作用域
    c.scopes[0] = &CompilerScope{
        instructions: make([]byte, 0),
        constants:   make([]any, 0),
        lastOp:      bytecode.OpNull,
    }
    return c
}
```

#### 2.2 进入新作用域（编译函数）

当编译函数定义时，创建新的作用域：

```go
func (c *Compiler) enterScope() {
    scope := &CompilerScope{
        instructions: make([]byte, 0),
        constants:   make([]any, 0),
        lastOp:      bytecode.OpNull,
    }
    c.scopes = append(c.scopes, scope)
    c.scopeIndex++
}
```

#### 2.3 退出作用域

当函数编译完成后，退出当前作用域：

```go
func (c *Compiler) leaveScope() *CompilerScope {
    scope := c.scopes[c.scopeIndex]
    c.scopes = c.scopes[:c.scopeIndex]
    c.scopeIndex--
    return scope
}
```

### 3. 编译函数定义

函数定义的编译流程：

```go
func (c *Compiler) CompileFunction(fn *ast.Function) error {
    // 1. 进入新作用域
    c.enterScope()

    // 2. 编译函数参数
    for _, param := range fn.Parameters {
        // 将参数定义为局部变量
        c.defineLocal(param.Value)
    }

    // 3. 编译函数体
    err := c.Compile(fn.Body)
    if err != nil {
        return err
    }

    // 4. 添加返回指令
    c.Emit(bytecode.OpReturn)

    // 5. 退出作用域，获取编译结果
    scope := c.leaveScope()

    // 6. 创建编译后的函数对象
    compiledFn := &bytecode.CompiledFunction{
        Instructions:  scope.instructions,
        NumLocals:     len(scope.locals),
        NumParameters: len(fn.Parameters),
    }

    // 7. 将函数对象添加到常量池
    constIndex := c.addConstant(compiledFn)

    // 8. 生成 OpConstant 指令，将函数对象压入栈
    c.Emit(bytecode.OpConstant, constIndex)

    return nil
}
```

### 4. 指令生成

所有指令都生成到当前作用域中：

```go
func (c *Compiler) Emit(op bytecode.Opcode, operands ...int) int {
    ins := bytecode.Make(op, operands...)
    // 将指令添加到当前作用域的 instructions 中
    currentScope := c.scopes[c.scopeIndex]
    currentScope.instructions = append(currentScope.instructions, ins...)
    return len(currentScope.instructions) - len(ins)
}
```

### 5. 变量定义和访问

#### 5.1 全局变量定义

```go
func (c *Compiler) defineGlobal(name string) error {
    // 1. 将变量名添加到常量池
    constIndex := c.addConstant(name)

    // 2. 生成 OpSetGlobal 指令
    c.Emit(bytecode.OpSetGlobal, constIndex)

    return nil
}
```

#### 5.2 局部变量定义

```go
func (c *Compiler) defineLocal(name string) error {
    // 1. 获取当前作用域
    currentScope := c.scopes[c.scopeIndex]

    // 2. 将变量添加到当前作用域的局部变量列表
    currentScope.locals = append(currentScope.locals, Local{
        Name: name,
        Index: len(currentScope.locals),
    })

    return nil
}
```

#### 5.3 全局变量访问

```go
func (c *Compiler) compileGlobalAccess(name string) error {
    // 1. 将变量名添加到常量池
    constIndex := c.addConstant(name)

    // 2. 生成 OpGetGlobal 指令
    c.Emit(bytecode.OpGetGlobal, constIndex)

    return nil
}
```

#### 5.4 局部变量访问

```go
func (c *Compiler) compileLocalAccess(name string) error {
    // 1. 在当前作用域中查找变量
    currentScope := c.scopes[c.scopeIndex]
    for _, local := range currentScope.locals {
        if local.Name == name {
            // 2. 生成 OpGetLocal 指令
            c.Emit(bytecode.OpGetLocal, local.Index)
            return nil
        }
    }

    // 如果找不到，尝试在父作用域中查找
    return c.compileParentScopeAccess(name)
}
```

### 6. 获取编译结果

```go
func (c *Compiler) Bytecode() *bytecode.CompiledFunction {
    // 只返回主函数（第一个作用域）的指令
    mainScope := c.scopes[0]
    return &bytecode.CompiledFunction{
        Instructions:  mainScope.instructions,
        NumLocals:     0,
        NumParameters: 0,
    }
}
```

## 虚拟机的作用域实现

### 1. 虚拟机数据结构

虚拟机使用 `Frame` 结构来表示一个执行帧：

```go
type Frame struct {
    fn          *bytecode.CompiledFunction  // 要执行的函数
    ip          int                         // 指令指针
    basePointer int                         // 基指针，指向函数对象在栈中的位置
}
```

虚拟机维护一个帧栈：

```go
type VM struct {
    frames     []*Frame  // 帧栈
    frameIndex int       // 当前帧索引
    stack      []any     // 值栈
    sp         int       // 栈指针
    env        *env.Environment  // 全局环境
    constants  []any     // 常量池
    // ... 其他字段
}
```

### 2. 栈结构

虚拟机的栈结构如下：

```
栈底
┌─────────────────────────────────────┐
│  全局变量                            │
│  主函数的局部变量                     │
│  函数对象 (basePointer 指向这里)      │
│  参数 1 (basePointer + 1)            │
│  参数 2 (basePointer + 2)            │
│  ...                                 │
│  局部变量 1                          │
│  局部变量 2                          │
│  ...                                 │
│  临时值 (sp 指向这里)                 │
└─────────────────────────────────────┘
栈顶
```

### 3. 全局变量访问

#### 3.1 OpSetGlobal 指令

```go
v.RegisterOpenCodeHandler(bytecode.OpSetGlobal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
    frame := v.currentFrame()
    // 1. 从指令中读取全局变量索引
    globalIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))

    // 2. 从栈中弹出值
    value := v.pop()

    // 3. 从常量池中获取变量名
    varName := v.constants[globalIndex].(string)

    // 4. 使用 Define 方法来定义全局变量
    err := v.env.Define(token.Token{Type: token.IDENT, Value: varName}, value)
    if err != nil {
        return nil, err
    }

    frame.ip += 3
    return value, nil
})
```

#### 3.2 OpGetGlobal 指令

```go
v.RegisterOpenCodeHandler(bytecode.OpGetGlobal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
    frame := v.currentFrame()
    // 1. 从指令中读取全局变量索引
    globalIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))

    // 2. 从常量池中获取变量名
    varName := v.constants[globalIndex].(string)

    // 3. 从环境中获取值
    value, ok := v.env.GetFast(varName)
    if !ok {
        return nil, fmt.Errorf("variable %s not found", varName)
    }

    // 4. 压入栈
    v.push(value)

    frame.ip += 3
    return value, nil
})
```

### 4. 局部变量访问

#### 4.1 OpGetLocal 指令

```go
v.RegisterOpenCodeHandler(bytecode.OpGetLocal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
    frame := v.currentFrame()
    // 1. 从指令中读取局部变量索引
    localIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))

    // 2. 从栈中获取局部变量
    // basePointer 指向函数对象的位置
    // 参数从 basePointer+1 开始
    value := v.stack[frame.basePointer+1+localIndex]

    // 3. 压入栈
    v.push(value)

    frame.ip += 3
    return value, nil
})
```

#### 4.2 OpSetLocal 指令

```go
v.RegisterOpenCodeHandler(bytecode.OpSetLocal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
    frame := v.currentFrame()
    // 1. 从指令中读取局部变量索引
    localIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))

    // 2. 从栈中弹出值
    value := v.pop()

    // 3. 设置局部变量
    v.stack[frame.basePointer+1+localIndex] = value

    frame.ip += 3
    return value, nil
})
```

### 5. 函数调用

#### 5.1 OpCall 指令

```go
v.RegisterOpenCodeHandler(bytecode.OpCall, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
    frame := v.currentFrame()
    // 1. 从指令中读取参数数量
    argCount := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))

    // 2. 从栈中弹出函数
    fn := v.stack[v.sp-1-argCount]

    // 3. 检查函数类型
    switch fn := fn.(type) {
    case *bytecode.CompiledFunction:
        // 4. 创建新的帧
        newFrame := &Frame{
            fn:          fn,
            ip:          0,
            basePointer: v.sp - argCount - 1, // 指向函数对象的位置
        }
        v.frames = append(v.frames, newFrame)
        v.frameIndex++

        // 5. 将参数从当前帧复制到新帧的栈中
        for i := 0; i < argCount; i++ {
            argValue := v.stack[v.sp-argCount+i]
            v.stack[newFrame.basePointer+1+i] = argValue // 参数从函数对象后面开始
        }

        // 6. 设置栈指针
        v.sp = newFrame.basePointer + argCount + 1

    case func(...any) (any, error):
        // 处理 Go 函数调用
        args := make([]any, argCount)
        for i := 0; i < argCount; i++ {
            args[i] = v.stack[v.sp-argCount+i]
        }
        result, err := fn(args...)
        if err != nil {
            return nil, err
        }

        // 清除栈上的函数和参数
        v.sp -= argCount + 1

        // 将返回值压入栈
        if result != nil {
            v.push(result)
        }

    default:
        // 尝试使用 reflect 调用函数
        // ...
    }

    frame.ip += 3
    return nil, nil
})
```

### 6. 函数返回

#### 6.1 OpReturn 指令

```go
v.RegisterOpenCodeHandler(bytecode.OpReturn, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
    frame := v.currentFrame()

    // 1. 从栈中弹出返回值
    var returnValue any
    if v.sp > frame.basePointer {
        returnValue = v.pop()
    }

    // 2. 弹出当前帧
    v.frames = v.frames[:v.frameIndex]
    v.frameIndex--

    // 3. 恢复栈指针
    v.sp = frame.basePointer

    // 4. 将返回值压入栈
    if returnValue != nil {
        v.push(returnValue)
    }

    return returnValue, nil
})
```

## 完整示例

### 源代码

```vine
fn b_fn(msg) {
    print("你给我的值为：" + msg)
    return msg
}

b_fn("hello world")
```

### 编译过程

1. **编译主函数**
   - 创建主作用域
   - 编译函数定义 `b_fn`
   - 编译函数调用 `b_fn("hello world")`

2. **编译函数定义 `b_fn`**
   - 进入新作用域
   - 编译参数 `msg`
   - 编译函数体
   - 退出作用域，获取编译结果
   - 将函数对象添加到常量池
   - 生成 `OpConstant` 指令，将函数对象压入栈
   - 生成 `OpSetGlobal` 指令，将函数对象设置为全局变量 `b_fn`

3. **编译函数调用 `b_fn("hello world")`**
   - 生成 `OpGetGlobal` 指令，获取全局变量 `b_fn`
   - 生成 `OpConstant` 指令，将字符串 `"hello world"` 压入栈
   - 生成 `OpCall` 指令，调用函数

### 生成的字节码

```
主函数：
0000 OpConstant(2) <CompiledFunction>
0003 OpSetGlobal [3]  // b_fn
0006 OpGetGlobal [4]  // b_fn
0009 OpConstant(5) "hello world"
0012 OpCall [1]

函数 b_fn：
0000 OpGetGlobal [0]  // print
0003 OpConstant(1) "你给我的值为："
0006 OpGetLocal [0]   // msg
0009 OpPlus
0010 OpCall [1]
0013 OpGetLocal [0]   // msg
0016 OpReturn
```

### 执行过程

1. **执行主函数**
   - 执行 `OpConstant`，将函数对象压入栈
   - 执行 `OpSetGlobal`，将函数对象设置为全局变量 `b_fn`
   - 执行 `OpGetGlobal`，获取全局变量 `b_fn`
   - 执行 `OpConstant`，将字符串 `"hello world"` 压入栈
   - 执行 `OpCall`，调用函数

2. **执行函数调用**
   - 创建新的帧
   - 将参数从当前帧复制到新帧的栈中
   - 执行函数体的指令
   - 执行 `OpReturn`，返回结果

3. **输出结果**
   ```
   你给我的值为：hello world
   ```

## 总结

Vine Lang 的作用域系统通过以下方式实现：

1. **编译器**：
   - 使用作用域栈管理嵌套的作用域
   - 每个函数定义创建新的作用域
   - 将变量和函数的定义编译到正确的作用域中
   - 生成正确的字节码指令来访问和修改变量

2. **虚拟机**：
   - 使用帧栈管理函数调用
   - 使用基指针（basePointer）来定位局部变量和参数
   - 使用全局环境来管理全局变量
   - 正确地处理函数调用和返回

这种设计使得 Vine Lang 能够支持嵌套函数、闭包等高级特性，同时保持高效的执行性能。
