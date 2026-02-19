package env

import (
	"errors"
	"path/filepath"
	"strings"
)

type WorkSpace interface {
	GetBasePath() string
	GetRoot() string
	GetFileName() string
	Cd(path string) error
	isEmpty() bool
}

type Workspace struct {
	Root     string
	BasePath string
	FileName string
}

func NewWorkspace(root, basePath string) WorkSpace {
	cleanRoot := filepath.Clean(root)
	if basePath == "" {
		basePath = root
	}
	cleanBasePath := filepath.Clean(basePath)

	return &Workspace{
		Root:     cleanRoot,
		BasePath: cleanBasePath,
	}
}

func (w *Workspace) GetBasePath() string {
	return w.BasePath
}

func (w *Workspace) GetRoot() string {
	return w.Root
}

func (w *Workspace) GetFileName() string {
	return w.FileName
}

func (w *Workspace) Cd(path string) error {
	if path == "" {
		return nil
	}

	var targetPath string
	// 根据前缀决定路径解析方式
	switch {
	case strings.HasPrefix(path, "@"):
		relPath := strings.TrimPrefix(path, "@")
		targetPath = filepath.Join(w.Root, relPath)
	case filepath.IsAbs(path):
		targetPath = path
	default:
		targetPath = filepath.Join(w.BasePath, path)
	}

	targetPath = filepath.Clean(targetPath)

	rel, err := filepath.Rel(w.Root, targetPath)
	if err != nil {
		return err
	}

	// 检查路径是否超出根目录
	if strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "../") {
		return errors.New("permission denied: path escapes root directory")
	}

	w.BasePath = targetPath
	return nil
}

func (w *Workspace) isEmpty() bool {
	return w.Root == "" && w.BasePath == ""
}
