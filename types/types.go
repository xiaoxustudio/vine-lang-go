package types

type LibsStoreMap = map[string]any
type LibsKeywords string

const (
	Unknown LibsKeywords = "unknown"
	Global  LibsKeywords = "global"
)

type GetNumberAndTypeENUM int

const (
	GNT_Unknown GetNumberAndTypeENUM = iota
	GNT_INT
	GNT_FLOAT
)
