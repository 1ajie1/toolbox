package text

import (
	"fmt"
	"os"

	"tuleaj_tools/tool-box/pkg/textproc"

	"github.com/spf13/cobra"
)

// textGrepCmd 表示文本搜索命令
var textGrepCmd = &cobra.Command{
	Use:   "grep [模式] [文件路径...]",
	Short: "搜索文本内容",
	Long: `在文件或标准输入中搜索匹配指定模式的文本行。

支持正则表达式搜索，可以高亮显示匹配部分，统计匹配行数等。

示例:
  %[1]s text grep "error" log.txt           # 搜索log.txt文件中包含"error"的行
  %[1]s text grep "^[0-9]+" file.txt        # 使用正则表达式搜索以数字开头的行
  cat file.txt | %[1]s text grep "pattern"  # 从标准输入搜索
  %[1]s text grep -n "pattern" file.txt     # 显示行号
  %[1]s text grep -i "pattern" file.txt     # 忽略大小写搜索
  %[1]s text grep -r "pattern" ./src        # 递归搜索目录
  %[1]s text grep -r -f "*.go" "func" ./src # 递归搜索目录中的go文件`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("错误: 必须指定搜索模式")
			cmd.Help()
			os.Exit(1)
		}

		// 获取选项
		pattern := args[0]
		ignoreCase, _ := cmd.Flags().GetBool("ignore-case")
		showLineNum, _ := cmd.Flags().GetBool("line-number")
		invertMatch, _ := cmd.Flags().GetBool("invert-match")
		onlyCount, _ := cmd.Flags().GetBool("count")
		colorOutput, _ := cmd.Flags().GetBool("color")
		contextLines, _ := cmd.Flags().GetInt("context")
		recursive, _ := cmd.Flags().GetBool("recursive")
		filePattern, _ := cmd.Flags().GetString("file-pattern")
		excludeDirs, _ := cmd.Flags().GetStringSlice("exclude-dir")

		// 创建grep选项
		options := textproc.GrepOptions{
			Pattern:      pattern,
			IgnoreCase:   ignoreCase,
			ShowLineNum:  showLineNum,
			InvertMatch:  invertMatch,
			OnlyCount:    onlyCount,
			ColorOutput:  colorOutput,
			ContextLines: contextLines,
			Recursive:    recursive,
			FilePattern:  filePattern,
			ExcludeDirs:  excludeDirs,
		}

		// 确定输入源
		var sources []string
		if len(args) > 1 {
			sources = args[1:]
		} else {
			// 检查是否有标准输入
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				sources = []string{"-"} // 表示从标准输入读取
			} else {
				fmt.Println("错误: 未指定输入文件，且无标准输入")
				cmd.Help()
				os.Exit(1)
			}
		}

		// 处理每个输入源
		totalMatches := 0
		for _, source := range sources {
			// 递归处理目录
			if recursive {
				// 检查是否是目录
				fileInfo, err := os.Stat(source)
				if err != nil {
					fmt.Printf("错误: 无法访问 %s: %v\n", source, err)
					continue
				}

				if fileInfo.IsDir() {
					// 是目录，递归搜索
					result, err := textproc.GrepDirectory(source, os.Stdout, options)
					if err != nil {
						fmt.Printf("错误: %v\n", err)
						continue
					}

					totalMatches += result.Matches

					if len(sources) > 1 && !onlyCount {
						fmt.Println() // 源之间添加空行
					}

					continue
				}
				// 不是目录，按普通文件处理
			}

			var file *os.File
			var sourceName string

			if source == "-" {
				file = os.Stdin
				sourceName = "标准输入"
			} else {
				var err error
				file, err = os.Open(source)
				if err != nil {
					fmt.Printf("错误: 无法打开文件 %s: %v\n", source, err)
					continue
				}
				defer file.Close()
				sourceName = source
			}

			// 执行搜索
			result, err := textproc.ExecuteGrep(file, os.Stdout, options, sourceName)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}

			totalMatches += result.Matches

			if len(sources) > 1 && !onlyCount {
				fmt.Println() // 文件之间添加空行
			}
		}

		// 如果只需计数，输出匹配总数
		if onlyCount && !recursive {
			fmt.Println(totalMatches)
		}
	},
}

func init() {
	TextCmd.AddCommand(textGrepCmd)

	// 添加命令行标志
	textGrepCmd.Flags().BoolP("ignore-case", "i", false, "忽略大小写")
	textGrepCmd.Flags().BoolP("line-number", "n", true, "显示行号")
	textGrepCmd.Flags().BoolP("invert-match", "v", false, "反向匹配（显示不匹配的行）")
	textGrepCmd.Flags().BoolP("count", "c", false, "只显示匹配的行数")
	textGrepCmd.Flags().BoolP("color", "", true, "彩色输出匹配部分")
	textGrepCmd.Flags().IntP("context", "C", 0, "显示匹配行前后的上下文行数")
	textGrepCmd.Flags().BoolP("recursive", "r", false, "递归搜索目录")
	textGrepCmd.Flags().StringP("file-pattern", "f", "", "文件名匹配模式（正则表达式）")
	textGrepCmd.Flags().StringSliceP("exclude-dir", "e", []string{}, "排除的目录名（可重复使用此选项指定多个目录）")
}
