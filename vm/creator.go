package vm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"vine-lang/bytecode"
	"vine-lang/compiler"
	"vine-lang/env"
	"vine-lang/object/store"
	"vine-lang/token"
	"vine-lang/types"
)

// compareValues 比较两个值，返回比较结果
func compareValues(left, right any, operator string) bool {
	// 快速路径：使用类型开关
	switch left := left.(type) {
	case int64:
		switch right := right.(type) {
		case int64:
			switch operator {
			case "==":
				return left == right
			case "!=":
				return left != right
			case "<":
				return left < right
			case "<=":
				return left <= right
			case ">":
				return left > right
			case ">=":
				return left >= right
			}
		case float64:
			leftFloat := float64(left)
			switch operator {
			case "==":
				return leftFloat == right
			case "!=":
				return leftFloat != right
			case "<":
				return leftFloat < right
			case "<=":
				return leftFloat <= right
			case ">":
				return leftFloat > right
			case ">=":
				return leftFloat >= right
			}
		}
	case float64:
		switch right := right.(type) {
		case int64:
			rightFloat := float64(right)
			switch operator {
			case "==":
				return left == rightFloat
			case "!=":
				return left != rightFloat
			case "<":
				return left < rightFloat
			case "<=":
				return left <= rightFloat
			case ">":
				return left > rightFloat
			case ">=":
				return left >= rightFloat
			}
		case float64:
			switch operator {
			case "==":
				return left == right
			case "!=":
				return left != right
			case "<":
				return left < right
			case "<=":
				return left <= right
			case ">":
				return left > right
			case ">=":
				return left >= right
			}
		}
	case string:
		if right, ok := right.(string); ok {
			switch operator {
			case "==":
				return left == right
			case "!=":
				return left != right
			case "<":
				return left < right
			case "<=":
				return left <= right
			case ">":
				return left > right
			case ">=":
				return left >= right
			}
		}
	case bool:
		if right, ok := right.(bool); ok {
			switch operator {
			case "==":
				return left == right
			case "!=":
				return left != right
			}
		}
	}

	// 使用 reflect 进行通用比较
	switch operator {
	case "==":
		return reflect.DeepEqual(left, right)
	case "!=":
		return !reflect.DeepEqual(left, right)
	default:
		return false
	}
}

func NewVM(c *compiler.Compiler, env *env.Environment) *VM {
	v := &VM{
		constants:  c.GetConstantRaw(),
		stack:      make([]any, 256),
		sp:         0,
		frames:     make([]*Frame, 1),
		frameIndex: 0,
		globals:    make([]any, 256),
		locals:     make([]any, 256),
		handlers:   [256]VMFunc{},
		env:        env,
	}

	// 初始化主帧
	mainFrame := &Frame{
		fn:          c.Bytecode(),
		ip:          0,
		basePointer: 0,
	}
	v.frames[0] = mainFrame

	v.RegisterOpenCodeHandler(bytecode.OpConstant, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) { // 0x01
		frame := v.frames[v.frameIndex]
		// 从指令中读取操作数（常量索引，使用Little Endian解码）
		constIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		if constIndex >= len(v.constants) {
			return nil, fmt.Errorf("constant index %d out of range", constIndex)
		}
		constant := v.constants[constIndex]
		v.push(constant)
		// 跳过操作数
		frame.ip += 3
		return constant, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpFalse, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) { // 0x02
		v.push(false)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return false, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpTrue, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) { // 0x03
		v.push(true)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return true, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpPlus, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()

		var result any
		var err error

		switch left := left.(type) {
		case string:
			if rightStr, ok := right.(string); ok {
				result = left + rightStr
			} else {
				result = fmt.Sprintf("%s%v", left, right)
			}
		case int64:
			switch right := right.(type) {
			case int64:
				result = left + right
			case float64:
				result = float64(left) + right
			default:
				err = errors.New("unsupported types for addition")
			}
		case float64:
			switch right := right.(type) {
			case int64:
				result = left + float64(right)
			case float64:
				result = left + right
			default:
				err = errors.New("unsupported types for addition")
			}
		default:
			err = errors.New("unsupported types for addition")
		}

		if err != nil {
			return nil, err
		}

		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpMinus, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()

		var result any
		var err error

		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case int64:
				result = left - right
			case float64:
				result = float64(left) - right
			default:
				err = errors.New("unsupported types for subtraction")
			}
		case float64:
			switch right := right.(type) {
			case int64:
				result = left - float64(right)
			case float64:
				result = left - right
			default:
				err = errors.New("unsupported types for subtraction")
			}
		default:
			err = errors.New("unsupported types for subtraction")
		}

		if err != nil {
			return nil, err
		}

		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpMul, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()

		var result any
		var err error

		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case int64:
				result = left * right
			case float64:
				result = float64(left) * right
			default:
				err = errors.New("unsupported types for multiplication")
			}
		case float64:
			switch right := right.(type) {
			case int64:
				result = left * float64(right)
			case float64:
				result = left * right
			default:
				err = errors.New("unsupported types for multiplication")
			}
		default:
			err = errors.New("unsupported types for multiplication")
		}

		if err != nil {
			return nil, err
		}

		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpDiv, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()

		var result any
		var err error

		// 检查除数是否为零
		switch right := right.(type) {
		case int64:
			if right == 0 {
				return nil, errors.New("division by zero")
			}
		case float64:
			if right == 0 {
				return nil, errors.New("division by zero")
			}
		}

		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case int64:
				if left%right == 0 {
					result = left / right
				} else {
					result = float64(left) / float64(right)
				}
			case float64:
				result = float64(left) / right
			default:
				err = errors.New("unsupported types for division")
			}
		case float64:
			switch right := right.(type) {
			case int64:
				result = left / float64(right)
			case float64:
				result = left / right
			default:
				err = errors.New("unsupported types for division")
			}
		default:
			err = errors.New("unsupported types for division")
		}

		if err != nil {
			return nil, err
		}

		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpIncrement, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		val := v.pop()
		switch val := val.(type) {
		case int64:
			switch op {
			case bytecode.OpIncrement:
				v.push(val + 1)
			case bytecode.OpDecrement:
				v.push(val - 1)
			default:
				return nil, errors.New("unsupported types for increment/decrement")
			}
		case float64:
			switch op {
			case bytecode.OpIncrement:
				v.push(val + 1)
			case bytecode.OpDecrement:
				v.push(val - 1)
			default:
				return nil, errors.New("unsupported types for increment/decrement")
			}
		default:
			return nil, errors.New("unsupported types for increment")
		}
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpDecrement, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		val := v.pop()
		switch val := val.(type) {
		case int64:
			v.push(val - 1)
		case float64:
			v.push(val - 1)
		default:
			return nil, errors.New("unsupported types for decrement")
		}
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return nil, nil
	})

	// 比较操作处理器
	v.RegisterOpenCodeHandler(bytecode.OpEqual, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()
		result := compareValues(left, right, "==")
		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpNotEqual, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()
		result := compareValues(left, right, "!=")
		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpLessThan, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()
		result := compareValues(left, right, "<")
		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpLessEqual, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()
		result := compareValues(left, right, "<=")
		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpGreaterThan, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()
		result := compareValues(left, right, ">")
		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpGreaterEqual, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()
		result := compareValues(left, right, ">=")
		v.push(result)
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpPop, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		v.pop()
		frame := v.frames[v.frameIndex]
		frame.ip += 1
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpJump, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从指令中读取跳转偏移量
		offset := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		frame.ip += offset
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpJumpIfFalse, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从栈中弹出条件值
		condition := v.pop()
		// 从指令中读取跳转偏移量
		offset := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 如果条件为假，跳转
		if !isTruthy(condition) {
			frame.ip += offset
		} else {
			frame.ip += 3
		}
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpJumpIfTrue, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从栈中弹出条件值
		condition := v.pop()
		// 从指令中读取跳转偏移量
		offset := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 如果条件为真，跳转
		if isTruthy(condition) {
			frame.ip += offset
		} else {
			frame.ip += 3
		}
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpLoop, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从指令中读取跳转偏移量（有符号整数）
		offset := int16(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 跳转到循环开始位置
		frame.ip += int(offset)
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpBreak, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 跳出循环，直接跳转到循环结束位置
		// 这里需要实现循环栈来跟踪循环位置
		// 暂时简单地跳过循环体
		frame.ip += 1
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpContinue, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 跳到下一次迭代，跳转到循环开始位置
		// 这里需要实现循环栈来跟踪循环位置
		// 暂时简单地跳过循环体
		frame.ip += 1
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpSetGlobal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]

		globalIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		if globalIndex >= len(v.constants) {
			return nil, fmt.Errorf("global index %d out of range", globalIndex)
		}
		// 从栈中弹出值
		value := v.pop()
		// 从常量池中获取变量名
		varName := v.constants[globalIndex].(string)
		// 检查变量是否已存在
		nameToken := token.Token{Type: token.IDENT, Value: varName}
		_, exists := v.env.Get(nameToken)
		if exists {
			// 变量已存在，更新其值
			v.env.Set(nameToken, value)
		} else {
			// 变量不存在，定义它
			err := v.env.Define(nameToken, value)
			if err != nil {
				return nil, err
			}
		}
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpSetConst, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从指令中读取常量索引
		constIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		if constIndex >= len(v.constants) {
			return nil, fmt.Errorf("constant index %d out of range", constIndex)
		}
		// 从栈中弹出值
		value := v.pop()
		// 从常量池中获取变量名
		constName := v.constants[constIndex].(string)
		// 定义常量到环境中
		v.env.DefineConst(token.Token{Type: token.IDENT, Value: constName}, value)
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpGetGlobal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]

		globalIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		if globalIndex >= len(v.constants) {
			return nil, fmt.Errorf("global index %d out of range", globalIndex)
		}
		// 从常量池中获取变量名
		varName := v.constants[globalIndex].(string)
		// 从环境中获取值
		value, ok := v.env.GetFast(varName)
		if !ok {
			// 如果在当前环境中找不到，尝试从导入的模块中查找
			nameToken := token.Token{Type: token.IDENT, Value: varName}
			value, ok = v.env.Get(nameToken)
			if !ok {
				return nil, fmt.Errorf("variable %s not found", varName)
			}
		}
		// 压入栈
		v.push(value)
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpGetLocal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]

		localIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 从locals数组中获取局部变量
		if localIndex >= len(v.locals) {
			return nil, fmt.Errorf("local index %d out of range", localIndex)
		}
		value := v.locals[localIndex]
		// 压入栈
		v.push(value)
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpSetLocal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]

		localIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 从栈中弹出值
		value := v.pop()
		// 设置局部变量
		if localIndex >= len(v.locals) {
			return nil, fmt.Errorf("local index %d out of range", localIndex)
		}
		v.locals[localIndex] = value
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpCall, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从指令中读取参数数量
		argCount := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 从栈中弹出函数
		fn := v.stack[v.sp-1-argCount]

		switch fn := fn.(type) {
		case *bytecode.CompiledFunction:
			// 创建新的帧
			// 栈结构: [..., 函数对象, 参数1, 参数2, ...]
			// sp指向参数后的位置
			// 函数对象在 sp-argCount-1
			// 第一个参数在 sp-argCount
			newFrame := &Frame{
				fn:          fn,
				ip:          0,
				basePointer: v.sp - argCount - 1, // 指向函数对象的位置
			}
			v.frames = append(v.frames, newFrame)
			v.frameIndex++
			// 将参数从当前帧复制到新帧的栈中
			// 栈结构: [..., 函数对象, 参数1, 参数2, ...]
			// sp指向参数后的位置
			base := newFrame.basePointer + 1
			copy(v.stack[base:base+argCount], v.stack[v.sp-argCount:v.sp])
			// 调整栈指针，指向参数之后的位置
			// 这样函数执行时可以从这个位置开始使用栈
			v.sp = base + argCount
		case func(...any) (any, error):
			// 处理Go函数调用
			args := make([]any, argCount)
			copy(args, v.stack[v.sp-argCount:v.sp])
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
			// 更新指令指针
			frame.ip += 3
			return result, nil
		default:
			// 尝试使用reflect调用函数
			if reflect.TypeOf(fn).Kind() != reflect.Func {
				return nil, fmt.Errorf("calling non-function: %T , callee : %v", fn, fn)
			}
			args := make([]reflect.Value, argCount)
			for i := 0; i < argCount; i++ {
				args[i] = reflect.ValueOf(v.stack[v.sp-argCount+i])
			}
			results := reflect.ValueOf(fn).Call(args)
			// 先清除栈上的函数和参数
			v.sp -= argCount + 1
			// 再将返回值压入栈
			if len(results) > 0 {
				v.push(results[0].Interface())
			}
			// 更新指令指针
			frame.ip += 3
			return nil, nil
		}

		frame.ip += 3
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpReturn, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 获取返回值
		var returnValue any
		if v.sp > frame.basePointer {
			returnValue = v.pop()
		}

		// 弹出当前帧
		v.frameIndex--
		if v.frameIndex < 0 {
			// 返回到主函数
			return returnValue, nil
		}

		v.sp = frame.basePointer

		// 将返回值压入上一个帧的栈
		if returnValue != nil {
			v.push(returnValue)
		}

		return returnValue, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpGetMember, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从指令中读取属性名索引
		memberIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		if memberIndex >= len(v.constants) {
			return nil, fmt.Errorf("member index %d out of range", memberIndex)
		}
		// 从常量池中获取属性名
		memberName := v.constants[memberIndex].(string)
		// 从栈中弹出对象
		obj := v.pop()

		// 根据对象类型获取属性
		var value any

		switch obj := obj.(type) {
		case map[string]any:
			var ok bool
			value, ok = obj[memberName]
			if !ok {
				return nil, fmt.Errorf("member %s not found in map", memberName)
			}
		case types.LibsModule:
			// 从模块中获取属性
			if val, ok := obj.Get(token.Token{Type: token.IDENT, Value: memberName}); ok {
				value = val
			} else {
				return nil, fmt.Errorf("member %s not found in module", memberName)
			}
		case *store.StoreObject:
			// 从存储对象中获取属性
			if val, ok := obj.Get(token.Token{Type: token.IDENT, Value: memberName}); ok {
				value = val
			} else {
				return nil, fmt.Errorf("member %s not found in object", memberName)
			}
		default:
			// 使用反射获取属性
			r := reflect.ValueOf(obj)
			if r.Kind() == reflect.Pointer {
				r = r.Elem()
			}
			if r.Kind() == reflect.Struct {
				field := r.FieldByName(memberName)
				if field.IsValid() {
					value = field.Interface()
				} else {
					return nil, fmt.Errorf("member %s not found in struct", memberName)
				}
			} else if r.Kind() == reflect.Map {
				field := r.MapIndex(reflect.ValueOf(memberName))
				if field.IsValid() {
					value = field.Interface()
				} else {
					return nil, fmt.Errorf("member %s not found in map", memberName)
				}
			} else {
				return nil, fmt.Errorf("cannot get member from type %T", obj)
			}
		}

		// 将获取的值压入栈
		v.push(value)
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpIndex, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从栈中弹出索引
		index := v.pop()
		// 从栈中弹出数组/对象
		obj := v.pop()

		var value any
		var err error

		switch obj := obj.(type) {
		case []any:
			// 数组索引访问
			idx, ok := index.(int64)
			if !ok {
				return nil, fmt.Errorf("array index must be integer, got %T", index)
			}
			if idx < 0 || int(idx) >= len(obj) {
				return nil, fmt.Errorf("array index %d out of range", idx)
			}
			value = obj[idx]
		case map[string]any:
			// map索引访问
			key, ok := index.(string)
			if !ok {
				return nil, fmt.Errorf("map key must be string, got %T", index)
			}
			var exists bool
			value, exists = obj[key]
			if !exists {
				return nil, fmt.Errorf("key %s not found in map", key)
			}
		default:
			return nil, fmt.Errorf("cannot index type %T", obj)
		}

		// 将获取的值压入栈
		v.push(value)
		frame.ip += 1
		return value, err
	})

	v.RegisterOpenCodeHandler(bytecode.OpSetIndex, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从栈中弹出要设置的值
		value := v.pop()
		// 从栈中弹出索引
		index := v.pop()
		// 从栈中弹出数组/对象
		obj := v.pop()

		switch obj := obj.(type) {
		case []any:
			// 数组索引设置
			idx, ok := index.(int64)
			if !ok {
				return nil, fmt.Errorf("array index must be integer, got %T", index)
			}
			if idx < 0 || int(idx) >= len(obj) {
				return nil, fmt.Errorf("array index %d out of range", idx)
			}
			obj[idx] = value
		case map[string]any:
			// map索引设置
			key, ok := index.(string)
			if !ok {
				return nil, fmt.Errorf("map key must be string, got %T", index)
			}
			obj[key] = value
		default:
			return nil, fmt.Errorf("cannot set index on type %T", obj)
		}

		// 将设置的值压入栈
		v.push(value)
		frame.ip += 1
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpSetMember, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从指令中读取属性名索引
		memberIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		if memberIndex >= len(v.constants) {
			return nil, fmt.Errorf("member index %d out of range", memberIndex)
		}
		// 从常量池中获取属性名
		memberName := v.constants[memberIndex].(string)
		// 从栈中弹出要设置的值
		value := v.pop()
		// 从栈中弹出对象
		obj := v.pop()

		switch obj := obj.(type) {
		case map[string]any:
			// 直接设置map的属性，避免反射
			obj[memberName] = value
		case *store.StoreObject:
			// 设置存储对象的属性
			obj.Define(token.Token{Type: token.IDENT, Value: memberName}, value)
		default:
			// 使用反射设置属性
			r := reflect.ValueOf(obj)
			if r.Kind() == reflect.Pointer {
				r = r.Elem()
			}
			if r.Kind() == reflect.Struct {
				// 结构体字段不可修改，返回错误
				return nil, fmt.Errorf("cannot set member on struct")
			} else if r.Kind() == reflect.Map {
				// 设置map的值
				r.SetMapIndex(reflect.ValueOf(memberName), reflect.ValueOf(value))
			} else {
				return nil, fmt.Errorf("cannot set member on type %T", obj)
			}
		}

		// 将设置的值压入栈
		v.push(value)
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpExpose, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]

		// 从指令中读取变量名索引
		nameIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		if nameIndex >= len(v.constants) {
			return nil, fmt.Errorf("name index %d out of range", nameIndex)
		}

		// 从常量池中获取变量名
		varName := v.constants[nameIndex].(string)

		// 从环境中获取变量值
		nameToken := token.Token{Type: token.IDENT, Value: varName}
		value, exists := v.env.Get(nameToken)
		if !exists {
			return nil, fmt.Errorf("variable not found: %s", varName)
		}

		// 确保 Exports 对象存在
		if v.env.Exports == nil {
			v.env.Exports = store.NewStoreObject()
		}

		// 将变量添加到导出列表
		err := v.env.Exports.Define(nameToken, value)
		if err != nil {
			return nil, err
		}

		// 将导出的值压入栈
		v.push(value)
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpArray, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从指令中读取数组长度
		length := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 从栈中弹出数组元素
		array := make([]any, length)
		for i := length - 1; i >= 0; i-- {
			array[i] = v.pop()
		}
		// 将数组压入栈
		v.push(array)
		frame.ip += 3
		return array, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpObject, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.frames[v.frameIndex]
		// 从指令中读取对象属性数量
		propCount := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 创建对象map
		obj := make(map[string]any)
		// 从栈中弹出属性值和键
		for i := 0; i < propCount; i++ {
			// 弹出值
			value := v.pop()
			// 弹出键
			key := v.pop()
			var keyStr string
			switch k := key.(type) {
			case string:
				keyStr = k
			case int64:
				keyStr = strconv.FormatInt(k, 10)
			case float64:
				keyStr = strconv.FormatFloat(k, 'f', -1, 64)
			default:
				return nil, fmt.Errorf("object key must be string or number, got %T", key)
			}
			// 设置属性
			obj[keyStr] = value
		}
		// 将对象压入栈
		v.push(obj)
		frame.ip += 3
		return obj, nil
	})

	return v
}
