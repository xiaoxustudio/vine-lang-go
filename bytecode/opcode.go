package bytecode

import (
	"encoding/binary"
)

type Opcode byte

const (
	OpConstant Opcode = iota // 添加常量

	OpTrue
	OpFalse

	/* 运算 */
	OpPlus
	OpMinus
	OpMul
	OpDiv
	OpPop // 弹出栈顶元素
	OpJump

	OpIncrement // 自增
	OpDecrement // 自减

	/* 操作 */
	OpSetGlobal
	OpSetConst // 设置常量
	OpGetGlobal
	OpCall
	OpReturn
	OpIndex // 索引
	OpGetMember

	/* 结构 */
	OpArray
)

type Instructions []byte // 单个指令

// 创建指令
func Make(op Opcode, operands ...int) Instructions {
	instructions := make([]byte, 1+len(operands)*2)
	instructions[0] = byte(op)
	for i, operand := range operands {
		offset := 1 + i*2
		//  第一个字节是操作码，后续跟操作数
		binary.LittleEndian.PutUint16(instructions[offset:], uint16(operand))
	}
	return instructions
}

// 读取操作数
func ReadOperands(ins Instructions, offset int) ([]int, int) {
	operands := []int{}
	for offset < len(ins) {
		operand := int(binary.LittleEndian.Uint16(ins[offset:]))
		operands = append(operands, operand)
		offset += 2
	}
	return operands, offset
}
