package process

import (
	"github.com/spf13/cobra"
)

// ProcessCmd 表示进程管理命令
var ProcessCmd = &cobra.Command{
	Use:   "process",
	Short: "进程管理工具",
	Long: `进程管理工具提供了查看、查找、终止和控制系统进程的功能。

示例:
  %[1]s process list               # 列出所有进程
  %[1]s process list --filter chrome  # 列出包含'chrome'的进程
  %[1]s process info 1234         # 显示PID为1234的进程详情
  %[1]s process kill 1234         # 终止PID为1234的进程
  %[1]s process children 1234     # 列出PID为1234的所有子进程`,
}

func init() {
	// 子命令在各自文件的init函数中添加到ProcessCmd
}
