package time

import (
	"time"
	"vine-lang/object/store"
	"vine-lang/types"
)

type TimeModule struct {
	types.LibsModuleObject
}

func NewModule() types.LibsModule {
	g := &TimeModule{
		LibsModuleObject: types.LibsModuleObject{
			Store: *store.NewStoreObject(),
		},
	}
	g.LibsModuleObject.Register("Now", Now)
	g.LibsModuleObject.Register("Milli", Milli)
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

/* FN */
func Now(env any) any {
	return time.Now().Unix()
}
func Milli(env any) any {
	return time.Now().UnixMilli()
}
