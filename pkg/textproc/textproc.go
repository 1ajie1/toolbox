package textproc

import (
	"io"
	"os"
)

// TextProcessorInterface 定义文本处理器的接口
type TextProcessorInterface interface {
	// Grep 在文本中搜索匹配的模式
	Grep(input io.Reader, output io.Writer, options GrepOptions, sourceName string) (GrepResult, error)

	// GrepDir 在目录中递归搜索匹配的模式
	GrepDir(dir string, output io.Writer, options GrepOptions) (GrepResult, error)

	// Replace 替换文本内容
	Replace(input io.Reader, output io.Writer, options ReplaceOptions) (ReplaceResult, error)

	// Filter 过滤文本行
	Filter(input io.Reader, output io.Writer, options FilterOptions) (FilterResult, error)
}

// TextProcessor 实现了TextProcessorInterface接口
type TextProcessor struct{}

// NewTextProcessor 创建一个新的文本处理器
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{}
}

// Grep 执行文本搜索，实现接口方法
func (p *TextProcessor) Grep(input io.Reader, output io.Writer, options GrepOptions, sourceName string) (GrepResult, error) {
	return ExecuteGrep(input, output, options, sourceName)
}

// GrepDir 在目录中递归搜索，实现接口方法
func (p *TextProcessor) GrepDir(dir string, output io.Writer, options GrepOptions) (GrepResult, error) {
	return GrepDirectory(dir, output, options)
}

// Replace 执行文本替换，实现接口方法
func (p *TextProcessor) Replace(input io.Reader, output io.Writer, options ReplaceOptions) (ReplaceResult, error) {
	return ExecuteReplace(input, output, options)
}

// Filter 执行文本过滤，实现接口方法
func (p *TextProcessor) Filter(input io.Reader, output io.Writer, options FilterOptions) (FilterResult, error) {
	return ExecuteFilter(input, output, options)
}

// 以下是一些便捷的函数，直接调用对应的实现

// Grep 文本搜索便捷函数
func Grep(input io.Reader, output io.Writer, options GrepOptions, sourceName string) (GrepResult, error) {
	return ExecuteGrep(input, output, options, sourceName)
}

// GrepDir 目录递归搜索便捷函数
func GrepDir(dir string, output io.Writer, options GrepOptions) (GrepResult, error) {
	return GrepDirectory(dir, output, options)
}

// Replace 文本替换便捷函数
func Replace(input io.Reader, output io.Writer, options ReplaceOptions) (ReplaceResult, error) {
	return ExecuteReplace(input, output, options)
}

// Filter 文本过滤便捷函数
func Filter(input io.Reader, output io.Writer, options FilterOptions) (FilterResult, error) {
	return ExecuteFilter(input, output, options)
}

// ProcessFile 处理文件的通用函数
func ProcessFile(filePath string, processor func(io.Reader, io.Writer) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return processor(file, os.Stdout)
}
