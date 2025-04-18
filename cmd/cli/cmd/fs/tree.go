package fs

import (
	"fmt"
	"os"

	"toolbox/pkg/fsutils"

	"github.com/spf13/cobra"
)

// treeCmd 表示 tree 命令
var treeCmd = &cobra.Command{
	Use:   "tree [目录路径]",
	Short: "显示目录结构",
	Long: `显示指定目录的文件和子目录结构，类似于Linux/Windows的tree命令。
该命令以树状图形方式展示目录结构，可以指定显示深度、过滤条件等。

示例:
  %[1]s fs tree                   # 显示当前目录的结构
  %[1]s fs tree /path/to/dir      # 显示指定目录的结构
  %[1]s fs tree -d 2              # 只显示两层深度
  %[1]s fs tree -a                # 显示隐藏文件
  %[1]s fs tree -L                # 跟随符号链接
  %[1]s fs tree -D                # 只显示目录
  %[1]s fs tree -s                # 显示文件大小`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取目录路径参数
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// 获取选项
		maxDepth, _ := cmd.Flags().GetInt("depth")
		showHidden, _ := cmd.Flags().GetBool("all")
		onlyDirs, _ := cmd.Flags().GetBool("dirs-only")
		followSymlink, _ := cmd.Flags().GetBool("follow")
		showSize, _ := cmd.Flags().GetBool("size")

		// 创建选项
		options := fsutils.TreeOptions{
			MaxDepth:      maxDepth,
			ShowHidden:    showHidden,
			OnlyDirs:      onlyDirs,
			FollowSymlink: followSymlink,
			ShowSize:      showSize,
		}

		// 执行目录树展示
		_, err := fsutils.DisplayTree(path, os.Stdout, options)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	FsCmd.AddCommand(treeCmd)

	// 添加命令行标志
	treeCmd.Flags().IntP("depth", "d", 0, "最大显示深度 (0表示无限制)")
	treeCmd.Flags().BoolP("all", "a", false, "显示所有文件，包括隐藏文件")
	treeCmd.Flags().BoolP("dirs-only", "D", false, "只显示目录")
	treeCmd.Flags().BoolP("follow", "L", false, "跟随符号链接")
	treeCmd.Flags().BoolP("size", "s", false, "显示文件大小")
}
