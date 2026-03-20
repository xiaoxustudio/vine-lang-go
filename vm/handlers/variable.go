package handlers

import (
	"encoding/binary"
	"fmt"
	"vine-lang/bytecode"
	iface "vine-lang/vm/interface"
)

// HandleSetGlobal 处理设置全局变量
func HandleSetGlobal(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	constants := v.GetConstants()

	globalIndex := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	if globalIndex >= len(constants) {
		return nil, fmt.Errorf("global index %d out of range", globalIndex)
	}

	// 从栈中弹出值
	value := v.Pop()
	// 从常量池中获取变量名
	varName := constants[globalIndex].(string)

	// 检查变量是否已存在
	_, exists := v.GetEnvVar(varName)
	if exists {
		// 变量已存在，更新其值
		err := v.SetEnvVar(varName, value)
		if err != nil {
			return nil, err
		}
	} else {
		// 变量不存在，定义它
		err := v.DefineEnvVar(varName, value)
		if err != nil {
			return nil, err
		}
	}
	frame.Ip += 3
	return value, nil
}

// HandleSetConst 处理设置常量
func HandleSetConst(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	constants := v.GetConstants()

	// 从指令中读取常量索引
	constIndex := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	if constIndex >= len(constants) {
		return nil, fmt.Errorf("constant index %d out of range", constIndex)
	}

	// 从栈中弹出值
	value := v.Pop()
	// 从常量池中获取常量名
	constName := constants[constIndex].(string)

	// 定义常量到环境中
	err := v.DefineEnvConst(constName, value)
	if err != nil {
		return nil, err
	}
	frame.Ip += 3
	return value, nil
}

// HandleGetGlobal 处理获取全局变量
func HandleGetGlobal(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	constants := v.GetConstants()

	// 使用当前帧的指令指针来读取操作数
	globalIndex := int(binary.LittleEndian.Uint16(frame.Fn.Instructions[frame.Ip+1:]))
	if globalIndex >= len(constants) {
		return nil, fmt.Errorf("global index %d out of range", globalIndex)
	}

	// 从常量池中获取变量名
	varName := constants[globalIndex].(string)

	// 从环境中获取值
	value, ok := v.GetEnvVar(varName)
	if !ok {
		return nil, fmt.Errorf("variable %s not found", varName)
	}

	// 压入栈
	v.Push(value)
	frame.Ip += 3
	return value, nil
}

// HandleGetLocal 处理获取局部变量
func HandleGetLocal(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	ip := frame.Ip
	localIndex := int(ins[ip+1]) | int(ins[ip+2])<<8
	value := v.LoadLocalToStack(localIndex)
	frame.Ip = ip + 3
	return value, nil
}

// HandleSetLocal 处理设置局部变量
func HandleSetLocal(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	ip := frame.Ip
	localIndex := int(ins[ip+1]) | int(ins[ip+2])<<8
	value := v.StoreLocalFromStack(localIndex)
	frame.Ip = ip + 3
	return value, nil
}
