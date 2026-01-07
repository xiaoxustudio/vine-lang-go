package types

type LibsStoreMap = map[string]any
type LibsKeywords string

const (
	Unknown LibsKeywords = "unknown"
	Global  LibsKeywords = "global"
)
