package fs

import (
	"github.com/spf13/cobra"
)

// FsCmd 表示文件系统相关命令组
var FsCmd = &cobra.Command{
	Use:   "fs",
	Short: "文件系统工具集",
	Long: `文件系统工具集，包含多种文件和目录操作相关命令。

示例:
  %[1]s fs tree             # 显示当前目录的树状结构
  %[1]s fs tree /path/to/dir -d 2 -a  # 显示指定目录的结构(深度2，包含隐藏文件)`,
}

func init() {
	// 添加子命令
	FsCmd.AddCommand(treeCmd)
}
