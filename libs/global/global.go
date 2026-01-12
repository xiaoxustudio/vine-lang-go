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
	g.LibsModuleInterface.Register("print", Print)
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

/* FN */
func Print(env any, rangeArgs ...any) {
	if len(rangeArgs) == 0 {
		return
	}

	for _, arg := range rangeArgs {
		fmt.Print(utils.TrasformPrintStringWithColor(arg))
	}
	fmt.Println()
}
