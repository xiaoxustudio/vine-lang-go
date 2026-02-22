package env

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"vine-lang/utils"

	"github.com/goccy/go-yaml"
)

type WorkSpace interface {
	GetBasePath() string
	GetRoot() string
	GetFileName() string
	Cd(path string) error
	IsEmpty() bool
}

type WorkSpaceProjectInfo struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Main    string `yaml:"main"`
	Author  string `yaml:"author"`
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

func (w *Workspace) IsEmpty() bool {
	return w.Root == "" && w.BasePath == ""
}

func (w *Workspace) Info() (*WorkSpaceProjectInfo, error) {
	var op []byte
	for _, p := range utils.ProjectConfigFile {
		infoPath := filepath.Join(w.GetBasePath(), p)
		_op, err := os.ReadFile(infoPath)
		if err != nil {
			continue
		}
		op = _op
	}
	if len(op) == 0 {
		return nil, errors.New("no project info found")
	}
	var info WorkSpaceProjectInfo
	err := yaml.Unmarshal(op, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
