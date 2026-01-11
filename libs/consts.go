package libs

import (
	"vine-lang/libs/global"
	"vine-lang/libs/time"
	"vine-lang/types"
)

var LibsMap = map[types.LibsKeywords]types.LibsModule{
	types.Global: global.NewModule(),
	types.Time:   time.NewModule(),
}
