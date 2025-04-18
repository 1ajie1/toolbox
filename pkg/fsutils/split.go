package fsutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// SplitOptions 分片选项
type SplitOptions struct {
	SourceDir    string         // 源目录
	OutputDir    string         // 输出目录
	ChunkSize    int64          // 分片大小（字节）
	CompressType CompressFormat // 压缩类型
	ThreadCount  int            // 线程数
	DeleteSource bool           // 是否删除源文件
}

// validateSplitOptions 验证分片选项
func validateSplitOptions(opts *SplitOptions) error {
	// 检查源目录
	if opts.SourceDir == "" {
		return fmt.Errorf("源目录不能为空")
	}
	if _, err := os.Stat(opts.SourceDir); err != nil {
		return fmt.Errorf("源目录不存在: %v", err)
	}

	// 检查输出目录
	if opts.OutputDir == "" {
		opts.OutputDir = opts.SourceDir + "_chunks"
	}

	// 检查分片大小
	if opts.ChunkSize <= 0 {
		return fmt.Errorf("分片大小必须大于0")
	}

	// 检查线程数
	maxThreads := runtime.NumCPU()
	if opts.ThreadCount <= 0 {
		opts.ThreadCount = maxThreads
	} else if opts.ThreadCount > maxThreads {
		opts.ThreadCount = maxThreads
	}

	return nil
}

// SplitArchive 将目录打包并分片
func SplitArchive(opts *SplitOptions) error {
	// 验证选项
	if err := validateSplitOptions(opts); err != nil {
		return err
	}

	// 创建输出目录
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 生成临时压缩文件名和基础文件名
	tempArchive := filepath.Join(opts.OutputDir, "temp_archive")
	baseFileName := filepath.Base(opts.SourceDir)
	switch opts.CompressType {
	case ZIP:
		tempArchive += ".zip"
		baseFileName += ".zip"
	case TARGZ:
		tempArchive += ".tar.gz"
		baseFileName += ".tar.gz"
	case TARBZ2:
		tempArchive += ".tar.bz2"
		baseFileName += ".tar.bz2"
	case TARXZ:
		tempArchive += ".tar.xz"
		baseFileName += ".tar.xz"
	default:
		return fmt.Errorf("不支持的压缩格式: %v", opts.CompressType)
	}

	// 先将目录压缩
	if err := Compress(opts.SourceDir, tempArchive, CompressOptions{
		Format: opts.CompressType,
		Level:  6, // 使用默认压缩级别
	}); err != nil {
		return fmt.Errorf("压缩失败: %v", err)
	}
	defer os.Remove(tempArchive) // 最后清理临时文件

	// 获取压缩文件大小
	stat, err := os.Stat(tempArchive)
	if err != nil {
		return fmt.Errorf("获取压缩文件大小失败: %v", err)
	}

	// 计算分片数量
	totalSize := stat.Size()
	chunkCount := (totalSize + opts.ChunkSize - 1) / opts.ChunkSize

	// 准备任务通道
	type chunkTask struct {
		index int
		start int64
		size  int64
	}
	tasks := make(chan chunkTask, chunkCount)
	errors := make(chan error, chunkCount)
	var wg sync.WaitGroup

	// 创建工作线程
	for i := 0; i < opts.ThreadCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				if err := splitChunk(tempArchive, opts.OutputDir, baseFileName, task.index, task.start, task.size); err != nil {
					errors <- fmt.Errorf("分片 %d 处理失败: %v", task.index, err)
					return
				}
			}
		}()
	}

	// 分发任务
	for i := int64(0); i < chunkCount; i++ {
		start := i * opts.ChunkSize
		size := opts.ChunkSize
		if start+size > totalSize {
			size = totalSize - start
		}
		tasks <- chunkTask{
			index: int(i + 1),
			start: start,
			size:  size,
		}
	}
	close(tasks)

	// 等待所有任务完成
	wg.Wait()
	close(errors)

	// 检查是否有错误
	if err := <-errors; err != nil {
		return err
	}

	// 如果需要删除源目录
	if opts.DeleteSource {
		if err := os.RemoveAll(opts.SourceDir); err != nil {
			return fmt.Errorf("删除源目录失败: %v", err)
		}
	}

	return nil
}

// splitChunk 处理单个分片
func splitChunk(srcFile, outDir, baseFileName string, index int, start, size int64) error {
	// 打开源文件
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	// 创建分片文件
	chunkFile := filepath.Join(outDir, fmt.Sprintf("%s.%03d", baseFileName, index))
	dst, err := os.Create(chunkFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	// 定位到起始位置
	if _, err := src.Seek(start, 0); err != nil {
		return err
	}

	// 复制指定大小的数据
	written, err := io.CopyN(dst, src, size)
	if err != nil && err != io.EOF {
		return err
	}
	if written != size {
		return fmt.Errorf("写入大小不匹配：期望 %d，实际 %d", size, written)
	}

	return nil
}

// MergeChunks 合并分片文件
func MergeChunks(chunksDir string, outputFile string, deleteChunks bool) error {
	// 打开输出文件
	dst, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer dst.Close()

	// 获取所有分片文件
	pattern := filepath.Join(chunksDir, "*.[0-9][0-9][0-9]")
	chunks, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("查找分片文件失败: %v", err)
	}
	if len(chunks) == 0 {
		return fmt.Errorf("未找到分片文件")
	}

	// 按序号排序分片文件
	sortChunks(chunks)

	// 依次合并分片
	buffer := make([]byte, 1024*1024) // 1MB缓冲区
	for _, chunk := range chunks {
		// 打开分片文件
		src, err := os.Open(chunk)
		if err != nil {
			return fmt.Errorf("打开分片文件失败: %v", err)
		}

		// 复制内容
		_, err = io.CopyBuffer(dst, src, buffer)
		src.Close()
		if err != nil {
			return fmt.Errorf("合并分片失败: %v", err)
		}

		// 如果需要删除分片
		if deleteChunks {
			if err := os.Remove(chunk); err != nil {
				return fmt.Errorf("删除分片文件失败: %v", err)
			}
		}
	}

	// 如果需要删除分片目录
	if deleteChunks {
		if err := os.Remove(chunksDir); err != nil {
			return fmt.Errorf("删除分片目录失败: %v", err)
		}
	}

	return nil
}

// sortChunks 对分片文件按序号排序
func sortChunks(chunks []string) {
	// 使用冒泡排序（因为通常分片数量不会太多）
	for i := 0; i < len(chunks)-1; i++ {
		for j := 0; j < len(chunks)-i-1; j++ {
			if filepath.Base(chunks[j]) > filepath.Base(chunks[j+1]) {
				chunks[j], chunks[j+1] = chunks[j+1], chunks[j]
			}
		}
	}
}
