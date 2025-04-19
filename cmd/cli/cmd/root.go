package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	fmt_local "toolbox/cmd/cli/cmd/fmt"
	"toolbox/cmd/cli/cmd/fs"
	"toolbox/cmd/cli/cmd/network"
	"toolbox/cmd/cli/cmd/process"
	"toolbox/cmd/cli/cmd/text"
	"toolbox/cmd/cli/cmd/version"

	"github.com/spf13/cobra"
)

// 全局变量存储程序名
var programName string

// rootCmd 表示基础命令
var rootCmd = &cobra.Command{
	Use:   "toolbox",
	Short: "一个功能丰富的命令行工具箱",
	Long:  `Toolbox 是一个集成了多种实用功能的命令行工具箱，具体功能使用 -h 查看`,
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

	// 添加模块
	rootCmd.AddCommand(network.NetworkCmd)
	rootCmd.AddCommand(fmt_local.FmtCmd)
	rootCmd.AddCommand(fs.FsCmd)
	rootCmd.AddCommand(text.TextCmd)
	rootCmd.AddCommand(process.ProcessCmd)
	rootCmd.AddCommand(version.VersionCmd)
}
