package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// 全局变量存储程序名
var programName string

// rootCmd 表示基础命令
var rootCmd = &cobra.Command{
	Use:   "%[1]s",
	Short: "网络诊断工具箱",
	Long: `网络诊断工具箱是一个用Go语言编写的运维工具集合，
提供了多种网络诊断和管理功能，如Ping测试、端口扫描、DNS查询等。

使用方法示例:
  %[1]s ping example.com
  %[1]s portscan example.com --start-port 80 --end-port 100
  %[1]s dns example.com --type mx
  %[1]s traceroute example.com
  %[1]s speedtest
  %[1]s ipinfo 8.8.8.8`,
}

// Execute 将所有子命令添加到根命令并设置标志。
// 这由 main.main() 调用。它只需要对 rootCmd 调用一次。
func Execute() {
	// 获取程序名
	programName = getProgramName()

	// 设置根命令的说明
	rootCmd.Use = fmt.Sprintf(rootCmd.Use, programName)
	rootCmd.Long = fmt.Sprintf(rootCmd.Long, programName)

	// 遍历所有子命令，替换模板变量
	for _, cmd := range rootCmd.Commands() {
		if cmd.Long != "" {
			cmd.Long = fmt.Sprintf(strings.ReplaceAll(cmd.Long, "{{.CommandPath}}", "%[1]s"), programName)
		}
		if cmd.Example != "" {
			cmd.Example = fmt.Sprintf(strings.ReplaceAll(cmd.Example, "{{.CommandPath}}", "%[1]s"), programName)
		}
	}

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// getProgramName 获取程序名
func getProgramName() string {
	// 获取程序完整路径
	path := os.Args[0]
	// 获取文件名（不含路径）
	name := filepath.Base(path)
	// 移除可能的扩展名
	name = strings.TrimSuffix(name, filepath.Ext(name))
	return name
}

func init() {
	// 初始化程序名
	programName = getProgramName()
}
