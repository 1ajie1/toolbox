package fmt

import (
	"fmt"
	"os"
	"strings"
	"tuleaj_tools/tool-box/pkg/formatter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// formatCmd 实现格式化命令
var formatCmd = &cobra.Command{
	Use:   "fmt [文件路径|文本内容]",
	Short: "格式化数据文件或文本内容",
	Long: `格式化数据文件或文本内容，支持JSON/XML/YAML格式的美化和压缩。

示例:
  %[1]s fmt data.json --pretty --color    # 美化并着色JSON文件
  %[1]s fmt data.xml --pretty             # 美化XML文件
  %[1]s fmt data.json --compact           # 压缩JSON文件
  %[1]s fmt data.yaml --pretty            # 美化YAML文件
  %[1]s fmt '{"name":"John"}' --format json --pretty  # 美化JSON文本
  %[1]s fmt -s '<root><item>1</item></root>' --format xml --pretty  # 美化XML文本内容
  %[1]s fmt -s '#{"name":"网络工具箱"}#' --format json --pretty --delimiter '#'  # 使用自定义分隔符`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取参数
		format, _ := cmd.Flags().GetString("format")
		pretty, _ := cmd.Flags().GetBool("pretty")
		compact, _ := cmd.Flags().GetBool("compact")
		indent, _ := cmd.Flags().GetInt("indent")
		useColor, _ := cmd.Flags().GetBool("color")
		output, _ := cmd.Flags().GetString("output")
		isString, _ := cmd.Flags().GetBool("string")
		delimiter, _ := cmd.Flags().GetString("delimiter")

		// 创建格式化选项
		opts := formatter.Options{
			Pretty:  pretty,
			Compact: compact,
			Indent:  indent,
			Color:   useColor,
		}

		// 判断输入来源
		if isString {
			// 从命令行参数获取字符串内容
			if len(args) < 1 {
				fmt.Println("错误: 使用 --string 选项时必须提供文本内容")
				cmd.Help()
				os.Exit(1)
			}

			// 获取文本内容
			content := args[0]

			// 如果指定了分隔符，尝试提取内容
			if delimiter != "" {
				if extractedContent, found := formatter.ExtractContentWithDelimiter(content, delimiter); found {
					content = extractedContent
					fmt.Printf("已从分隔符 '%s' 中提取内容\n", delimiter)
				} else {
					fmt.Printf("警告: 未找到使用分隔符 '%s' 包围的内容\n", delimiter)
				}
			}

			// 必须指定格式
			if format == "" {
				fmt.Println("错误: 处理文本内容时必须使用 --format 指定格式")
				os.Exit(1)
			}

			opts.Format = formatter.FormatType(format)

			// 执行文本格式化
			executeStringFmt(content, opts, output)
		} else {
			// 从文件读取
			if len(args) < 1 {
				fmt.Println("错误: 必须指定数据文件路径或使用 --string 选项")
				cmd.Help()
				os.Exit(1)
			}

			filePath := args[0]

			// 如果没有指定格式，尝试从文件扩展名推断
			if format == "" {
				format = getFormatFromFileName(filePath)
				if format == "" {
					fmt.Println("错误: 无法从文件名推断格式，请使用 --format 指定格式")
					os.Exit(1)
				}
			}

			opts.Format = formatter.FormatType(format)

			// 执行文件格式化
			executeFileFmt(filePath, opts, output)
		}
	},
}

func init() {
	// 将实现命令添加到父命令
	FmtCmd.AddCommand(formatCmd)

	// 将父命令的标志也添加到实现命令
	formatCmd.Flags().StringP("format", "f", "", "指定格式 (json, xml, yaml)")
	formatCmd.Flags().BoolP("pretty", "p", false, "美化输出")
	formatCmd.Flags().BoolP("compact", "c", false, "压缩输出（仅JSON/XML）")
	formatCmd.Flags().IntP("indent", "i", 0, "缩进空格数 (默认: json/xml=4, yaml=2)")
	formatCmd.Flags().BoolP("color", "", false, "彩色输出")
	formatCmd.Flags().StringP("output", "o", "", "输出到文件而非标准输出")
	formatCmd.Flags().BoolP("string", "s", false, "将参数作为字符串内容而非文件路径")
	formatCmd.Flags().StringP("delimiter", "d", "", "指定包围内容的分隔符，如 # 或 --- 等")

	// 设置FmtCmd的Run字段指向formatCmd的Run函数
	FmtCmd.Run = formatCmd.Run
}

// getFormatFromFileName 根据文件名推断格式
func getFormatFromFileName(path string) string {
	lowerPath := strings.ToLower(path)
	if strings.HasSuffix(lowerPath, ".json") {
		return "json"
	} else if strings.HasSuffix(lowerPath, ".xml") {
		return "xml"
	} else if strings.HasSuffix(lowerPath, ".yaml") || strings.HasSuffix(lowerPath, ".yml") {
		return "yaml"
	}
	return ""
}

// executeFileFmt 执行文件格式化操作
func executeFileFmt(filePath string, opts formatter.Options, outputPath string) {
	// 使用粗体黄色打印
	boldYellow := color.New(color.FgYellow, color.Bold)
	boldYellow.Printf("格式化文件: %s\n", filePath)
	printFormatMode(boldYellow, opts)

	// 执行格式化
	result, err := formatter.FormatFile(filePath, opts)
	if err != nil {
		fmt.Printf("格式化失败: %v\n", err)
		os.Exit(1)
	}

	// 显示结果
	displayResult(result, outputPath)
}

// executeStringFmt 执行文本格式化操作
func executeStringFmt(content string, opts formatter.Options, outputPath string) {
	// 使用粗体黄色打印
	boldYellow := color.New(color.FgYellow, color.Bold)
	boldYellow.Println("格式化文本内容")
	printFormatMode(boldYellow, opts)

	// PowerShell 转义字符处理
	content = formatter.HandlePowerShellEscaping(content)

	// 调试信息
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("处理前的内容: %s\n", content)
	}

	// 执行格式化
	reader := strings.NewReader(content)
	result, err := formatter.Format(reader, opts)
	if err != nil {
		fmt.Printf("格式化失败: %v\n", err)

		// 只有在JSON格式且确实解析失败时才显示帮助提示
		if opts.Format == "json" && !gjson.Valid(content) {
			fmt.Println("提示: 您的输入似乎是未正确格式化的JSON。请确保：")
			fmt.Println("1. 所有的键名和字符串值都使用双引号")
			fmt.Println("2. Windows PowerShell中使用双引号包裹JSON字符串，并转义内部引号")
			fmt.Println("例如: '{\"name\":\"值\",\"array\":[1,2,3]}'")
			fmt.Println("或使用 PowerShell 的 @\"...\"@ 语法避免转义：")
			fmt.Println("$json = @\"")
			fmt.Println("{\"name\":\"值\",\"array\":[1,2,3]}")
			fmt.Println("\"@")
			fmt.Println("go run .\\cmd\\cli\\main.go fmt -s $json -f json -p")
		} else {
			fmt.Println("请检查输入格式是否正确，特别是JSON中的引号、括号和逗号。")
		}

		os.Exit(1)
	}

	// 显示结果
	displayResult(result, outputPath)
}

// printFormatMode 打印格式化模式
func printFormatMode(printer *color.Color, opts formatter.Options) {
	if opts.Pretty {
		printer.Println("模式: 美化")
	} else if opts.Compact {
		printer.Println("模式: 压缩")
	} else {
		printer.Println("模式: 标准")
	}
}

// displayResult 显示格式化结果
func displayResult(result *formatter.Result, outputPath string) {
	if outputPath != "" {
		// 保存到文件
		if err := result.ToFile(outputPath); err != nil {
			fmt.Printf("保存结果失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("已保存到: %s (大小: %d 字节)\n", outputPath, result.OutputSize)
	} else {
		// 直接输出到终端
		fmt.Println("\n------ 格式化结果 ------")
		fmt.Println(result.Output)
		fmt.Printf("\n------ 结果统计 ------\n")
		fmt.Printf("输入大小: %d 字节\n", result.InputSize)
		fmt.Printf("输出大小: %d 字节\n", result.OutputSize)
		fmt.Printf("处理耗时: %s\n", result.Duration)
	}
}
