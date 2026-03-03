package vm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"vine-lang/bytecode"
	"vine-lang/compiler"
	"vine-lang/env"
	"vine-lang/object/store"
	"vine-lang/token"
	"vine-lang/types"
)

func NewVM(c *compiler.Compiler, env *env.Environment) *VM {
	v := &VM{
		constants:  c.GetConstantRaw(),
		stack:      make([]any, 256),
		sp:         0,
		frames:     make([]*Frame, 1),
		frameIndex: 0,
		globals:    make([]any, 256),
		handlers:   make(map[bytecode.Opcode]VMFunc),
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
		frame := v.currentFrame()
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
		frame := v.currentFrame()
		frame.ip += 1
		return false, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpTrue, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) { // 0x03
		v.push(true)
		frame := v.currentFrame()
		frame.ip += 1
		return true, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpPlus, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()

		var result any
		var err error

		// 处理字符串连接
		if leftStr, ok := left.(string); ok {
			if rightStr, ok := right.(string); ok {
				result = leftStr + rightStr
			} else {
				result = fmt.Sprintf("%s%v", leftStr, right)
			}
		} else if rightStr, ok := right.(string); ok {
			result = fmt.Sprintf("%v%s", left, rightStr)
		} else if leftInt, ok := left.(int64); ok {
			if rightInt, ok := right.(int64); ok {
				result = leftInt + rightInt
			} else if rightFloat, ok := right.(float64); ok {
				result = float64(leftInt) + rightFloat
			} else {
				err = errors.New("unsupported types for addition")
			}
		} else if leftFloat, ok := left.(float64); ok {
			if rightInt, ok := right.(int64); ok {
				result = leftFloat + float64(rightInt)
			} else if rightFloat, ok := right.(float64); ok {
				result = leftFloat + rightFloat
			} else {
				err = errors.New("unsupported types for addition")
			}
		} else {
			err = errors.New("unsupported types for addition")
		}

		if err != nil {
			return nil, err
		}

		v.push(result)
		frame := v.currentFrame()
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpMinus, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()

		var result any
		var err error

		if leftInt, ok := left.(int64); ok {
			if rightInt, ok := right.(int64); ok {
				result = leftInt - rightInt
			} else if rightFloat, ok := right.(float64); ok {
				result = float64(leftInt) - rightFloat
			} else {
				err = errors.New("unsupported types for subtraction")
			}
		} else if leftFloat, ok := left.(float64); ok {
			if rightInt, ok := right.(int64); ok {
				result = leftFloat - float64(rightInt)
			} else if rightFloat, ok := right.(float64); ok {
				result = leftFloat - rightFloat
			} else {
				err = errors.New("unsupported types for subtraction")
			}
		} else {
			err = errors.New("unsupported types for subtraction")
		}

		if err != nil {
			return nil, err
		}

		v.push(result)
		frame := v.currentFrame()
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpMul, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()

		var result any
		var err error

		if leftInt, ok := left.(int64); ok {
			if rightInt, ok := right.(int64); ok {
				result = leftInt * rightInt
			} else if rightFloat, ok := right.(float64); ok {
				result = float64(leftInt) * rightFloat
			} else {
				err = errors.New("unsupported types for multiplication")
			}
		} else if leftFloat, ok := left.(float64); ok {
			if rightInt, ok := right.(int64); ok {
				result = leftFloat * float64(rightInt)
			} else if rightFloat, ok := right.(float64); ok {
				result = leftFloat * rightFloat
			} else {
				err = errors.New("unsupported types for multiplication")
			}
		} else {
			err = errors.New("unsupported types for multiplication")
		}

		if err != nil {
			return nil, err
		}

		v.push(result)
		frame := v.currentFrame()
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpDiv, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		right := v.pop()
		left := v.pop()

		var result any
		var err error

		// 检查除数是否为零
		if rightInt, ok := right.(int64); ok {
			if rightInt == 0 {
				return nil, errors.New("division by zero")
			}
		} else if rightFloat, ok := right.(float64); ok {
			if rightFloat == 0 {
				return nil, errors.New("division by zero")
			}
		}

		if leftInt, ok := left.(int64); ok {
			if rightInt, ok := right.(int64); ok {
				if leftInt%rightInt == 0 {
					result = leftInt / rightInt
				} else {
					result = float64(leftInt) / float64(rightInt)
				}
			} else if rightFloat, ok := right.(float64); ok {
				result = float64(leftInt) / rightFloat
			} else {
				err = errors.New("unsupported types for division")
			}
		} else if leftFloat, ok := left.(float64); ok {
			if rightInt, ok := right.(int64); ok {
				result = leftFloat / float64(rightInt)
			} else if rightFloat, ok := right.(float64); ok {
				result = leftFloat / rightFloat
			} else {
				err = errors.New("unsupported types for division")
			}
		} else {
			err = errors.New("unsupported types for division")
		}

		if err != nil {
			return nil, err
		}

		v.push(result)
		frame := v.currentFrame()
		frame.ip += 1
		return result, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpIncrement, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		val := v.pop()
		if valInt, ok := val.(int64); ok {
			switch op {
			case bytecode.OpIncrement:
				v.push(valInt + 1)
			case bytecode.OpDecrement:
				v.push(valInt - 1)
			default:
				return nil, errors.New("unsupported types for increment/decrement")
			}
		} else if valFloat, ok := val.(float64); ok {
			switch op {
			case bytecode.OpIncrement:
				v.push(valFloat + 1)
			case bytecode.OpDecrement:
				v.push(valFloat - 1)
			default:
				return nil, errors.New("unsupported types for increment/decrement")
			}
		} else {
			return nil, errors.New("unsupported types for increment")
		}
		frame := v.currentFrame()
		frame.ip += 1
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpDecrement, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		val := v.pop()
		if valInt, ok := val.(int64); ok {
			v.push(valInt - 1)
		} else if valFloat, ok := val.(float64); ok {
			v.push(valFloat - 1)
		} else {
			return nil, errors.New("unsupported types for decrement")
		}
		frame := v.currentFrame()
		frame.ip += 1
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpPop, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		v.pop()
		frame := v.currentFrame()
		frame.ip += 1
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpJump, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取跳转偏移量（使用Little Endian解码）
		offset := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		frame.ip += offset
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpSetGlobal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取全局变量索引（使用Little Endian解码）
		globalIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		if globalIndex >= len(v.constants) {
			return nil, fmt.Errorf("global index %d out of range", globalIndex)
		}
		// 从栈中弹出值
		value := v.pop()
		// 从常量池中获取变量名
		varName := v.constants[globalIndex].(string)
		// 使用 Define 方法来定义全局变量
		err := v.env.Define(token.Token{Type: token.IDENT, Value: varName}, value)
		if err != nil {
			return nil, err
		}
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpSetConst, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取常量索引（使用Little Endian解码）
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
		frame := v.currentFrame()
		// 从指令中读取全局变量索引（使用Little Endian解码）
		globalIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		if globalIndex >= len(v.constants) {
			return nil, fmt.Errorf("global index %d out of range", globalIndex)
		}
		// 从常量池中获取变量名
		varName := v.constants[globalIndex].(string)
		// 从环境中获取值
		value, ok := v.env.GetFast(varName)
		if !ok {
			return nil, fmt.Errorf("variable %s not found", varName)
		}
		// 压入栈
		v.push(value)
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpGetLocal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取局部变量索引（使用Little Endian解码）
		localIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 从栈中获取局部变量
		// basePointer指向函数对象的位置
		// 参数从basePointer+1开始
		value := v.stack[frame.basePointer+1+localIndex]
		// 压入栈
		v.push(value)
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpSetLocal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取局部变量索引（使用Little Endian解码）
		localIndex := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 从栈中弹出值
		value := v.pop()
		// 设置局部变量
		// basePointer指向函数对象的位置
		// 参数从basePointer+1开始
		v.stack[frame.basePointer+1+localIndex] = value
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpCall, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取参数数量（使用Little Endian解码）
		argCount := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 从栈中弹出函数
		fn := v.stack[v.sp-1-argCount]

		// 检查函数类型
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
			for i := 0; i < argCount; i++ {
				argValue := v.stack[v.sp-argCount+i]
				v.stack[newFrame.basePointer+1+i] = argValue // 参数从函数对象后面开始
			}
			// 清除栈上的函数和参数
			// 设置栈指针到参数之后的位置，避免覆盖参数
			v.sp = newFrame.basePointer + argCount + 1
		case func(...any) (any, error):
			// 处理Go函数调用
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
		}

		frame.ip += 3
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpReturn, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
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

		// 恢复前一帧的栈指针
		prevFrame := v.currentFrame()
		v.sp = prevFrame.basePointer

		// 将返回值压入栈
		if returnValue != nil {
			v.push(returnValue)
		}

		return returnValue, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpGetMember, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取属性名索引（使用Little Endian解码）
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
		var err error

		switch obj := obj.(type) {
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
		return value, err
	})

	v.RegisterOpenCodeHandler(bytecode.OpIndex, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
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
		frame := v.currentFrame()
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
		frame := v.currentFrame()
		// 从指令中读取属性名索引（使用Little Endian解码）
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
		case *store.StoreObject:
			// 设置存储对象的属性
			obj.Define(token.Token{Type: token.IDENT, Value: memberName}, value)
		case map[string]any:
			// 设置map的属性
			obj[memberName] = value
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

	v.RegisterOpenCodeHandler(bytecode.OpArray, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取数组长度（使用Little Endian解码）
		length := int(binary.LittleEndian.Uint16(ins[frame.ip+1:]))
		// 从栈中弹出数组元素
		array := make([]any, length)
		for i := 0; i < length; i++ {
			array[length-i-1] = v.pop()
		}
		// 将数组压入栈
		v.push(array)
		frame.ip += 3
		return array, nil
	})

	return v
}
