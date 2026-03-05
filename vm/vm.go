package vm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"vine-lang/bytecode"
	"vine-lang/env"
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

type VMFunc func(v *VM, op bytecode.Opcode, ins bytecode.Instructions) (any, error)

type VM struct {
	constants []any
	stack     []any
	sp        int // stack pointer

	frames     []*Frame // 函数调用栈
	frameIndex int

	globals  []any            // 全局变量
	locals   []any            // 局部变量
	handlers [256]VMFunc      // 使用数组代替map，避免哈希查找开销
	env      *env.Environment // 环境引用
}

type Frame struct {
	fn          *bytecode.CompiledFunction
	ip          int // instruction pointer
	basePointer int // 栈基址，用于局部变量
}

func (v *VM) RegisterOpenCodeHandler(op bytecode.Opcode, handler VMFunc) {
	v.handlers[op] = handler
}

func (v *VM) CallOpenCodeHandler(op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	handler := v.handlers[op]
	if handler == nil {
		return nil, errors.New("unknown opcode")
	}

	return handler(v, op, ins)
}
func (v *VM) Run() (any, error) {
	var result any
	for v.frameIndex >= 0 {
		frame := v.frames[v.frameIndex]
		// 检查是否超出指令范围
		if frame.ip >= len(frame.fn.Instructions) {
			break
		}
		opcode := bytecode.Opcode(frame.fn.Instructions[frame.ip])

		switch opcode {
		case bytecode.OpConstant:
			constIndex := int(binary.LittleEndian.Uint16(frame.fn.Instructions[frame.ip+1:]))
			if constIndex >= len(v.constants) {
				return nil, fmt.Errorf("constant index %d out of range", constIndex)
			}
			constant := v.constants[constIndex]
			v.stack[v.sp] = constant
			v.sp++
			frame.ip += 3
			result = constant
		case bytecode.OpGetLocal:
			localIndex := int(binary.LittleEndian.Uint16(frame.fn.Instructions[frame.ip+1:]))
			if localIndex >= len(v.locals) {
				return nil, fmt.Errorf("local index %d out of range", localIndex)
			}
			value := v.locals[localIndex]
			v.stack[v.sp] = value
			v.sp++
			frame.ip += 3
			result = value
		case bytecode.OpSetLocal:
			localIndex := int(binary.LittleEndian.Uint16(frame.fn.Instructions[frame.ip+1:]))
			if localIndex >= len(v.locals) {
				return nil, fmt.Errorf("local index %d out of range", localIndex)
			}
			v.sp--
			v.locals[localIndex] = v.stack[v.sp]
			frame.ip += 3
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
			frame.ip += 1
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
			frame.ip += 1
			result = cmpResult
		case bytecode.OpEqual:
			v.sp -= 2
			right := v.stack[v.sp+1]
			left := v.stack[v.sp]
			var cmpResult bool

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

			v.stack[v.sp] = cmpResult
			v.sp++
			frame.ip += 1
			result = cmpResult
		case bytecode.OpJumpIfFalse:
			v.sp--
			condition := v.stack[v.sp]
			offset := int(binary.LittleEndian.Uint16(frame.fn.Instructions[frame.ip+1:]))
			if !isTruthy(condition) {
				frame.ip += offset
			} else {
				frame.ip += 3
			}
		case bytecode.OpJumpIfTrue:
			v.sp--
			condition := v.stack[v.sp]
			offset := int(binary.LittleEndian.Uint16(frame.fn.Instructions[frame.ip+1:]))
			if isTruthy(condition) {
				frame.ip += offset
			} else {
				frame.ip += 3
			}
		case bytecode.OpJump:
			offset := int(binary.LittleEndian.Uint16(frame.fn.Instructions[frame.ip+1:]))
			frame.ip += offset
		case bytecode.OpLoop:
			offset := int16(binary.LittleEndian.Uint16(frame.fn.Instructions[frame.ip+1:]))
			frame.ip += int(offset)
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
			frame.ip += 1
		case bytecode.OpPop:
			v.sp--
			frame.ip += 1
		default:
			handler := v.handlers[opcode]
			if handler == nil {
				return nil, errors.New("unknown opcode")
			}
			r, err := handler(v, opcode, frame.fn.Instructions)
			if err != nil {
				return nil, err
			}
			result = r
		}
	}
	return result, nil
}

func (v *VM) pushFrame(fn *bytecode.CompiledFunction) {
	v.frameIndex++
	v.frames = append(v.frames, &Frame{fn: fn, ip: 0, basePointer: v.sp})
}

// currentFrame 获取当前帧，内联优化
func (v *VM) currentFrame() *Frame {
	return v.frames[v.frameIndex]
}

func (v *VM) RunLine(op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	return v.CallOpenCodeHandler(op, ins)
}

func (v *VM) push(value any) {
	v.stack[v.sp] = value
	v.sp++
}

func (v *VM) pop() any {
	v.sp--
	value := v.stack[v.sp]
	return value
}
