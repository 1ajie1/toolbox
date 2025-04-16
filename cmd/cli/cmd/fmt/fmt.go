package fmt

import (
	"github.com/spf13/cobra"
)

// FmtCmd 表示数据格式化命令组
var FmtCmd = &cobra.Command{
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
}

func init() {
	// 添加命令行标志
	FmtCmd.Flags().StringP("format", "f", "", "指定格式 (json, xml, yaml)")
	FmtCmd.Flags().BoolP("pretty", "p", false, "美化输出")
	FmtCmd.Flags().BoolP("compact", "c", false, "压缩输出（仅JSON/XML）")
	FmtCmd.Flags().IntP("indent", "i", 0, "缩进空格数 (默认: json/xml=4, yaml=2)")
	FmtCmd.Flags().BoolP("color", "", false, "彩色输出")
	FmtCmd.Flags().StringP("output", "o", "", "输出到文件而非标准输出")
	FmtCmd.Flags().BoolP("string", "s", false, "将参数作为字符串内容而非文件路径")
	FmtCmd.Flags().StringP("delimiter", "d", "#", "指定包围内容的分隔符，如 # 或 --- 等")

	// 添加子命令
	FmtCmd.AddCommand(formatCmd)
}
