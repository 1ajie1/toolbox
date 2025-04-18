package process

import (
	"fmt"
	"os"
	"sort"
	"time"
	"toolbox/pkg/process"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd 表示列出进程的命令
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出系统进程",
	Long: `列出系统中的进程信息，支持过滤和排序。

示例:
  %[1]s process list                # 列出所有进程
  %[1]s process list --filter chrome  # 列出名称包含'chrome'的进程
  %[1]s process list --sort cpu     # 按CPU使用率排序
  %[1]s process list --sort memory  # 按内存使用率排序
  %[1]s process list --show-system  # 显示系统进程
  %[1]s process list --no-empty     # 不显示没有名称的进程
  %[1]s process list --full-cmd     # 显示完整命令行`,
	Run: func(cmd *cobra.Command, args []string) {
		// 开始计时
		startTime := time.Now()

		// 获取参数
		filter, _ := cmd.Flags().GetString("filter")
		sortBy, _ := cmd.Flags().GetString("sort")
		top, _ := cmd.Flags().GetInt("top")
		showSystem, _ := cmd.Flags().GetBool("show-system")
		noEmpty, _ := cmd.Flags().GetBool("no-empty")
		fullCmd, _ := cmd.Flags().GetBool("full-cmd")

		var processList []process.ProcessInfo
		var err error

		// 按名称过滤
		if filter != "" {
			// 使用名称筛选
			processList, err = process.FilterProcessesByName(filter)
			if err != nil {
				fmt.Printf("获取进程列表失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("找到 %d 个匹配 '%s' 的进程\n", len(processList), filter)
		} else {
			// 获取所有进程
			processList, err = process.GetProcessList()
			if err != nil {
				fmt.Printf("获取进程列表失败: %v\n", err)
				os.Exit(1)
			}
		}

		// 过滤系统进程和空名称进程
		if !showSystem || noEmpty {
			var filtered []process.ProcessInfo
			for _, p := range processList {
				// 如果不显示空名称进程，过滤掉无名称的进程
				if noEmpty && p.Name == "" {
					continue
				}

				filtered = append(filtered, p)
			}
			processList = filtered
		}

		// 按特定字段排序
		if sortBy != "" {
			sortProcessList(processList, sortBy)
		} else {
			// 默认按PID排序
			sortProcessList(processList, "pid")
		}

		// 限制显示的数量
		if top > 0 && top < len(processList) {
			processList = processList[:top]
		}

		// 输出结果
		printProcessList(processList, fullCmd)

		// 显示执行时间
		fmt.Printf("执行时间: %.2f秒\n", time.Since(startTime).Seconds())
	},
}

func init() {
	ProcessCmd.AddCommand(listCmd)

	// 添加命令行标志
	listCmd.Flags().StringP("filter", "f", "", "按进程名称过滤")
	listCmd.Flags().StringP("sort", "s", "", "排序方式 (pid, cpu, memory)")
	listCmd.Flags().IntP("top", "n", 0, "只显示前N个进程")
	listCmd.Flags().BoolP("show-system", "S", false, "显示系统进程")
	listCmd.Flags().BoolP("no-empty", "e", false, "不显示没有名称的进程")
	listCmd.Flags().BoolP("full-cmd", "c", false, "显示完整命令行")
}

// 根据指定字段对进程列表进行排序
func sortProcessList(processes []process.ProcessInfo, sortBy string) {
	switch sortBy {
	case "cpu":
		// 按CPU使用率从高到低排序
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].CPU > processes[j].CPU
		})
	case "memory":
		// 按内存使用率从高到低排序
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].Memory > processes[j].Memory
		})
	case "pid":
		// 按PID从小到大排序
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].PID < processes[j].PID
		})
	default:
		// 默认按PID排序
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].PID < processes[j].PID
		})
	}
}

// 打印进程列表
func printProcessList(processes []process.ProcessInfo, fullCmd bool) {
	// 创建一个新的表格写入器
	table := tablewriter.NewWriter(os.Stdout)

	// 设置表头
	table.SetHeader([]string{"PID", "PPID", "CPU%", "MEM%", "用户", "名称", "命令"})

	// 设置表格样式
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(true)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	// 为每个进程创建一行数据
	for _, p := range processes {
		// 格式化进程名称
		name := p.Name
		if name == "" {
			name = color.YellowString("[无名称]")
		}

		// 格式化用户名
		username := p.Username
		if username == "" {
			username = "-"
		}

		// 获取命令行
		cmdLine := formatCmdLine(p.CmdLine, fullCmd)
		if cmdLine == "" && p.Executable != "" {
			cmdLine = p.Executable
		}

		// 添加行数据
		table.Append([]string{
			fmt.Sprintf("%d", p.PID),
			fmt.Sprintf("%d", p.PPID),
			fmt.Sprintf("%.1f", p.CPU),
			fmt.Sprintf("%.1f", p.Memory),
			username,
			name,
			cmdLine,
		})
	}

	// 渲染表格
	table.Render()

	// 添加总结信息
	fmt.Printf("\n共 %d 个进程\n", len(processes))
}

// 格式化命令行
func formatCmdLine(cmdLine []string, fullCmd bool) string {
	if len(cmdLine) == 0 {
		return ""
	}

	if fullCmd {
		// 显示完整命令行
		result := ""
		for i, arg := range cmdLine {
			if i > 0 {
				result += " "
			}
			result += arg
		}
		return result
	} else {
		// 显示精简命令行
		if len(cmdLine) == 1 {
			result := cmdLine[0]
			if len(result) > 50 {
				result = result[:47] + "..."
			}
			return result
		} else {
			result := cmdLine[0]
			if len(cmdLine) > 1 {
				result += " " + cmdLine[1]
				if len(cmdLine) > 2 {
					result += "..."
				}
			}
			if len(result) > 50 {
				result = result[:47] + "..."
			}
			return result
		}
	}
}
