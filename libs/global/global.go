package global

import (
	"fmt"
	"vine-lang/object/store"
	"vine-lang/types"
	"vine-lang/utils"
)

type GlobalModule struct {
	types.LibsModuleObject
}

func NewModule() types.LibsModule {
	g := &GlobalModule{
		LibsModuleObject: types.LibsModuleObject{
			Store: *store.NewStoreObject(),
		},
	}
	g.LibsModuleObject.Register("print", Print)
	g.LibsModuleObject.Register("id", Id)
	return g
}

func (g *GlobalModule) ID() types.LibsKeywords {
	return types.Global
}

func (g *GlobalModule) Name() string {
	return "global"
}

/* FN */

// 打印
func Print(env any, rangeArgs ...any) {
	if len(rangeArgs) == 0 {
		return
	}

	for _, arg := range rangeArgs {
		if v, ok := arg.(*store.StoreObject); ok {
			fmt.Print(store.StoreObjectToReadableJSON(v), " ")
			continue
		}
		if v, ok := arg.(*types.FunctionLikeValNode); ok {
			fmt.Print(fmt.Sprintf("<fn %p>", v), " ")
			continue
		}
		if v, ok := arg.(*types.LibsModuleObject); ok {
			fmt.Print(fmt.Sprintf("<module %p>", v), " ")
			continue
		}
		if v, ok := arg.(*types.TaskToValNode); ok {
			fmt.Print(fmt.Sprintf("<task %p>", v), " ")
			continue
		}
		if v, ok := arg.(*types.ErrorValNode); ok {
			fmt.Print(fmt.Sprintf("<error %p>", v), " ")
		}

		fmt.Print(utils.TrasformPrintStringWithColor(arg), " ")
	}
	fmt.Println()
}

func PrintWithColor(env any, rangeArgs ...any) {
	if len(rangeArgs) == 0 {
		return
	}

	for _, arg := range rangeArgs {
		fmt.Print(utils.TrasformPrintStringWithColor(arg), " ")
	}
	fmt.Println()
}

// 获取对象内存地址
func Id(env any, val any) any {
	return fmt.Sprintf("%p", &val)
}
