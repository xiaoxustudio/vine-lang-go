package bytecode

import "fmt"

type Definition struct {
	Instruction   string // 指令名称
	OperandWidths []int  // 操作数宽度
	Op            Opcode
}

var definitions = map[Opcode]*Definition{
	OpConstant:     {"OpConstant", []int{2}, OpConstant},
	OpPop:          {"OpPop", nil, OpPop},
	OpTrue:         {"OpTrue", nil, OpTrue},
	OpFalse:        {"OpFalse", nil, OpFalse},
	OpPlus:         {"OpPlus", nil, OpPlus},
	OpMinus:        {"OpMinus", nil, OpMinus},
	OpMul:          {"OpMul", nil, OpMul},
	OpDiv:          {"OpDiv", nil, OpDiv},
	OpJump:         {"OpJump", []int{2}, OpJump},
	OpJumpIfFalse:  {"OpJumpIfFalse", []int{2}, OpJumpIfFalse},
	OpJumpIfTrue:   {"OpJumpIfTrue", []int{2}, OpJumpIfTrue},
	OpLoop:         {"OpLoop", []int{2}, OpLoop},
	OpBreak:        {"OpBreak", nil, OpBreak},
	OpContinue:     {"OpContinue", nil, OpContinue},
	OpSetGlobal:    {"OpSetGlobal", []int{2}, OpSetGlobal},
	OpSetConst:     {"OpSetConst", []int{2}, OpSetConst},
	OpGetGlobal:    {"OpGetGlobal", []int{2}, OpGetGlobal},
	OpGetLocal:     {"OpGetLocal", []int{2}, OpGetLocal},
	OpSetLocal:     {"OpSetLocal", []int{2}, OpSetLocal},
	OpCall:         {"OpCall", []int{2}, OpCall},
	OpCallTask:     {"OpCallTask", []int{2}, OpCallTask},
	OpReturn:       {"OpReturn", nil, OpReturn},
	OpIndex:        {"OpIndex", nil, OpIndex},
	OpLen:          {"OpLen", nil, OpLen},
	OpSetIndex:     {"OpSetIndex", nil, OpSetIndex},
	OpGetMember:    {"OpGetMember", []int{2}, OpGetMember},
	OpSetMember:    {"OpSetMember", []int{2}, OpSetMember},
	OpExpose:       {"OpExpose", []int{2}, OpExpose},
	OpIncrement:    {"OpIncrement", nil, OpIncrement},
	OpDecrement:    {"OpDecrement", nil, OpDecrement},
	OpEqual:        {"OpEqual", nil, OpEqual},
	OpNotEqual:     {"OpNotEqual", nil, OpNotEqual},
	OpLessThan:     {"OpLessThan", nil, OpLessThan},
	OpLessEqual:    {"OpLessEqual", nil, OpLessEqual},
	OpGreaterThan:  {"OpGreaterThan", nil, OpGreaterThan},
	OpGreaterEqual: {"OpGreaterEqual", nil, OpGreaterEqual},
	OpArray:        {"OpArray", []int{2}, OpArray},
	OpObject:       {"OpObject", []int{2}, OpObject},
	OpTask:         {"OpTask", []int{2}, OpTask},
	OpTo:           {"OpTo", nil, OpTo},
}

func Lookup(op Opcode, constants []any) (*Definition, error) {
	for _, def := range definitions {
		if def.Op == op {
			return def, nil
		}
	}
	return nil, fmt.Errorf("opcode %d undefined", op)
}

// 将指定的指令集转换为字符串切片
func Disassemble(ins Instructions, constants []any) []string {
	var instructions []string
	i := 0
	for i < len(ins) {
		def, err := Lookup(Opcode(ins[i]), constants)
		if err != nil {
			instructions = append(instructions, fmt.Sprintf("ERROR: %s", err.Error()))
			continue
		}

		// 计算操作数宽度
		operandsWidth := 0
		for _, w := range def.OperandWidths {
			operandsWidth += w
		}

		// 读取操作数
		operands, _ := ReadOperands(ins[i+1:i+1+operandsWidth], 0)

		// 格式化指令
		switch def.Op {
		case OpConstant:
			idx := operands[0]
			if idx < len(constants) {
				instructions = append(instructions, fmt.Sprintf("%04d %s(%d) %v", i, def.Instruction, idx, constants[idx]))
			} else {
				instructions = append(instructions, fmt.Sprintf("%04d %s(%d) [ERROR: constant index out of range]", i, def.Instruction, idx))
			}
		case OpGetGlobal:
			instructions = append(instructions, fmt.Sprintf("%04d %s %v", i, def.Instruction, constants[operands[0]]))
		default:
			if len(operands) > 0 {
				instructions = append(instructions, fmt.Sprintf("%04d %s %v", i, def.Instruction, operands))
			} else {
				instructions = append(instructions, fmt.Sprintf("%04d %s", i, def.Instruction))
			}
		}

		// 移动到下一个指令
		i += 1 + operandsWidth
	}

	return instructions
}
