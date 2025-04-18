package text

import (
	"github.com/spf13/cobra"
)

// TextCmd 表示文本处理命令
var TextCmd = &cobra.Command{
	Use:   "text",
	Short: "文本处理工具",
	Long: `文本处理工具集，提供类似sed、grep、awk的功能。

包含以下子命令:
  grep - 搜索文本内容
  replace - 替换文本内容
  filter - 过滤文本行`,
}

func init() {
	// 添加子命令
}
