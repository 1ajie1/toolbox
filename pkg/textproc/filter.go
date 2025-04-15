package textproc

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// FilterOptions 定义了filter命令的选项
type FilterOptions struct {
	Expression    string
	FieldSep      string
	PrintPattern  string
}

// FilterResult 存储过滤的结果
type FilterResult struct {
	LinesProcessed int
	LinesOutput    int
}

// ExecuteFilter 执行文本过滤
func ExecuteFilter(input io.Reader, output io.Writer, options FilterOptions) (FilterResult, error) {
	scanner := bufio.NewScanner(input)
	result := FilterResult{}

	for scanner.Scan() {
		result.LinesProcessed++
		line := scanner.Text()
		fields := strings.Split(line, options.FieldSep)

		// 评估表达式
		if evaluateExpression(options.Expression, line, fields) {
			result.LinesOutput++
			if options.PrintPattern != "" {
				// 使用打印模式格式化输出
				fmt.Fprintln(output, formatOutput(options.PrintPattern, fields))
			} else {
				// 输出原始行
				fmt.Fprintln(output, line)
			}
		}
	}

	if scanner.Err() != nil {
		return result, fmt.Errorf("读取错误: %v", scanner.Err())
	}

	return result, nil
}

// evaluateExpression 评估过滤表达式
func evaluateExpression(expr string, line string, fields []string) bool {
	// 简化的表达式评估，仅支持基本功能
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