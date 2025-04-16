package process

import (
	"fmt"
	"strconv"
	"tuleaj_tools/tool-box/pkg/process"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// 定义颜色常量，用于错误显示
var (
	errorColor = color.New(color.FgRed, color.Bold)
)

// treeCmd 表示以树形结构显示进程的命令
var treeCmd = &cobra.Command{
	Use:   "tree [pid]",
	Short: "以树形结构显示进程",
	Long: `以树形结构显示进程及其子进程的层级关系。
如果提供了PID参数，则显示该进程及其子进程的树形结构；
如果未提供PID，则显示所有进程的树形结构。

示例:
  %[1]s process tree       # 显示所有进程的树形结构
  %[1]s process tree 1234  # 显示PID为1234的进程及其子进程的树形结构`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取所有进程
		processList, err := process.GetProcessList()
		if err != nil {
			errorColor.Printf("获取进程列表失败: %v\n", err)
			return
		}

		// 使用筛选名称参数
		filter, _ := cmd.Flags().GetString("filter")
		// 获取是否显示详细信息
		showDetail, _ := cmd.Flags().GetBool("detail")
		// 获取是否显示彩色输出
		noColor, _ := cmd.Flags().GetBool("no-color")

		// 构建进程树选项
		options := process.ProcessTreeOptions{
			RootPID:       0, // 默认从系统进程开始
			Filter:        filter,
			IncludeOrphan: true,
		}

		// 如果提供了PID参数，则从该进程开始
		if len(args) > 0 {
			if pid, err := strconv.ParseInt(args[0], 10, 32); err == nil {
				options.RootPID = int32(pid)
			} else {
				errorColor.Printf("无效的PID: %v\n", err)
				return
			}
		}

		// 构建进程树
		tree, err := process.BuildProcessTree(processList, options)
		if err != nil {
			if options.RootPID != 0 {
				errorColor.Printf("构建进程树失败: %v\n", err)
				return
			}
			// 如果是根进程树构建失败，尝试使用一个备用方案
			errorColor.Printf("警告: %v, 尝试使用备用方法...\n", err)

			// 创建一个模拟的系统节点
			tree = &process.ProcessTreeNode{
				Process: process.ProcessInfo{
					PID:  0,
					Name: "System",
				},
				Children: []*process.ProcessTreeNode{},
			}
		}

		// 创建渲染器
		renderer := process.NewTableRenderer(showDetail, noColor)

		// 设置标题
		if options.RootPID == 0 {
			renderer.Title = "系统进程树:"
		} else {
			renderer.Title = fmt.Sprintf("进程 %d (%s) 的进程树:",
				tree.Process.PID, tree.Process.Name)
		}

		// 渲染进程树
		if err := renderer.Render(tree); err != nil {
			errorColor.Printf("渲染进程树失败: %v\n", err)
			return
		}
	},
}

func init() {
	ProcessCmd.AddCommand(treeCmd)

	// 添加命令行标志
	treeCmd.Flags().StringP("filter", "f", "", "按进程名称过滤")
	treeCmd.Flags().BoolP("detail", "d", false, "显示详细信息，包括内存和CPU使用情况")
	treeCmd.Flags().Bool("no-color", false, "禁用彩色输出")
}
