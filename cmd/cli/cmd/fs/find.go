package fs

import (
	"fmt"
	"os"
	"strings"
	"time"

	"toolbox/pkg/fsutils"

	"github.com/spf13/cobra"
)

// findCmd 表示 find 命令
var findCmd = &cobra.Command{
	Use:   "find [目录路径]",
	Short: "搜索文件和目录",
	Long: `在指定目录中搜索文件和目录，支持多种搜索条件。

示例:
  %[1]s fs find .                         # 列出当前目录下的所有文件
  %[1]s fs find /path -name "*.go"        # 搜索Go源文件
  %[1]s fs find . -type f                 # 只搜索普通文件
  %[1]s fs find . -type d                 # 只搜索目录
  %[1]s fs find . -size +1M              # 搜索大于1MB的文件
  %[1]s fs find . -mtime -7              # 搜索7天内修改的文件
  %[1]s fs find . -regex ".*\\.txt$"     # 使用正则表达式搜索txt文件
  %[1]s fs find . -maxdepth 2            # 最大搜索深度为2层
  %[1]s fs find . -exclude "node_modules" # 排除node_modules目录
  %[1]s fs find . -include "src,lib"     # 只在src和lib目录中搜索`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取搜索根目录
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		// 获取命令行选项
		name, _ := cmd.Flags().GetString("name")
		fileType, _ := cmd.Flags().GetString("type")
		minSize, _ := cmd.Flags().GetString("minsize")
		maxSize, _ := cmd.Flags().GetString("maxsize")
		minDepth, _ := cmd.Flags().GetInt("mindepth")
		maxDepth, _ := cmd.Flags().GetInt("maxdepth")
		mtime, _ := cmd.Flags().GetInt("mtime")
		regex, _ := cmd.Flags().GetString("regex")
		excludeDirs, _ := cmd.Flags().GetStringSlice("exclude")
		includeDirs, _ := cmd.Flags().GetStringSlice("include")
		followSymlinks, _ := cmd.Flags().GetBool("follow")

		// 创建搜索选项
		options := fsutils.FindOptions{
			Name:           name,
			Type:           fileType,
			MinDepth:       minDepth,
			MaxDepth:       maxDepth,
			Regex:          regex,
			ExcludeDirs:    excludeDirs,
			IncludeDirs:    includeDirs,
			FollowSymlinks: followSymlinks,
		}

		// 处理文件大小选项
		if minSize != "" {
			size, err := parseSize(minSize)
			if err != nil {
				fmt.Printf("错误: 无效的最小文件大小: %v\n", err)
				os.Exit(1)
			}
			options.MinSize = size
		}
		if maxSize != "" {
			size, err := parseSize(maxSize)
			if err != nil {
				fmt.Printf("错误: 无效的最大文件大小: %v\n", err)
				os.Exit(1)
			}
			options.MaxSize = size
		}

		// 处理修改时间
		if mtime != 0 {
			if mtime > 0 {
				options.ModifiedBefore = time.Now().AddDate(0, 0, -mtime)
			} else {
				options.ModifiedAfter = time.Now().AddDate(0, 0, mtime)
			}
		}

		// 执行搜索
		err := fsutils.ExecuteFind(root, os.Stdout, options)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	FsCmd.AddCommand(findCmd)

	// 添加命令行标志
	findCmd.Flags().StringP("name", "n", "", "按文件名搜索（支持通配符）")
	findCmd.Flags().StringP("type", "t", "", "按类型搜索 (f:文件, d:目录, l:符号链接)")
	findCmd.Flags().StringP("minsize", "", "", "最小文件大小 (例如: 1M, 500K)")
	findCmd.Flags().StringP("maxsize", "", "", "最大文件大小 (例如: 10M, 1G)")
	findCmd.Flags().IntP("mindepth", "", 0, "最小搜索深度")
	findCmd.Flags().IntP("maxdepth", "", 0, "最大搜索深度")
	findCmd.Flags().IntP("mtime", "m", 0, "按修改时间搜索（天数，负数表示之内，正数表示之前）")
	findCmd.Flags().StringP("regex", "r", "", "使用正则表达式匹配文件名")
	findCmd.Flags().StringSliceP("exclude", "e", nil, "排除的目录（可多次使用）")
	findCmd.Flags().StringSliceP("include", "i", nil, "只在指定目录中搜索（可多次使用）")
	findCmd.Flags().BoolP("follow", "L", false, "跟随符号链接")
}

// parseSize 解析文件大小字符串（如 1K, 2M, 3G）
func parseSize(sizeStr string) (int64, error) {
	var size float64
	var unit string
	_, err := fmt.Sscanf(sizeStr, "%f%s", &size, &unit)
	if err != nil {
		return 0, fmt.Errorf("无效的大小格式")
	}

	unit = strings.ToUpper(unit)
	var multiplier float64
	switch unit {
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "B", "":
		multiplier = 1
	default:
		return 0, fmt.Errorf("未知的单位: %s", unit)
	}

	return int64(size * multiplier), nil
}
