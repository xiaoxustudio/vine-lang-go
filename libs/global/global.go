package global

import (
	"fmt"
	"vine-lang/types"
	"vine-lang/utils"
)

type GlobalModule struct {
	types.LibsModuleInterface
}

func NewModule() types.LibsModule {
	g := &GlobalModule{
		LibsModuleInterface: types.LibsModuleInterface{
			Store: make(types.LibsStoreMap),
		},
	}
	g.LibsModuleInterface.Register("print", func(env any, rangeArgs []any) {
		for _, arg := range rangeArgs {
			fmt.Print(utils.TrasformPrintString(arg))
		}
		fmt.Println()
	})
	return g
}

func (g *GlobalModule) ID() types.LibsKeywords {
	return types.Global
}

func (g *GlobalModule) Name() string {
	return "global"
}

func (g *GlobalModule) IsInside() bool {
	return true
}
