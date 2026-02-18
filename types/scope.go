package types

import (
	"vine-lang/token"
)

type Scope interface {
	Get(t token.Token) (any, bool)
	Set(t token.Token, val any)
	Print()
	ForEach(fn func(tk token.Token, val any))
	Define(t token.Token, val any)
}
