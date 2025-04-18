package process

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"toolbox/pkg/process"

	"github.com/spf13/cobra"
)

// childrenCmd 表示列出子进程的命令
var childrenCmd = &cobra.Command{
	Use:   "children [pid]",
	Short: "列出指定进程的子进程",
	Long: `列出指定PID进程的所有子进程。

示例:
  %[1]s process children 1234     # 列出PID为1234的进程的所有子进程`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 开始计时
		startTime := time.Now()

		// 解析PID
		pid, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			fmt.Printf("无效的PID: %v\n", err)
			os.Exit(1)
		}

		// 获取父进程信息
		parentInfo, err := process.GetProcessByPID(int32(pid))
		if err != nil {
			fmt.Printf("获取进程信息失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("正在查找进程 %d (%s) 的子进程...\n\n", parentInfo.PID, parentInfo.Name)

		// 获取子进程
		children, err := process.GetChildProcesses(int32(pid))
		if err != nil {
			fmt.Printf("获取子进程失败: %v\n", err)
			os.Exit(1)
		}

		if len(children) == 0 {
			fmt.Printf("进程 %d 没有子进程\n", pid)
			return
		}

		// 获取命令行显示选项
		fullCmd, _ := cmd.Flags().GetBool("full-cmd")

		// 输出子进程列表
		fmt.Printf("找到 %d 个子进程:\n\n", len(children))
		printProcessList(children, fullCmd)

		// 显示执行时间
		fmt.Printf("执行时间: %.2f秒\n", time.Since(startTime).Seconds())
	},
}

func init() {
	ProcessCmd.AddCommand(childrenCmd)

	// 添加命令行标志
	childrenCmd.Flags().BoolP("full-cmd", "c", false, "显示完整命令行")
}
