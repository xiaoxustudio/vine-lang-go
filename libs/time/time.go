package time

import (
	"time"
	"vine-lang/types"
)

type TimeModule struct {
	types.LibsModuleInterface
}

func NewModule() types.LibsModule {
	g := &TimeModule{
		LibsModuleInterface: types.LibsModuleInterface{
			Store: make(types.LibsStoreMap),
		},
	}
	g.LibsModuleInterface.Register("now", func(env any) any {
		return time.Now().Unix()
	})
	return g
}

func (g *TimeModule) ID() types.LibsKeywords {
	return types.Time
}

func (g *TimeModule) Name() string {
	return "Time"
}

func (g *TimeModule) IsInside() bool {
	return true
}
