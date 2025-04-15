package textproc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

// ReplaceOptions 定义了replace命令的选项
type ReplaceOptions struct {
	Pattern       string
	Replacement   string
	IgnoreCase    bool
	GlobalReplace bool
}

// ReplaceResult 存储替换的结果
type ReplaceResult struct {
	LinesProcessed int
	Replacements   int
}

// ExecuteReplace 执行文本替换
func ExecuteReplace(input io.Reader, output io.Writer, options ReplaceOptions) (ReplaceResult, error) {
	scanner := bufio.NewScanner(input)
	result := ReplaceResult{}

	// 编译正则表达式
	var regexpOpt string
	if options.IgnoreCase {
		regexpOpt = "(?i)"
	}
	re, err := regexp.Compile(regexpOpt + options.Pattern)
	if err != nil {
		return result, fmt.Errorf("无效的正则表达式: %v", err)
	}

	for scanner.Scan() {
		line := scanner.Text()
		result.LinesProcessed++

		var newLine string
		if options.GlobalReplace {
			// 全局替换（每行多次）
			beforeLen := len(line)
			newLine = re.ReplaceAllString(line, options.Replacement)
			if beforeLen != len(newLine) {
				result.Replacements++
			}
		} else {
			// 每行只替换一次
			loc := re.FindStringIndex(line)
			if loc != nil {
				result.Replacements++
				newLine = line[:loc[0]] + re.ReplaceAllString(line[loc[0]:loc[1]], options.Replacement) + line[loc[1]:]
			} else {
				newLine = line
			}
		}

		fmt.Fprintln(output, newLine)
	}

	if scanner.Err() != nil {
		return result, fmt.Errorf("读取错误: %v", scanner.Err())
	}

	return result, nil
}

// CreateBackup 创建文件备份
func CreateBackup(srcPath, suffix string) error {
	dstPath := srcPath + suffix
	return CopyFile(srcPath, dstPath)
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
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