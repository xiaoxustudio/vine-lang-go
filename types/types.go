package types

type LibsKeywords string

const (
	Unknown LibsKeywords = "unknown"
	Global  LibsKeywords = "glb"
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
