package text

import (
	"fmt"
	"os"

	"tuleaj_tools/tool-box/pkg/textproc"

	"github.com/spf13/cobra"
)

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

func init() {
	TextCmd.AddCommand(textReplaceCmd)

	// 添加命令行标志
	textReplaceCmd.Flags().BoolP("ignore-case", "i", false, "忽略大小写")
	textReplaceCmd.Flags().BoolP("global", "g", false, "全局替换（每行多次）")
	textReplaceCmd.Flags().BoolP("in-place", "I", false, "原地修改文件")
	textReplaceCmd.Flags().StringP("backup", "b", "", "创建备份，指定备份后缀")
}
