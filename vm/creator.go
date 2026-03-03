package vm

import (
	"errors"
	"fmt"
	"reflect"
	"vine-lang/bytecode"
	"vine-lang/compiler"
	"vine-lang/env"
	"vine-lang/libs/global"
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
		// 从指令中读取操作数（常量索引）
		constIndex := int(ins[frame.ip+1]) | int(ins[frame.ip+2])<<8
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

		if leftInt, ok := left.(int64); ok {
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
		// 从指令中读取跳转偏移量
		offset := int(ins[frame.ip+1]) | int(ins[frame.ip+2])<<8
		frame.ip += offset
		return nil, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpSetGlobal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取全局变量索引
		globalIndex := int(ins[frame.ip+1]) | int(ins[frame.ip+2])<<8
		if globalIndex >= len(v.constants) {
			return nil, fmt.Errorf("global index %d out of range", globalIndex)
		}
		// 从栈中弹出值
		value := v.pop()
		// 从常量池中获取变量名
		varName := v.constants[globalIndex].(string)
		// 设置到环境中
		v.env.SetFast(varName, value)
		frame.ip += 3
		return value, nil
	})

	v.RegisterOpenCodeHandler(bytecode.OpGetGlobal, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取全局变量索引
		globalIndex := int(ins[frame.ip+1]) | int(ins[frame.ip+2])<<8
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

	v.RegisterOpenCodeHandler(bytecode.OpCall, func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
		frame := v.currentFrame()
		// 从指令中读取参数数量
		argCount := int(ins[frame.ip+1]) | int(ins[frame.ip+2])<<8
		// 从栈中弹出函数
		fn := v.stack[v.sp-1-argCount]

		// 检查函数类型
		switch fn := fn.(type) {
		case *bytecode.CompiledFunction:
			// 创建新的帧
			newFrame := &Frame{
				fn:          fn,
				ip:          0,
				basePointer: v.sp - argCount,
			}
			v.pushFrame(fn)
			// 将参数从当前帧复制到新帧的栈中
			for i := 0; i < argCount; i++ {
				v.stack[newFrame.basePointer+i] = v.stack[v.sp-argCount+i]
			}
			// 清除栈上的函数和参数
			v.sp = newFrame.basePointer
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
				return nil, fmt.Errorf("calling non-function: %T", fn)
			}
			args := make([]reflect.Value, argCount)
			for i := 0; i < argCount; i++ {
				args[i] = reflect.ValueOf(v.stack[v.sp-argCount+i])
			}
			results := reflect.ValueOf(fn).Call(args)
			if len(results) > 0 {
				v.push(results[0].Interface())
			}
			v.sp -= argCount + 1 // 清除栈上的函数和参数
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
		// 从指令中读取属性名索引
		memberIndex := int(ins[frame.ip+1]) | int(ins[frame.ip+2])<<8
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

		var data types.LibsModule = global.NewModule()
		data.Get(token.Token{Type: token.IDENT, Value: "test"})

		switch obj := obj.(type) {
		case *store.StoreObject:
			// 从存储对象中获取属性
			if val, ok := obj.Get(token.Token{Type: token.IDENT, Value: memberName}); ok {
				value = val
			} else {
				return nil, fmt.Errorf("member %s not found in object", memberName)
			}
		default:
			if obj, ok := obj.(types.LibsModule); ok {
				// 从模块中获取属性
				if val, ok := obj.Get(token.Token{Type: token.IDENT, Value: memberName}); ok {
					value = val
				} else {
					return nil, fmt.Errorf("member %s not found in module", memberName)
				}
			} else {
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
		}

		// 将获取的值压入栈
		v.push(value)
		frame.ip += 3
		return value, err
	})
	return v
}
