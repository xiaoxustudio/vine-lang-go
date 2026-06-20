package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vine-lang/env"
	"vine-lang/runner"
)

var examplesDir = "examples"

func TestExamples(t *testing.T) {
	err := filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".vine") {
			return nil
		}

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

func testVineFile(t *testing.T, filepath_ string) {
	content, err := os.ReadFile(filepath_)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filepath_, err)
	}

	fileDir := filepath.Dir(filepath_)

	wk := env.Workspace{
		Root:     ".",
		BasePath: fileDir,
		FileName: filepath_,
	}

	_, err = runner.ExecuteCode(filepath_, string(content), wk)
	if err != nil {
		t.Errorf("Failed to execute %s: %v", filepath_, err)
	}
}

func BenchmarkExamples(b *testing.B) {
	var files []string
	err := filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".vine") {
			return nil
		}

		if filepath.Dir(path) != examplesDir {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		b.Fatalf("Failed to walk examples directory: %v", err)
	}

	for _, file := range files {
		b.Run(file, func(b *testing.B) {
			benchmarkVineFile(b, file)
		})
	}
}

func benchmarkVineFile(b *testing.B, filepath string) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		b.Fatalf("Failed to read file %s: %v", filepath, err)
	}

	wk := env.Workspace{
		Root:     ".",
		BasePath: ".",
		FileName: filepath,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := runner.ExecuteCode(filepath, string(content), wk)
		if err != nil {
			b.Errorf("Failed to execute %s: %v", filepath, err)
		}
	}
}
