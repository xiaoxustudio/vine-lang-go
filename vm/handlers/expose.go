package handlers

import (
	"encoding/binary"
	"fmt"
	"vine-lang/bytecode"
	"vine-lang/object/store"
	"vine-lang/token"
	iface "vine-lang/vm/interface"
)

// HandleExpose 处理导出变量
func HandleExpose(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	constants := v.GetConstants()
	env := v.GetEnv()

	// 从指令中读取变量名索引
	nameIndex := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	if nameIndex >= len(constants) {
		return nil, fmt.Errorf("name index %d out of range", nameIndex)
	}

	// 从常量池中获取变量名
	varName := constants[nameIndex].(string)

	// 从环境中获取变量值
	nameToken := token.Token{Type: token.IDENT, Value: varName}
	value, exists := v.GetEnvVar(varName)
	if !exists {
		return nil, fmt.Errorf("variable not found: %s", varName)
	}

	// 确保 Exports 对象存在
	if env.Exports == nil {
		env.Exports = store.NewStoreObject()
	}

	// 将变量添加到导出列表
	err := env.Exports.Define(nameToken, value)
	if err != nil {
		return nil, err
	}

	// 将导出的值压入栈
	v.Push(value)
	frame.Ip += 3
	return value, nil
}
