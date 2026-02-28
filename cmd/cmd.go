package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"vine-lang/env"
	"vine-lang/pprof"
	"vine-lang/repl"
	"vine-lang/utils"

	"github.com/spf13/cobra"
)

const version = "v1.0.1"

var (
	cpuProfile string
	memProfile string
)

var rootCmd = &cobra.Command{
	Use:     "vine [file]",
	Short:   "Vine Language",
	Long:    "Vine Language for xuran",
	Version: version,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 启动CPU性能分析
		if err := pprof.StartCPUProfile(cpuProfile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()

		// 无参数时启动 REPL
		if len(args) == 0 {
			repl.Start()
			return
		}

		// 有参数时执行文件或项目
		RunProjectOrFile(cmd, args)

		// 写入内存性能分析
		if err := pprof.WriteHeapProfile(memProfile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v", err)
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

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a simple vine",
	Long:  `This command will be create a project for Vine Language`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := createProject(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "run a vine script file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RunProjectOrFile(cmd, args)
	},
}

// 执行文件或项目
func RunProjectOrFile(cmd *cobra.Command, args []string) {
	wk, err := GetWorkSpaceWithArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s does not exist\n", args[0])
		os.Exit(1)
	}

	var targetFileName = args[0]

	if info.IsDir() {
		infoM, err := wk.Info()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		targetFileName = filepath.Join(targetFileName, infoM.Main)
	}

	finnal, err := filepath.Abs(targetFileName)

	if err := executeVineFile(finnal, *wk); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// 添加子命令
	rootCmd.AddCommand(replCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(createCmd)

	// 添加pprof标志
	rootCmd.PersistentFlags().StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to file")
	rootCmd.PersistentFlags().StringVar(&memProfile, "memprofile", "", "write memory profile to file")

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
	markers := append([]string{".git"}, utils.ProjectConfigFile...)

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

// 创建项目
func createProject(name string) error {
	// 判断本地是否存在
	if _, err := os.Stat(name); err == nil {
		fmt.Fprintf(os.Stderr, "Error: %s already exists\n", name)
		os.Exit(1)
	}
	// 创建项目
	os.Mkdir(name, os.ModePerm)
	os.Mkdir(filepath.Join(name, "src"), os.ModePerm)
	os.WriteFile(filepath.Join(name, "src", "main.vine"), []byte("print('Hello, World!')"), os.ModePerm)
	os.WriteFile(filepath.Join(name, "vine.project.yml"), []byte(fmt.Sprintf(
		`name: %s
version: %s
main: src/main.vine
author: vine
`, name, version)), os.ModePerm)
	return nil
}
