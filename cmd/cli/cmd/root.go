package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	fmt_local "tuleaj_tools/tool-box/cmd/cli/cmd/fmt"
	"tuleaj_tools/tool-box/cmd/cli/cmd/fs"
	"tuleaj_tools/tool-box/cmd/cli/cmd/network"
	"tuleaj_tools/tool-box/cmd/cli/cmd/text"

	"github.com/spf13/cobra"
)

// 全局变量存储程序名
var programName string

// rootCmd 表示基础命令
var rootCmd = &cobra.Command{
	Use:   "%[1]s",
	Short: "网络诊断和数据处理工具箱",
	Long: `网络诊断和数据处理工具箱是一个用Go语言编写的运维工具集合，
提供了多种网络诊断、数据格式化和管理功能。

使用方法示例:
  %[1]s network ping example.com
  %[1]s network portscan example.com --start-port 80 --end-port 100
  %[1]s network dns example.com --type mx
  %[1]s network traceroute example.com
  %[1]s network speedtest
  %[1]s network ipinfo 8.8.8.8
  %[1]s network sniff eth0 --filter "tcp and port 80"
  %[1]s network sniff --list-interfaces
  %[1]s fmt data.json --pretty
  %[1]s fmt config.xml --pretty --output formatted.xml
  %[1]s text grep "error" log.txt
  %[1]s text replace "old" "new" file.txt
  %[1]s text filter '$1 > 100' data.txt
  %[1]s fs tree /path/to/dir -d 2 -a`,
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
}
