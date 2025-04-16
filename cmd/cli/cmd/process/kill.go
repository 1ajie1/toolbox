package process

import (
	"fmt"
	"os"
	"strconv"
	"tuleaj_tools/tool-box/pkg/process"

	"github.com/spf13/cobra"
)

// killCmd 表示终止进程的命令
var killCmd = &cobra.Command{
	Use:   "kill [pid]",
	Short: "终止指定进程",
	Long: `终止指定PID的进程，尝试先优雅地终止，如果失败则强制终止。

示例:
  %[1]s process kill 1234     # 终止PID为1234的进程`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 解析PID
		pid, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			fmt.Printf("无效的PID: %v\n", err)
			os.Exit(1)
		}

		// 获取进程信息
		procInfo, err := process.GetProcessByPID(int32(pid))
		if err != nil {
			fmt.Printf("获取进程信息失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("正在终止进程 %d (%s)...\n", procInfo.PID, procInfo.Name)

		// 终止进程
		err = process.KillProcess(int32(pid))
		if err != nil {
			fmt.Printf("终止进程失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("进程 %d 已成功终止\n", pid)
	},
}

func init() {
	ProcessCmd.AddCommand(killCmd)
}
