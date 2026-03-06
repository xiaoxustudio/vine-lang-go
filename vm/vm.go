package vm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"vine-lang/bytecode"
	"vine-lang/env"
	"vine-lang/token"
	iface "vine-lang/vm/interface"
)

// isTruthy 判断值是否为真
func isTruthy(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	case nil:
		return false
	default:
		// 对于其他类型，假设为真
		return true
	}
}

type VM struct {
	constants []any
	stack     []any
	sp        int // stack pointer

	frames     []*iface.Frame // 函数调用栈
	frameIndex int

	globals  []any             // 全局变量
	locals   []any             // 局部变量
	handlers [256]iface.VMFunc // 使用数组代替map，避免哈希查找开销
	env      *env.Environment  // 环境引用
}

func (v *VM) RegisterOpenCodeHandler(op bytecode.Opcode, handler iface.VMFunc) {
	v.handlers[op] = handler
}

func (v *VM) CallOpenCodeHandler(op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	handler := v.handlers[op]
	if handler == nil {
		return nil, errors.New("unknown opcode")
	}

	return handler(v, op, ins)
}

// GetConstants 获取常量池
func (v *VM) GetConstants() []any {
	return v.constants
}

// GetStack 获取栈
func (v *VM) GetStack() []any {
	return v.stack
}

// GetSP 获取栈指针
func (v *VM) GetSP() int {
	return v.sp
}

// SetSP 设置栈指针
func (v *VM) SetSP(sp int) {
	v.sp = sp
}

// GetFrameIndex 获取帧索引
func (v *VM) GetFrameIndex() int {
	return v.frameIndex
}

// SetFrameIndex 设置帧索引
func (v *VM) SetFrameIndex(index int) {
	v.frameIndex = index
}

// GetFrames 获取帧数组
func (v *VM) GetFrames() []*iface.Frame {
	return v.frames
}

// SetFrames 设置帧数组
func (v *VM) SetFrames(frames []*iface.Frame) {
	v.frames = frames
}

// GetLocals 获取局部变量
func (v *VM) GetLocals() []any {
	return v.locals
}

// SetLocals 设置局部变量
func (v *VM) SetLocals(locals []any) {
	v.locals = locals
}

// GetEnv 获取环境
func (v *VM) GetEnv() *env.Environment {
	return v.env
}

// GetEnvVar 从环境获取变量
func (v *VM) GetEnvVar(name string) (any, bool) {
	nameToken := token.Token{Type: token.IDENT, Value: name}
	return v.env.Get(nameToken)
}

// SetEnvVar 设置环境变量
func (v *VM) SetEnvVar(name string, value any) error {
	nameToken := token.Token{Type: token.IDENT, Value: name}
	v.env.Set(nameToken, value)
	return nil
}

// DefineEnvVar 定义环境变量
func (v *VM) DefineEnvVar(name string, value any) error {
	nameToken := token.Token{Type: token.IDENT, Value: name}
	return v.env.Define(nameToken, value)
}

// DefineEnvConst 定义环境常量
func (v *VM) DefineEnvConst(name string, value any) error {
	nameToken := token.Token{Type: token.IDENT, Value: name}
	v.env.DefineConst(nameToken, value)
	return nil
}

func (v *VM) Run() (any, error) {
	var result any
	for v.frameIndex >= 0 {
		frame := v.frames[v.frameIndex]
		// 检查是否超出指令范围
		if frame.Ip >= len(frame.Fn.Instructions) {
			break
		}
		opcode := bytecode.Opcode(frame.Fn.Instructions[frame.Ip])

		switch opcode {
		case bytecode.OpConstant:
			constIndex := int(binary.LittleEndian.Uint16(frame.Fn.Instructions[frame.Ip+1:]))
			if constIndex >= len(v.constants) {
				return nil, fmt.Errorf("constant index %d out of range", constIndex)
			}
			constant := v.constants[constIndex]
			v.stack[v.sp] = constant
			v.sp++
			frame.Ip += 3
			result = constant
		case bytecode.OpGetLocal:
			localIndex := int(binary.LittleEndian.Uint16(frame.Fn.Instructions[frame.Ip+1:]))
			if localIndex >= len(v.locals) {
				return nil, fmt.Errorf("local index %d out of range", localIndex)
			}
			value := v.locals[localIndex]
			v.stack[v.sp] = value
			v.sp++
			frame.Ip += 3
			result = value
		case bytecode.OpSetLocal:
			localIndex := int(binary.LittleEndian.Uint16(frame.Fn.Instructions[frame.Ip+1:]))
			if localIndex >= len(v.locals) {
				return nil, fmt.Errorf("local index %d out of range", localIndex)
			}
			v.sp--
			v.locals[localIndex] = v.stack[v.sp]
			frame.Ip += 3
		case bytecode.OpPlus, bytecode.OpMinus, bytecode.OpMul, bytecode.OpDiv:
			v.sp -= 2
			right := v.stack[v.sp+1]
			left := v.stack[v.sp]
			var calcResult any
			var err error

			switch opcode {
			case bytecode.OpPlus:
				switch left := left.(type) {
				case int64:
					switch right := right.(type) {
					case int64:
						calcResult = left + right
					case float64:
						calcResult = float64(left) + right
					default:
						err = errors.New("unsupported types for addition")
					}
				case float64:
					switch right := right.(type) {
					case int64:
						calcResult = left + float64(right)
					case float64:
						calcResult = left + right
					default:
						err = errors.New("unsupported types for addition")
					}
				default:
					err = errors.New("unsupported types for addition")
				}
			case bytecode.OpMinus:
				switch left := left.(type) {
				case int64:
					switch right := right.(type) {
					case int64:
						calcResult = left - right
					case float64:
						calcResult = float64(left) - right
					default:
						err = errors.New("unsupported types for subtraction")
					}
				case float64:
					switch right := right.(type) {
					case int64:
						calcResult = left - float64(right)
					case float64:
						calcResult = left - right
					default:
						err = errors.New("unsupported types for subtraction")
					}
				default:
					err = errors.New("unsupported types for subtraction")
				}
			case bytecode.OpMul:
				switch left := left.(type) {
				case int64:
					switch right := right.(type) {
					case int64:
						calcResult = left * right
					case float64:
						calcResult = float64(left) * right
					default:
						err = errors.New("unsupported types for multiplication")
					}
				case float64:
					switch right := right.(type) {
					case int64:
						calcResult = left * float64(right)
					case float64:
						calcResult = left * right
					default:
						err = errors.New("unsupported types for multiplication")
					}
				default:
					err = errors.New("unsupported types for multiplication")
				}
			default: // OpDiv
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
							calcResult = left / right
						} else {
							calcResult = float64(left) / float64(right)
						}
					case float64:
						calcResult = float64(left) / right
					default:
						err = errors.New("unsupported types for division")
					}
				case float64:
					switch right := right.(type) {
					case int64:
						calcResult = left / float64(right)
					case float64:
						calcResult = left / right
					default:
						err = errors.New("unsupported types for division")
					}
				default:
					err = errors.New("unsupported types for division")
				}
			}

			if err != nil {
				return nil, err
			}
			v.stack[v.sp] = calcResult
			v.sp++
			frame.Ip += 1
			result = calcResult
		case bytecode.OpLessThan:
			v.sp -= 2
			right := v.stack[v.sp+1]
			left := v.stack[v.sp]
			var cmpResult bool

			switch left := left.(type) {
			case int64:
				switch right := right.(type) {
				case int64:
					cmpResult = left < right
				case float64:
					cmpResult = float64(left) < right
				}
			case float64:
				switch right := right.(type) {
				case int64:
					cmpResult = left < float64(right)
				case float64:
					cmpResult = left < right
				}
			}

			v.stack[v.sp] = cmpResult
			v.sp++
			frame.Ip += 1
			result = cmpResult
		case bytecode.OpEqual:
			// 弹出栈顶两个值进行比较
			v.sp -= 2
			right := v.stack[v.sp+1]
			left := v.stack[v.sp]
			var cmpResult bool

			// 比较两个值
			switch left := left.(type) {
			case int64:
				switch right := right.(type) {
				case int64:
					cmpResult = left == right
				case float64:
					cmpResult = float64(left) == right
				}
			case float64:
				switch right := right.(type) {
				case int64:
					cmpResult = left == float64(right)
				case float64:
					cmpResult = left == right
				}
			case string:
				if right, ok := right.(string); ok {
					cmpResult = left == right
				}
			}

			// 将比较结果压入栈顶
			v.stack[v.sp] = cmpResult
			v.sp++
			frame.Ip += 1
			result = cmpResult
		case bytecode.OpJumpIfFalse:
			v.sp--
			condition := v.stack[v.sp]
			offset := int(binary.LittleEndian.Uint16(frame.Fn.Instructions[frame.Ip+1:]))
			if !isTruthy(condition) {
				frame.Ip += offset
			} else {
				frame.Ip += 3
			}
		case bytecode.OpJumpIfTrue:
			v.sp--
			condition := v.stack[v.sp]
			offset := int(binary.LittleEndian.Uint16(frame.Fn.Instructions[frame.Ip+1:]))
			if isTruthy(condition) {
				frame.Ip += offset
			} else {
				frame.Ip += 3
			}
		case bytecode.OpJump:
			offset := int(binary.LittleEndian.Uint16(frame.Fn.Instructions[frame.Ip+1:]))
			frame.Ip += offset
		case bytecode.OpLoop:
			offset := int16(binary.LittleEndian.Uint16(frame.Fn.Instructions[frame.Ip+1:]))
			frame.Ip += int(offset)
		case bytecode.OpIncrement:
			v.sp--
			val := v.stack[v.sp]
			switch val := val.(type) {
			case int64:
				v.stack[v.sp] = val + 1
			case float64:
				v.stack[v.sp] = val + 1
			}
			v.sp++
			frame.Ip += 1
		case bytecode.OpPop:
			v.sp--
			frame.Ip += 1
		default:
			handler := v.handlers[opcode]
			if handler == nil {
				return nil, errors.New("unknown opcode")
			}
			r, err := handler(v, opcode, frame.Fn.Instructions)
			if err != nil {
				return nil, err
			}
			result = r
		}
	}
	return result, nil
}

func (v *VM) PushFrame(fn *bytecode.CompiledFunction) {
	v.frameIndex++
	v.frames = append(v.frames, &iface.Frame{Fn: fn, Ip: 0, BasePointer: v.sp})
}

// CurrentFrame 获取当前帧，内联优化
func (v *VM) CurrentFrame() *iface.Frame {
	return v.frames[v.frameIndex]
}

func (v *VM) RunLine(op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	return v.CallOpenCodeHandler(op, ins)
}

func (v *VM) Push(value any) {
	v.stack[v.sp] = value
	v.sp++
}

func (v *VM) Pop() any {
	v.sp--
	value := v.stack[v.sp]
	return value
}
