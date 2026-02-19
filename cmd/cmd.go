package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"vine-lang/env"
	"vine-lang/repl"

	"github.com/spf13/cobra"
)

const version = "v1.0.0"

var rootCmd = &cobra.Command{
	Use:     "vine [file]",
	Short:   "Vine Language",
	Long:    "Vine Language for xuran",
	Version: version,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 无参数时启动 REPL
		if len(args) == 0 {
			repl.Start()
			return
		}

		wk, err := GetWorkSpaceWithArgs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// 执行文件
		filepath := args[0]
		if err := executeVineFile(filepath, *wk); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start interactive REPL",
	Long:  `Launch the interactive REPL environment for Vine Language`,
	Run: func(cmd *cobra.Command, args []string) {
		repl.Start()
	},
}

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "run a vine script file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filepath := args[0]
		wk, err := GetWorkSpaceWithArgs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := executeVineFile(filepath, *wk); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// 添加子命令
	rootCmd.AddCommand(replCmd)
	rootCmd.AddCommand(runCmd)

	// 自定义版本输出
	rootCmd.SetVersionTemplate(`Vine Language {{.Version}} for xuran`)
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// AddCommand 添加自定义命令
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

// findRootPath 从指定目录向上查找标志文件，确定根目录
func findRootPath(startDir string) (string, error) {
	markers := []string{".git", "vine.project"}

	current := startDir
	for {
		// 检查当前目录是否包含标志文件
		for _, marker := range markers {
			path := filepath.Join(current, marker)
			if _, err := os.Stat(path); err == nil {
				return current, nil // 找到了根目录
			}
		}

		// 向上一级
		parent := filepath.Dir(current)

		// 如果已经到达文件系统根目录，停止查找
		if parent == current {
			// 如果找不到标志文件，可以将根目录默认设为 BasePath
			// 或者返回错误
			return startDir, nil
		}
		current = parent
	}
}

// 获取工作区参数
func GetWorkSpaceWithArgs(args []string) (*env.Workspace, error) {
	inputPath := "."
	if len(args) > 0 {
		inputPath = args[0]
	}

	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Invalid path: %v\n", err)
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Path not found: %v\n", err)
		return nil, err
	}

	basePath := absPath
	if !info.IsDir() {
		basePath = filepath.Dir(absPath)
	}

	rootPath, err := findRootPath(basePath)
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Could not determine root path: %v\n", err)
		return nil, err
	}

	ws := &env.Workspace{
		Root:     rootPath,
		BasePath: basePath,
		FileName: filepath.Base(inputPath),
	}

	return ws, nil
}
