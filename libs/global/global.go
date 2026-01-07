package global

import (
	"fmt"
	"vine-lang/token"
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
	g.LibsModuleInterface.Register("print", func(args ...any) {
		e := args[0].(types.Scope)
		rangeArgs := args[1]
		for _, arg := range rangeArgs.([]any) {
			val, _ := e.Get(arg.(token.Token))
			fmt.Println(utils.TrasformPrintString(val))
		}
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
