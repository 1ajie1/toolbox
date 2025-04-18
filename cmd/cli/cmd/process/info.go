package process

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"toolbox/pkg/process"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// infoCmd 表示显示进程详情的命令
var infoCmd = &cobra.Command{
	Use:   "info [pid]",
	Short: "显示进程详情",
	Long: `显示指定PID进程的详细信息，包括CPU使用率、内存使用情况、启动时间等。

示例:
  %[1]s process info 1234     # 显示PID为1234的进程详细信息`,
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

		// 打印进程详情
		printProcessInfo(procInfo)
	},
}

func init() {
	ProcessCmd.AddCommand(infoCmd)
}

// 打印进程详细信息
func printProcessInfo(p process.ProcessInfo) {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	fmt.Println("==============进程详情==============")
	bold.Printf("进程ID: ")
	fmt.Printf("%d\n", p.PID)

	bold.Printf("父进程ID: ")
	fmt.Printf("%d\n", p.PPID)

	bold.Printf("进程名称: ")
	fmt.Printf("%s\n", p.Name)

	bold.Printf("可执行文件: ")
	fmt.Printf("%s\n", p.Executable)

	bold.Printf("用户: ")
	fmt.Printf("%s\n", p.Username)

	bold.Printf("状态: ")
	fmt.Printf("%s\n", p.Status)

	bold.Printf("创建时间: ")
	fmt.Printf("%s (%s前)\n",
		p.CreateTime.Format("2006-01-02 15:04:05"),
		formatDuration(time.Since(p.CreateTime)))

	bold.Printf("CPU使用率: ")
	fmt.Printf("%.2f%%\n", p.CPU)

	bold.Printf("内存使用: ")
	fmt.Printf("%.2f%% (RSS: %s, 虚拟: %s)\n",
		p.Memory,
		formatBytes(p.MemoryInfo.RSS),
		formatBytes(p.MemoryInfo.VMS))

	bold.Printf("线程数: ")
	fmt.Printf("%d\n", p.Threads)

	// 打印命令行
	bold.Println("命令行:")
	if len(p.CmdLine) > 0 {
		cmdLine := ""
		for i, arg := range p.CmdLine {
			if i > 0 {
				cmdLine += " "
			}
			cmdLine += arg
		}
		cyan.Printf("  %s\n", cmdLine)
	} else {
		yellow.Println("  [无法获取命令行]")
	}

	// 打印打开的文件
	if len(p.OpenFiles) > 0 {
		bold.Println("打开的文件:")
		for _, file := range p.OpenFiles {
			fmt.Printf("  %s\n", file)
		}
	}

	fmt.Println("===================================")
}

// 格式化时间间隔
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天%d小时", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分钟%d秒", minutes, seconds)
	}
	return fmt.Sprintf("%d秒", seconds)
}

// 格式化字节数为人类可读格式
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
