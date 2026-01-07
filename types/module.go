package types

type LibsModule interface {
	ID() LibsKeywords
	Name() string
	IsInside() bool
	Return() LibsStoreMap
}

type LibsModuleInterface struct {
	LibsModule
	Path  func() string
	Store LibsStoreMap
}

func (l *LibsModuleInterface) ID() LibsKeywords {
	return Unknown
}

func (l *LibsModuleInterface) Get(key string) any {
	return l.Store[key]
}

func (l *LibsModuleInterface) Set(key string, value any) {
	l.Store[key] = value
}

func (l *LibsModuleInterface) Return() LibsStoreMap {
	return l.Store
}

func (l *LibsModuleInterface) IsInside() bool {
	return true
}

func (l *LibsModuleInterface) Register(key string, fn any) {
	l.Set(key, fn)
}
