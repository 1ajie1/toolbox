package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
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
  %[1]s text grep -i "pattern" file.txt     # 忽略大小写搜索`,
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

		// 编译正则表达式
		var regexpOpt string
		if ignoreCase {
			regexpOpt = "(?i)"
		}
		re, err := regexp.Compile(regexpOpt + pattern)
		if err != nil {
			fmt.Printf("错误: 无效的正则表达式: %v\n", err)
			os.Exit(1)
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
			matches := executeGrep(file, re, showLineNum, invertMatch, colorOutput, contextLines, sourceName)
			totalMatches += matches

			if len(sources) > 1 && !onlyCount {
				fmt.Println() // 文件之间添加空行
			}
		}

		// 如果只需计数，输出匹配总数
		if onlyCount {
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

		// 编译正则表达式
		var regexpOpt string
		if ignoreCase {
			regexpOpt = "(?i)"
		}
		re, err := regexp.Compile(regexpOpt + pattern)
		if err != nil {
			fmt.Printf("错误: 无效的正则表达式: %v\n", err)
			os.Exit(1)
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
				executeReplace(os.Stdin, os.Stdout, re, replacement, globalReplace, "")
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
						backupFile := source + backup
						if err := copyFile(source, backupFile); err != nil {
							fmt.Printf("错误: 无法创建备份 %s: %v\n", backupFile, err)
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
					executeReplace(origFile, tmpFile, re, replacement, globalReplace, source)

					// 关闭文件
					origFile.Close()
					tmpFile.Close()

					// 用临时文件替换原文件
					if err := os.Rename(tempFile, source); err != nil {
						fmt.Printf("错误: 无法替换原文件: %v\n", err)
						os.Remove(tempFile) // 清理临时文件
						continue
					}

					fmt.Printf("已更新文件: %s\n", source)
				} else {
					// 输出到标准输出模式
					file, err := os.Open(source)
					if err != nil {
						fmt.Printf("错误: 无法打开文件 %s: %v\n", source, err)
						continue
					}

					fmt.Printf("==> %s <==\n", source)
					executeReplace(file, os.Stdout, re, replacement, globalReplace, "")
					file.Close()

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
			executeFilter(file, expression, fieldSep, printPattern)

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
	textGrepCmd.Flags().BoolP("line-number", "n", false, "显示行号")
	textGrepCmd.Flags().BoolP("invert-match", "v", false, "反向匹配（显示不匹配的行）")
	textGrepCmd.Flags().BoolP("count", "c", false, "只显示匹配的行数")
	textGrepCmd.Flags().BoolP("color", "", true, "彩色输出匹配部分")
	textGrepCmd.Flags().IntP("context", "C", 0, "显示匹配行前后的上下文行数")

	// replace命令选项
	textReplaceCmd.Flags().BoolP("ignore-case", "i", false, "忽略大小写")
	textReplaceCmd.Flags().BoolP("global", "g", false, "全局替换（每行多次）")
	textReplaceCmd.Flags().BoolP("in-place", "I", false, "原地修改文件")
	textReplaceCmd.Flags().StringP("backup", "b", "", "创建备份，指定备份后缀")

	// filter命令选项
	textFilterCmd.Flags().StringP("field-separator", "F", " ", "字段分隔符")
	textFilterCmd.Flags().StringP("print", "p", "", "输出格式模式")
}

// executeGrep 执行文本搜索
func executeGrep(file *os.File, re *regexp.Regexp, showLineNum, invertMatch, colorOutput bool, contextLines int, sourceName string) int {
	scanner := bufio.NewScanner(file)

	// 彩色输出设置
	matchColor := color.New(color.FgRed, color.Bold).SprintFunc()
	lineNumColor := color.New(color.FgGreen).SprintFunc()
	filenameColor := color.New(color.FgBlue, color.Bold).SprintFunc()

	// 用于存储匹配结果的行和上下文
	type lineInfo struct {
		num     int
		content string
		matched bool
	}

	// 读取所有行
	var lines []lineInfo
	lineNum := 0
	matchCount := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matched := re.MatchString(line)

		if invertMatch {
			matched = !matched
		}

		if matched {
			matchCount++
		}

		lines = append(lines, lineInfo{lineNum, line, matched})
	}

	if scanner.Err() != nil {
		fmt.Printf("读取错误: %v\n", scanner.Err())
		return 0
	}

	// 如果只需要计数，直接返回
	if len(lines) == 0 {
		return 0
	}

	// 显示结果
	if len(sourceName) > 0 && sourceName != "标准输入" {
		fmt.Printf("==> %s <==\n", filenameColor(sourceName))
	}

	// 处理匹配行及其上下文
	for i := 0; i < len(lines); i++ {
		if !lines[i].matched && contextLines == 0 {
			continue // 非匹配行且不需要上下文
		}

		// 检查是否在匹配行的上下文范围内
		inContext := false
		if contextLines > 0 && !lines[i].matched {
			// 检查前后是否有匹配行
			for j := max(0, i-contextLines); j <= min(len(lines)-1, i+contextLines); j++ {
				if lines[j].matched {
					inContext = true
					break
				}
			}
		}

		if lines[i].matched || inContext {
			line := lines[i].content

			// 格式化输出
			if showLineNum {
				fmt.Printf("%s:", lineNumColor(fmt.Sprintf("%d", lines[i].num)))
			}

			if colorOutput && lines[i].matched {
				// 高亮显示匹配部分
				line = re.ReplaceAllStringFunc(line, func(match string) string {
					return matchColor(match)
				})
			}

			fmt.Println(line)
		}
	}

	return matchCount
}

// executeReplace 执行文本替换
func executeReplace(input *os.File, output *os.File, re *regexp.Regexp, replacement string, globalReplace bool, filename string) {
	scanner := bufio.NewScanner(input)
	lineCount := 0
	replaceCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		var newLine string
		if globalReplace {
			// 全局替换（每行多次）
			beforeLen := len(line)
			newLine = re.ReplaceAllString(line, replacement)
			if beforeLen != len(newLine) {
				replaceCount++
			}
		} else {
			// 每行只替换一次
			loc := re.FindStringIndex(line)
			if loc != nil {
				replaceCount++
				newLine = line[:loc[0]] + re.ReplaceAllString(line[loc[0]:loc[1]], replacement) + line[loc[1]:]
			} else {
				newLine = line
			}
		}

		fmt.Fprintln(output, newLine)
	}

	if scanner.Err() != nil {
		fmt.Printf("读取错误: %v\n", scanner.Err())
		return
	}

	if filename != "" {
		fmt.Printf("处理了 %d 行，替换了 %d 处\n", lineCount, replaceCount)
	}
}

// executeFilter 执行文本过滤
func executeFilter(file *os.File, expression string, fieldSep string, printPattern string) {
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		fields := strings.Split(line, fieldSep)

		// 简化的表达式解析和评估（仅支持基本功能）
		if evaluateExpression(expression, line, fields) {
			if printPattern != "" {
				// 使用打印模式格式化输出
				fmt.Println(formatOutput(printPattern, fields))
			} else {
				// 输出原始行
				fmt.Println(line)
			}
		}
	}

	if scanner.Err() != nil {
		fmt.Printf("读取错误: %v\n", scanner.Err())
	}
}

// evaluateExpression 评估过滤表达式
func evaluateExpression(expr string, line string, fields []string) bool {
	// 非常简化的表达式评估，仅支持基本功能
	// 真实实现需要一个合适的表达式解析器

	// 一些基本的模式识别
	expr = strings.TrimSpace(expr)

	// 字段引用：$0, $1, $2...
	for i := 0; i <= len(fields); i++ {
		var value string
		if i == 0 {
			value = line // $0 表示整行
		} else if i <= len(fields) {
			value = fields[i-1] // $1 表示第一个字段
		} else {
			value = ""
		}

		// 替换字段引用
		expr = strings.ReplaceAll(expr, fmt.Sprintf("$%d", i), fmt.Sprintf("\"%s\"", value))
	}

	// 尝试一些简单的比较
	// 等于: ==
	if strings.Contains(expr, "==") {
		parts := strings.Split(expr, "==")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			// 去掉可能的引号
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")
			return left == right
		}
	}

	// 不等于: !=
	if strings.Contains(expr, "!=") {
		parts := strings.Split(expr, "!=")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			// 去掉可能的引号
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")
			return left != right
		}
	}

	// 大于: >
	if strings.Contains(expr, ">") {
		parts := strings.Split(expr, ">")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			// 去掉可能的引号
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")
			// 尝试作为数字比较
			leftNum, leftErr := strconv.ParseFloat(left, 64)
			rightNum, rightErr := strconv.ParseFloat(right, 64)
			if leftErr == nil && rightErr == nil {
				return leftNum > rightNum
			}
			// 否则按字符串比较
			return left > right
		}
	}

	// 小于: <
	if strings.Contains(expr, "<") {
		parts := strings.Split(expr, "<")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			// 去掉可能的引号
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")
			// 尝试作为数字比较
			leftNum, leftErr := strconv.ParseFloat(left, 64)
			rightNum, rightErr := strconv.ParseFloat(right, 64)
			if leftErr == nil && rightErr == nil {
				return leftNum < rightNum
			}
			// 否则按字符串比较
			return left < right
		}
	}

	// 包含: contains
	if strings.Contains(expr, "contains") {
		parts := strings.Split(expr, "contains")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			// 去掉可能的引号
			left = strings.Trim(left, "\"")
			right = strings.Trim(right, "\"")
			return strings.Contains(left, right)
		}
	}

	// 正则匹配: ~
	if strings.Contains(expr, "~") {
		parts := strings.Split(expr, "~")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			// 去掉可能的引号
			left = strings.Trim(left, "\"")
			// 提取正则表达式
			re := strings.Trim(right, "/ ")
			pattern, err := regexp.Compile(re)
			if err == nil {
				return pattern.MatchString(left)
			}
		}
	}

	// 长度函数
	if strings.Contains(expr, "length(") {
		re := regexp.MustCompile(`length\("([^"]*)"\)`)
		matches := re.FindStringSubmatch(expr)
		if len(matches) > 1 {
			lengthStr := fmt.Sprintf("%d", len(matches[1]))
			expr = strings.Replace(expr, matches[0], lengthStr, 1)
			return evaluateExpression(expr, line, fields)
		}
	}

	// 最简单的情况 - 如果表达式是非空字符串，返回true
	if expr != "" && expr != "0" && expr != "false" {
		return true
	}

	return false
}

// formatOutput 根据模式格式化输出
func formatOutput(pattern string, fields []string) string {
	result := pattern

	// 替换${n}形式的字段引用
	re := regexp.MustCompile(`\$\{(\d+)\}`)
	result = re.ReplaceAllStringFunc(result, func(match string) string {
		numStr := re.FindStringSubmatch(match)[1]
		idx, err := strconv.Atoi(numStr)
		if err != nil || idx < 1 || idx > len(fields) {
			return ""
		}
		return fields[idx-1]
	})

	// 替换$n形式的字段引用
	result = regexp.MustCompile(`\$(\d+)`).ReplaceAllStringFunc(result, func(match string) string {
		numStr := match[1:]
		idx, err := strconv.Atoi(numStr)
		if err != nil || idx < 1 || idx > len(fields) {
			return ""
		}
		return fields[idx-1]
	})

	return result
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// min 返回两个整数的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
