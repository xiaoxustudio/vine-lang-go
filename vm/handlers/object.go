package handlers

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"vine-lang/bytecode"
	"vine-lang/object/store"
	"vine-lang/token"
	"vine-lang/types"
	iface "vine-lang/vm/interface"
)

// HandleGetMember 处理获取对象成员
func HandleGetMember(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	constants := v.GetConstants()

	// 从指令中读取属性名索引
	memberIndex := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	if memberIndex >= len(constants) {
		return nil, fmt.Errorf("member index %d out of range", memberIndex)
	}
	// 从常量池中获取属性名
	memberName := constants[memberIndex].(string)
	// 从栈中弹出对象
	obj := v.Pop()

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
	v.Push(value)
	frame.Ip += 3
	return value, nil
}

// HandleSetMember 处理设置对象成员
func HandleSetMember(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()
	constants := v.GetConstants()

	// 从指令中读取属性名索引
	memberIndex := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	if memberIndex >= len(constants) {
		return nil, fmt.Errorf("member index %d out of range", memberIndex)
	}
	// 从常量池中获取属性名
	memberName := constants[memberIndex].(string)
	// 从栈中弹出要设置的值
	value := v.Pop()
	// 从栈中弹出对象
	obj := v.Pop()

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
	v.Push(value)
	frame.Ip += 3
	return value, nil
}

// HandleIndex 处理索引访问
func HandleIndex(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()

	// 从栈中弹出索引
	index := v.Pop()
	// 从栈中弹出数组/对象
	obj := v.Pop()

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
	v.Push(value)
	frame.Ip += 1
	return value, err
}

// HandleSetIndex 处理设置索引
func HandleSetIndex(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()

	// 从栈中弹出要设置的值
	value := v.Pop()
	// 从栈中弹出索引
	index := v.Pop()
	// 从栈中弹出数组/对象
	obj := v.Pop()

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
	v.Push(value)
	frame.Ip += 1
	return value, nil
}

// HandleArray 处理创建数组
func HandleArray(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()

	// 从指令中读取数组长度
	length := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	// 从栈中弹出数组元素
	array := make([]any, length)
	for i := length - 1; i >= 0; i-- {
		array[i] = v.Pop()
	}
	// 将数组压入栈
	v.Push(array)
	frame.Ip += 3
	return array, nil
}

// HandleObject 处理创建对象
func HandleObject(v iface.VMInterface, op bytecode.Opcode, ins bytecode.Instructions) (any, error) {
	frame := v.CurrentFrame()

	// 从指令中读取对象属性数量
	propCount := int(binary.LittleEndian.Uint16(ins[frame.Ip+1:]))
	// 创建对象map
	obj := make(map[string]any)
	// 从栈中弹出属性值和键
	for i := 0; i < propCount; i++ {
		// 弹出值
		value := v.Pop()
		// 弹出键
		key := v.Pop()
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
	v.Push(obj)
	frame.Ip += 3
	return obj, nil
}
