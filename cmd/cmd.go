package cmd

import (
	"fmt"
	"os"

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

		// 执行文件
		filepath := args[0]
		if err := executeVineFile(filepath); err != nil {
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
		if err := executeVineFile(filepath); err != nil {
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
