package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vine-lang/env"
	"vine-lang/ipt"
	"vine-lang/lexer"
	"vine-lang/parser"
)

// TestExamples 测试examples文件夹下的所有.vine文件
func TestExamples(t *testing.T) {
	examplesDir := "examples"

	// 遍历examples文件夹
	err := filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理.vine文件
		if !strings.HasSuffix(path, ".vine") {
			return nil
		}

		// 跳过子目录中的测试文件
		if filepath.Dir(path) != examplesDir {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			testVineFile(t, path)
		})

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk examples directory: %v", err)
	}
}

// TestExamplesRecursive 递归测试examples文件夹下的所有.vine文件
func TestExamplesRecursive(t *testing.T) {
	examplesDir := "examples"

	// 遍历examples文件夹及其子目录
	err := filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理.vine文件
		if !strings.HasSuffix(path, ".vine") {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			testVineFile(t, path)
		})

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk examples directory: %v", err)
	}
}

// testVineFile 测试单个.vine文件
func testVineFile(t *testing.T, filepath string) {
	// 读取文件内容
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filepath, err)
	}

	// 创建工作区
	wk := &env.Workspace{
		Root:     ".",
		BasePath: ".",
		FileName: filepath,
	}

	// 执行代码
	_, err = executeCode(filepath, string(content), *wk)
	if err != nil {
		t.Errorf("Failed to execute %s: %v", filepath, err)
	}
}

// executeCode 执行vine代码
func executeCode(filename string, code string, wk env.Workspace) (any, error) {
	lex := lexer.New(filename, code)
	lex.Parse()

	p := parser.CreateParser(lex)

	e := env.New(wk)
	e.FileName = filename

	i := ipt.New(p, e)

	return i.EvalSafe()
}

// BenchmarkExamples 基准测试examples文件夹下的所有.vine文件
func BenchmarkExamples(b *testing.B) {
	examplesDir := "examples"

	// 收集所有.vine文件
	var files []string
	err := filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理.vine文件
		if !strings.HasSuffix(path, ".vine") {
			return nil
		}

		// 跳过子目录中的测试文件
		if filepath.Dir(path) != examplesDir {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		b.Fatalf("Failed to walk examples directory: %v", err)
	}

	// 对每个文件进行基准测试
	for _, file := range files {
		b.Run(file, func(b *testing.B) {
			benchmarkVineFile(b, file)
		})
	}
}

// benchmarkVineFile 基准测试单个.vine文件
func benchmarkVineFile(b *testing.B, filepath string) {
	// 读取文件内容
	content, err := os.ReadFile(filepath)
	if err != nil {
		b.Fatalf("Failed to read file %s: %v", filepath, err)
	}

	// 创建工作区
	wk := &env.Workspace{
		Root:     ".",
		BasePath: ".",
		FileName: filepath,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executeCode(filepath, string(content), *wk)
		if err != nil {
			b.Errorf("Failed to execute %s: %v", filepath, err)
		}
	}
}
