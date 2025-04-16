package text

import (
	"fmt"
	"os"

	"tuleaj_tools/tool-box/pkg/textproc"

	"github.com/spf13/cobra"
)

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
	TextCmd.AddCommand(textFilterCmd)

	// 添加命令行标志
	textFilterCmd.Flags().StringP("field-separator", "F", " ", "字段分隔符")
	textFilterCmd.Flags().StringP("print", "p", "", "输出格式模式")
}
