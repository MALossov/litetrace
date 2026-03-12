// Package cmd 实现 litetrace 的命令行接口
// 使用 cobra 框架构建子命令系统
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// debugFlag 全局调试标志，所有子命令共享
	debugFlag bool
)

// rootCmd 是 litetrace 的根命令
var rootCmd = &cobra.Command{
	Use:   "litetrace",
	Short: "A lightweight ftrace tool written in Go",
	Long:  `Litetrace is a pure Go implementation of a trace-cmd like tool for Linux ftrace.`,
}

// Execute 执行根命令，由 main.go 调用
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 添加全局 --debug 标志
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug mode (show tracefs file operations)")

	// 注册所有子命令
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(wizardCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(tldrCmd)
	rootCmd.AddCommand(terminateCmd)
}

// checkRoot 检查是否以 root 身份运行
func checkRoot() bool {
	return os.Geteuid() == 0
}
