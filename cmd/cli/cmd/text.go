package cmd

import (
	"fmt"
	"os"

	"tuleaj_tools/tool-box/pkg/textproc"

	"github.com/spf13/cobra"
)

// textCmd 表示文本处理命令
var textCmd = &cobra.Command{
	Use:   "text",
	Short: "文本处理工具",
	Long: `文本处理工具集，提供类似sed、grep、awk的功能。

包含以下子命令:
  grep - 搜索文本内容
  replace - 替换文本内容
  filter - 过滤文本行`,
}

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

// textReplaceCmd 表示文本替换命令
var textReplaceCmd = &cobra.Command{
	Use:   "replace [模式] [替换] [文件路径...]",
	Short: "替换文本内容",
	Long: `在文件或标准输入中查找并替换文本内容。

支持正则表达式和引用捕获组。

示例:
  %[1]s text replace "old" "new" file.txt                # 替换file.txt中的"old"为"new"
  %[1]s text replace "User-(\\d+)" "ID-$1" users.txt     # 使用正则表达式和引用
  cat file.txt | %[1]s text replace "pattern" "new" -    # 从标准输入替换并输出到标准输出
  %[1]s text replace -i "error" "warning" log.txt        # 忽略大小写替换
  %[1]s text replace -g "pattern" "new" file.txt         # 全局替换（每行多次）`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("错误: 必须指定搜索模式和替换文本")
			cmd.Help()
			os.Exit(1)
		}

		// 获取选项
		pattern := args[0]
		replacement := args[1]
		ignoreCase, _ := cmd.Flags().GetBool("ignore-case")
		globalReplace, _ := cmd.Flags().GetBool("global")
		inPlace, _ := cmd.Flags().GetBool("in-place")
		backup, _ := cmd.Flags().GetString("backup")

		// 创建replace选项
		options := textproc.ReplaceOptions{
			Pattern:       pattern,
			Replacement:   replacement,
			IgnoreCase:    ignoreCase,
			GlobalReplace: globalReplace,
		}

		// 确定输入源
		var sources []string
		if len(args) > 2 {
			sources = args[2:]
		} else {
			// 检查是否有标准输入
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				if inPlace {
					fmt.Println("错误: 无法对标准输入执行原地替换")
					os.Exit(1)
				}
				sources = []string{"-"} // 表示从标准输入读取
			} else {
				fmt.Println("错误: 未指定输入文件，且无标准输入")
				cmd.Help()
				os.Exit(1)
			}
		}

		// 处理每个输入源
		for _, source := range sources {
			if source == "-" {
				// 标准输入输出模式
				_, err := textproc.ExecuteReplace(os.Stdin, os.Stdout, options)
				if err != nil {
					fmt.Printf("错误: %v\n", err)
				}
			} else {
				// 文件模式
				if inPlace {
					// 原地替换模式
					tempFile := source + ".tmp"
					origFile, err := os.Open(source)
					if err != nil {
						fmt.Printf("错误: 无法打开文件 %s: %v\n", source, err)
						continue
					}

					// 创建备份（如果需要）
					if backup != "" {
						if err := textproc.CreateBackup(source, backup); err != nil {
							fmt.Printf("错误: 无法创建备份 %s: %v\n", source+backup, err)
							origFile.Close()
							continue
						}
					}

					// 创建临时文件用于写入
					tmpFile, err := os.Create(tempFile)
					if err != nil {
						fmt.Printf("错误: 无法创建临时文件: %v\n", err)
						origFile.Close()
						continue
					}

					// 执行替换
					result, err := textproc.ExecuteReplace(origFile, tmpFile, options)
					if err != nil {
						fmt.Printf("错误: %v\n", err)
						origFile.Close()
						tmpFile.Close()
						os.Remove(tempFile) // 清理临时文件
						continue
					}

					// 关闭文件
					origFile.Close()
					tmpFile.Close()

					// 用临时文件替换原文件
					if err := os.Rename(tempFile, source); err != nil {
						fmt.Printf("错误: 无法替换原文件: %v\n", err)
						os.Remove(tempFile) // 清理临时文件
						continue
					}

					fmt.Printf("已处理 %d 行，替换了 %d 处\n", result.LinesProcessed, result.Replacements)
				} else {
					// 输出到标准输出模式
					file, err := os.Open(source)
					if err != nil {
						fmt.Printf("错误: 无法打开文件 %s: %v\n", source, err)
						continue
					}
					defer file.Close()

					fmt.Printf("==> %s <==\n", source)
					textproc.ExecuteReplace(file, os.Stdout, options)

					if len(sources) > 1 {
						fmt.Println() // 文件之间添加空行
					}
				}
			}
		}
	},
}

// textFilterCmd 表示文本过滤命令
var textFilterCmd = &cobra.Command{
	Use:   "filter [表达式] [文件路径...]",
	Short: "过滤文本行",
	Long: `根据条件过滤文本行，类似awk的功能。

支持基本的条件表达式和字段选择。

示例:
  %[1]s text filter '$1 > 100' data.txt               # 过滤第一列大于100的行
  %[1]s text filter -F, '$2 == "ERROR"' log.csv       # 使用逗号分隔符，过滤第二列为ERROR的行
  %[1]s text filter 'length($0) > 80' file.txt        # 过滤长度大于80的行
  cat file.txt | %[1]s text filter '$3 ~ /pattern/'   # 过滤第三列匹配正则表达式的行
  %[1]s text filter -p '${1} ${3}' data.txt           # 只打印第1和第3列`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("错误: 必须指定过滤表达式")
			cmd.Help()
			os.Exit(1)
		}

		// 获取选项
		expression := args[0]
		fieldSep, _ := cmd.Flags().GetString("field-separator")
		printPattern, _ := cmd.Flags().GetString("print")

		// 创建filter选项
		options := textproc.FilterOptions{
			Expression:   expression,
			FieldSep:     fieldSep,
			PrintPattern: printPattern,
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
		for _, source := range sources {
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

			// 执行过滤
			if len(sources) > 1 {
				fmt.Printf("==> %s <==\n", sourceName)
			}

			_, err := textproc.ExecuteFilter(file, os.Stdout, options)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}

			if len(sources) > 1 {
				fmt.Println() // 文件之间添加空行
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(textCmd)
	textCmd.AddCommand(textGrepCmd)
	textCmd.AddCommand(textReplaceCmd)
	textCmd.AddCommand(textFilterCmd)

	// grep命令选项
	textGrepCmd.Flags().BoolP("ignore-case", "i", false, "忽略大小写")
	textGrepCmd.Flags().BoolP("line-number", "n", true, "显示行号")
	textGrepCmd.Flags().BoolP("invert-match", "v", false, "反向匹配（显示不匹配的行）")
	textGrepCmd.Flags().BoolP("count", "c", false, "只显示匹配的行数")
	textGrepCmd.Flags().BoolP("color", "", true, "彩色输出匹配部分")
	textGrepCmd.Flags().IntP("context", "C", 0, "显示匹配行前后的上下文行数")
	textGrepCmd.Flags().BoolP("recursive", "r", false, "递归搜索目录")
	textGrepCmd.Flags().StringP("file-pattern", "f", "", "文件名匹配模式（正则表达式）")
	textGrepCmd.Flags().StringSliceP("exclude-dir", "e", []string{}, "排除的目录名（可重复使用此选项指定多个目录）")

	// replace命令选项
	textReplaceCmd.Flags().BoolP("ignore-case", "i", false, "忽略大小写")
	textReplaceCmd.Flags().BoolP("global", "g", false, "全局替换（每行多次）")
	textReplaceCmd.Flags().BoolP("in-place", "I", false, "原地修改文件")
	textReplaceCmd.Flags().StringP("backup", "b", "", "创建备份，指定备份后缀")

	// filter命令选项
	textFilterCmd.Flags().StringP("field-separator", "F", " ", "字段分隔符")
	textFilterCmd.Flags().StringP("print", "p", "", "输出格式模式")
}
