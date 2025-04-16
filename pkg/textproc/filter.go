package textproc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// FilterOptions 定义文本过滤的配置选项
type FilterOptions struct {
	Expression   string // 过滤表达式
	FieldSep     string // 字段分隔符
	PrintPattern string // 打印模式
}

// FilterResult 存储过滤操作的结果
type FilterResult struct {
	LinesProcessed int // 处理的总行数
	Matches        int // 匹配的行数
}

// 过滤条件类型
const (
	opEquals     = "=="
	opNotEquals  = "!="
	opGreater    = ">"
	opLess       = "<"
	opGreaterEq  = ">="
	opLessEq     = "<="
	opContains   = "contains"
	opStartsWith = "startswith"
	opEndsWith   = "endswith"
	opMatches    = "~"
	opNotMatches = "!~"
	opLength     = "length"
)

// ExecuteFilter 执行文本过滤操作
func ExecuteFilter(input io.Reader, output io.Writer, options FilterOptions) (FilterResult, error) {
	if options.Expression == "" {
		return FilterResult{}, errors.New("必须指定过滤表达式")
	}

	scanner := bufio.NewScanner(input)
	result := FilterResult{}

	for scanner.Scan() {
		line := scanner.Text()
		result.LinesProcessed++

		// 解析行并应用过滤条件
		fields := parseFields(line, options.FieldSep)
		match, err := evaluateExpression(options.Expression, line, fields)
		if err != nil {
			return result, fmt.Errorf("行 %d: %v", result.LinesProcessed, err)
		}

		if match {
			result.Matches++
			if options.PrintPattern != "" {
				// 应用打印模式
				formattedOutput, err := applyPrintPattern(options.PrintPattern, line, fields)
				if err != nil {
					return result, fmt.Errorf("应用打印模式时出错：%v", err)
				}
				fmt.Fprintln(output, formattedOutput)
			} else {
				fmt.Fprintln(output, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("读取输入时出错：%v", err)
	}

	return result, nil
}

// parseFields 将一行文本分割为字段
func parseFields(line, sep string) []string {
	if sep == "" {
		sep = " "
	}
	fields := strings.Split(line, sep)
	// 清理字段，去除多余空格
	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}
	return fields
}

// evaluateExpression 评估过滤表达式
func evaluateExpression(expr, line string, fields []string) (bool, error) {
	// 检查是否是复合表达式（包含 && 或 ||）
	if strings.Contains(expr, "&&") {
		subExprs := strings.Split(expr, "&&")
		for _, subExpr := range subExprs {
			match, err := evaluateExpression(strings.TrimSpace(subExpr), line, fields)
			if err != nil {
				return false, err
			}
			if !match {
				return false, nil // AND 操作，一个不匹配则整体不匹配
			}
		}
		return true, nil
	}

	if strings.Contains(expr, "||") {
		subExprs := strings.Split(expr, "||")
		for _, subExpr := range subExprs {
			match, err := evaluateExpression(strings.TrimSpace(subExpr), line, fields)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil // OR 操作，一个匹配则整体匹配
			}
		}
		return false, nil
	}

	// 处理 length($0) 等函数
	if strings.HasPrefix(expr, "length(") && strings.HasSuffix(expr, ")") {
		// 提取字段索引
		fieldExpr := expr[7 : len(expr)-1]
		fieldValue, err := getFieldValue(fieldExpr, line, fields)
		if err != nil {
			return false, err
		}

		// 提取比较运算符和值
		parts := strings.Fields(expr)
		if len(parts) >= 3 {
			op := parts[1]
			valueStr := parts[2]
			length := len(fieldValue)
			value, err := strconv.Atoi(valueStr)
			if err != nil {
				return false, fmt.Errorf("无效的长度比较值：%s", valueStr)
			}

			switch op {
			case opEquals:
				return length == value, nil
			case opNotEquals:
				return length != value, nil
			case opGreater:
				return length > value, nil
			case opLess:
				return length < value, nil
			case opGreaterEq:
				return length >= value, nil
			case opLessEq:
				return length <= value, nil
			default:
				return false, fmt.Errorf("不支持的长度比较操作符：%s", op)
			}
		}
		return false, fmt.Errorf("无效的length表达式：%s", expr)
	}

	// 解析基本的比较表达式
	var fieldExpr, op, valueExpr string

	// 检查正则表达式匹配
	if regexMatch := regexp.MustCompile(`(\$\d+)\s*([~!~])\s*/(.*)/`).FindStringSubmatch(expr); len(regexMatch) == 4 {
		fieldExpr = regexMatch[1]
		op = regexMatch[2]
		valueExpr = regexMatch[3]

		fieldValue, err := getFieldValue(fieldExpr, line, fields)
		if err != nil {
			return false, err
		}

		regex, err := regexp.Compile(valueExpr)
		if err != nil {
			return false, fmt.Errorf("无效的正则表达式：%s", valueExpr)
		}

		if op == opMatches {
			return regex.MatchString(fieldValue), nil
		} else if op == opNotMatches {
			return !regex.MatchString(fieldValue), nil
		}
	}

	// 检查其他操作符
	for _, operator := range []string{opEquals, opNotEquals, opGreaterEq, opLessEq, opGreater, opLess, opContains, opStartsWith, opEndsWith} {
		if parts := strings.SplitN(expr, operator, 2); len(parts) == 2 {
			fieldExpr = strings.TrimSpace(parts[0])
			op = operator
			valueExpr = strings.TrimSpace(parts[1])
			break
		}
	}

	if op == "" {
		return false, fmt.Errorf("无效的表达式：%s", expr)
	}

	fieldValue, err := getFieldValue(fieldExpr, line, fields)
	if err != nil {
		return false, err
	}

	// 处理值表达式（可能是字面量或字段引用）
	value := valueExpr
	if strings.HasPrefix(valueExpr, "$") {
		value, err = getFieldValue(valueExpr, line, fields)
		if err != nil {
			return false, err
		}
	} else if strings.HasPrefix(valueExpr, "\"") && strings.HasSuffix(valueExpr, "\"") {
		// 处理引号包围的字符串
		value = valueExpr[1 : len(valueExpr)-1]
	}

	// 执行比较
	switch op {
	case opEquals:
		return fieldValue == value, nil
	case opNotEquals:
		return fieldValue != value, nil
	case opContains:
		return strings.Contains(fieldValue, value), nil
	case opStartsWith:
		return strings.HasPrefix(fieldValue, value), nil
	case opEndsWith:
		return strings.HasSuffix(fieldValue, value), nil
	case opGreater, opLess, opGreaterEq, opLessEq:
		// 尝试数值比较
		fieldNum, fieldErr := strconv.ParseFloat(fieldValue, 64)
		valueNum, valueErr := strconv.ParseFloat(value, 64)

		if fieldErr != nil || valueErr != nil {
			// 如果不能转换为数字，则进行字符串比较
			switch op {
			case opGreater:
				return fieldValue > value, nil
			case opLess:
				return fieldValue < value, nil
			case opGreaterEq:
				return fieldValue >= value, nil
			case opLessEq:
				return fieldValue <= value, nil
			}
		} else {
			// 数值比较
			switch op {
			case opGreater:
				return fieldNum > valueNum, nil
			case opLess:
				return fieldNum < valueNum, nil
			case opGreaterEq:
				return fieldNum >= valueNum, nil
			case opLessEq:
				return fieldNum <= valueNum, nil
			}
		}
	}

	return false, fmt.Errorf("不支持的操作符：%s", op)
}

// getFieldValue 从字段列表中获取指定字段的值
func getFieldValue(fieldExpr, line string, fields []string) (string, error) {
	if fieldExpr == "$0" {
		return line, nil
	}

	if strings.HasPrefix(fieldExpr, "$") {
		fieldIndex, err := strconv.Atoi(fieldExpr[1:])
		if err != nil {
			return "", fmt.Errorf("无效的字段索引：%s", fieldExpr)
		}

		if fieldIndex < 1 || fieldIndex > len(fields) {
			return "", fmt.Errorf("字段索引超出范围：%d（总字段数：%d）", fieldIndex, len(fields))
		}

		return fields[fieldIndex-1], nil
	}

	return fieldExpr, nil
}

// applyPrintPattern 应用打印模式格式化输出
func applyPrintPattern(pattern, line string, fields []string) (string, error) {
	result := pattern

	// 替换字段引用，如 ${1}, ${2} 等
	fieldPattern := regexp.MustCompile(`\$\{(\d+)\}`)
	result = fieldPattern.ReplaceAllStringFunc(result, func(match string) string {
		idxStr := match[2 : len(match)-1]
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return match
		}

		if idx == 0 {
			return line
		}

		if idx < 1 || idx > len(fields) {
			return "" // 超出范围的字段返回空字符串
		}

		return fields[idx-1]
	})

	// 替换简单字段引用，如 $1, $2 等
	simpleFieldPattern := regexp.MustCompile(`\$(\d+)`)
	result = simpleFieldPattern.ReplaceAllStringFunc(result, func(match string) string {
		idxStr := match[1:]
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return match
		}

		if idx == 0 {
			return line
		}

		if idx < 1 || idx > len(fields) {
			return "" // 超出范围的字段返回空字符串
		}

		return fields[idx-1]
	})

	return result, nil
}
