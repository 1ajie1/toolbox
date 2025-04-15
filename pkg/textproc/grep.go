package textproc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
)

// GrepOptions 定义了grep命令的选项
type GrepOptions struct {
	Pattern      string
	IgnoreCase   bool
	ShowLineNum  bool
	InvertMatch  bool
	OnlyCount    bool
	ColorOutput  bool
	ContextLines int
	Recursive    bool     // 是否递归搜索目录
	FilePattern  string   // 文件名匹配模式
	ExcludeDirs  []string // 排除的目录
}

// GrepResult 存储grep的结果
type GrepResult struct {
	Matches      int
	TotalLines   int
	MatchedFiles int
}

// ExecuteGrep 执行文本搜索
func ExecuteGrep(input io.Reader, output io.Writer, options GrepOptions, sourceName string) (GrepResult, error) {
	scanner := bufio.NewScanner(input)
	result := GrepResult{}

	// 彩色输出设置
	matchColor := color.New(color.FgRed, color.Bold).SprintFunc()
	lineNumColor := color.New(color.FgGreen).SprintFunc()
	filenameColor := color.New(color.FgBlue, color.Bold).SprintFunc()

	// 编译正则表达式
	var regexpOpt string
	if options.IgnoreCase {
		regexpOpt = "(?i)"
	}
	re, err := regexp.Compile(regexpOpt + options.Pattern)
	if err != nil {
		return result, fmt.Errorf("无效的正则表达式: %v", err)
	}

	// 用于存储匹配结果的行和上下文
	type lineInfo struct {
		num     int
		content string
		matched bool
	}

	// 读取所有行
	var lines []lineInfo
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matched := re.MatchString(line)

		if options.InvertMatch {
			matched = !matched
		}

		if matched {
			result.Matches++
		}

		lines = append(lines, lineInfo{lineNum, line, matched})
	}

	if scanner.Err() != nil {
		return result, fmt.Errorf("读取错误: %v", scanner.Err())
	}

	result.TotalLines = lineNum

	// 如果只需要计数，直接返回
	if options.OnlyCount {
		if result.Matches > 0 {
			fmt.Fprintln(output, result.Matches)
		}
		return result, nil
	}

	// 如果没有匹配项，则不显示任何内容
	if result.Matches == 0 {
		return result, nil
	}

	// 显示结果
	if len(sourceName) > 0 && sourceName != "标准输入" {
		fmt.Fprintf(output, "==> %s <==\n", filenameColor(sourceName))
	}

	// 处理匹配行及其上下文
	for i := 0; i < len(lines); i++ {
		if !lines[i].matched && options.ContextLines == 0 {
			continue // 非匹配行且不需要上下文
		}

		// 检查是否在匹配行的上下文范围内
		inContext := false
		if options.ContextLines > 0 && !lines[i].matched {
			// 检查前后是否有匹配行
			for j := max(0, i-options.ContextLines); j <= min(len(lines)-1, i+options.ContextLines); j++ {
				if lines[j].matched {
					inContext = true
					break
				}
			}
		}

		if lines[i].matched || inContext {
			line := lines[i].content

			// 格式化输出
			if options.ShowLineNum {
				// 改进行号显示，使用右对齐且加粗突出显示
				lineNumStr := fmt.Sprintf("%5d", lines[i].num) // 右对齐5位数
				fmt.Fprintf(output, "%s: ", lineNumColor(lineNumStr))
			}

			if options.ColorOutput && lines[i].matched {
				// 高亮显示匹配部分
				line = re.ReplaceAllStringFunc(line, func(match string) string {
					return matchColor(match)
				})
			}

			fmt.Fprintln(output, line)
		}
	}

	return result, nil
}

// GrepDirectory 在目录中递归查找匹配的文件
func GrepDirectory(dir string, output io.Writer, options GrepOptions) (GrepResult, error) {
	result := GrepResult{}

	// 编译文件名匹配正则（如果有）
	var fileRe *regexp.Regexp
	var err error
	if options.FilePattern != "" {
		fileRe, err = regexp.Compile(options.FilePattern)
		if err != nil {
			return result, fmt.Errorf("无效的文件模式: %v", err)
		}
	}

	// 彩色输出设置
	filenameColor := color.New(color.FgBlue, color.Bold).SprintFunc()

	// 遍历目录
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(output, "警告: 访问 %s 时出错: %v\n", path, err)
			return nil // 继续处理其他文件
		}

		// 跳过目录
		if info.IsDir() {
			// 检查是否是排除的目录
			if isExcludedDir(path, options.ExcludeDirs) {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件名是否匹配模式
		if fileRe != nil && !fileRe.MatchString(info.Name()) {
			return nil
		}

		// 打开文件
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(output, "警告: 无法打开 %s: %v\n", path, err)
			return nil // 继续处理其他文件
		}
		defer file.Close()

		// 搜索文件内容
		tempOptions := options
		// 确保保留文件名和行号显示
		if !tempOptions.OnlyCount {
			tempOptions.ShowLineNum = true
		}
		fileResult, err := ExecuteGrep(file, output, tempOptions, path)
		if err != nil {
			fmt.Fprintf(output, "警告: 处理 %s 时出错: %v\n", path, err)
			return nil
		}

		// 更新总结果
		if fileResult.Matches > 0 {
			result.Matches += fileResult.Matches
			result.MatchedFiles++

			// 如果只需要计数，只输出有匹配的文件名和匹配数
			if options.OnlyCount {
				fmt.Fprintf(output, "%s: %d\n", filenameColor(path), fileResult.Matches)
			}
		}

		return nil
	})

	if err != nil {
		return result, fmt.Errorf("遍历目录错误: %v", err)
	}

	// 打印总结
	if options.OnlyCount {
		fmt.Fprintf(output, "\n共找到 %d 个匹配项，在 %d 个文件中\n", result.Matches, result.MatchedFiles)
	}

	return result, nil
}

// isExcludedDir 检查目录是否应该被排除
func isExcludedDir(path string, excludeDirs []string) bool {
	for _, excludeDir := range excludeDirs {
		if matched, _ := filepath.Match(excludeDir, filepath.Base(path)); matched {
			return true
		}
	}
	return false
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
