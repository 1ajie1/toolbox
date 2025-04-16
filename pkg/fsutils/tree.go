package fsutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TreeOptions 表示目录树显示的选项
type TreeOptions struct {
	MaxDepth      int  // 最大显示深度，0表示不限制
	ShowHidden    bool // 是否显示隐藏文件
	OnlyDirs      bool // 是否只显示目录
	FollowSymlink bool // 是否跟踪符号链接
	ShowSize      bool // 是否显示文件大小
}

// TreeResult 表示目录树显示的结果
type TreeResult struct {
	DirCount  int // 目录数量
	FileCount int // 文件数量
}

// DisplayTree 显示指定目录的文件树结构
func DisplayTree(root string, writer io.Writer, options TreeOptions) (TreeResult, error) {
	result := TreeResult{
		DirCount:  0,
		FileCount: 0,
	}

	// 检查目录是否存在
	fi, err := os.Stat(root)
	if err != nil {
		return result, fmt.Errorf("无法访问目录 %s: %v", root, err)
	}

	if !fi.IsDir() {
		return result, fmt.Errorf("%s 不是一个目录", root)
	}

	// 显示根目录
	absPath, err := filepath.Abs(root)
	if err != nil {
		absPath = root
	}
	fmt.Fprintf(writer, "%s\n", absPath)

	// 存储已访问的目录路径，避免符号链接造成的循环
	visited := make(map[string]bool)
	visited[absPath] = true

	// 开始递归显示目录树
	err = displayTreeNode(root, "", writer, options, &result, 1, visited)
	if err != nil {
		return result, err
	}

	// 显示统计信息
	fmt.Fprintf(writer, "\n%d 个目录", result.DirCount)
	if !options.OnlyDirs {
		fmt.Fprintf(writer, "，%d 个文件", result.FileCount)
	}
	fmt.Fprintln(writer)

	return result, nil
}

// displayTreeNode 递归显示目录树节点
func displayTreeNode(path string, prefix string, writer io.Writer, options TreeOptions, result *TreeResult, depth int, visited map[string]bool) error {
	// 检查最大深度限制
	if options.MaxDepth > 0 && depth > options.MaxDepth {
		return nil
	}

	// 读取目录内容
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("无法读取目录 %s: %v", path, err)
	}

	// 过滤和排序条目
	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		name := entry.Name()

		// 是否跳过隐藏文件
		if !options.ShowHidden && strings.HasPrefix(name, ".") {
			continue
		}

		// 是否只显示目录
		if options.OnlyDirs && !entry.IsDir() {
			continue
		}

		filteredEntries = append(filteredEntries, entry)
	}

	// 按名称排序
	sort.Slice(filteredEntries, func(i, j int) bool {
		return filteredEntries[i].Name() < filteredEntries[j].Name()
	})

	// 处理每个条目
	for i, entry := range filteredEntries {
		// 确定是否是最后一个条目
		isLast := i == len(filteredEntries)-1

		// 确定当前行的前缀和下一级的前缀
		var currentPrefix, nextPrefix string
		if isLast {
			currentPrefix = prefix + "└── "
			nextPrefix = prefix + "    "
		} else {
			currentPrefix = prefix + "├── "
			nextPrefix = prefix + "│   "
		}

		// 获取完整路径
		entryPath := filepath.Join(path, entry.Name())

		// 处理符号链接
		isSymlink := entry.Type()&os.ModeSymlink != 0
		var linkTarget string
		if isSymlink {
			if target, err := os.Readlink(entryPath); err == nil {
				linkTarget = " -> " + target
			}
		}

		// 显示大小
		sizeStr := ""
		if options.ShowSize && !entry.IsDir() {
			if info, err := entry.Info(); err == nil {
				size := info.Size()
				if size < 1024 {
					sizeStr = fmt.Sprintf(" [%d B]", size)
				} else if size < 1024*1024 {
					sizeStr = fmt.Sprintf(" [%.1f KB]", float64(size)/1024)
				} else if size < 1024*1024*1024 {
					sizeStr = fmt.Sprintf(" [%.1f MB]", float64(size)/(1024*1024))
				} else {
					sizeStr = fmt.Sprintf(" [%.1f GB]", float64(size)/(1024*1024*1024))
				}
			}
		}

		// 输出当前节点
		fmt.Fprintf(writer, "%s%s%s%s\n", currentPrefix, entry.Name(), linkTarget, sizeStr)

		// 递归处理子目录
		if entry.IsDir() {
			result.DirCount++

			// 处理符号链接指向的目录
			if isSymlink && options.FollowSymlink {
				realPath, err := filepath.EvalSymlinks(entryPath)
				if err != nil {
					continue
				}

				// 避免循环引用
				if visited[realPath] {
					continue
				}
				visited[realPath] = true

				if err := displayTreeNode(realPath, nextPrefix, writer, options, result, depth+1, visited); err != nil {
					// 继续处理，忽略错误
					continue
				}
			} else if !isSymlink {
				if err := displayTreeNode(entryPath, nextPrefix, writer, options, result, depth+1, visited); err != nil {
					// 继续处理，忽略错误
					continue
				}
			}
		} else if !entry.IsDir() && !options.OnlyDirs {
			result.FileCount++
		}
	}

	return nil
}
