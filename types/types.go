package types

type LibsStoreMap = map[string]any
type LibsKeywords string

const (
	Unknown LibsKeywords = "unknown"
	Global  LibsKeywords = "global"
	Time    LibsKeywords = "time"
)

func (k LibsKeywords) IsUnknown() bool {
	return k == Unknown
}

func (k LibsKeywords) IsValidLibsKeyword() bool {
	switch k {
	case Unknown, Global, Time:
		return true
	default:
		return false
	}
}

type GetNumberAndTypeENUM int

const (
	GNT_Unknown GetNumberAndTypeENUM = iota
	GNT_INT
	GNT_FLOAT
	GNT_STRING
)
