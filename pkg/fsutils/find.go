package fsutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FindOptions 定义文件搜索的选项
type FindOptions struct {
	Name           string    // 文件名模式（支持通配符）
	Type           string    // 文件类型（f:文件, d:目录, l:符号链接）
	MinSize        int64     // 最小文件大小（字节）
	MaxSize        int64     // 最大文件大小（字节）
	MinDepth       int       // 最小搜索深度
	MaxDepth       int       // 最大搜索深度
	ModifiedAfter  time.Time // 在此时间后修改
	ModifiedBefore time.Time // 在此时间前修改
	Regex          string    // 正则表达式匹配文件名
	ExcludeDirs    []string  // 要排除的目录
	IncludeDirs    []string  // 要包含的目录（为空则搜索所有目录）
	FollowSymlinks bool      // 是否跟随符号链接
}

// FindResult 存储搜索结果
type FindResult struct {
	Path     string      // 文件路径
	FileInfo os.FileInfo // 文件信息
	Depth    int         // 相对于起始目录的深度
}

// ExecuteFind 执行文件搜索
func ExecuteFind(root string, output io.Writer, options FindOptions) error {
	// 编译正则表达式（如果提供）
	var re *regexp.Regexp
	var err error
	if options.Regex != "" {
		re, err = regexp.Compile(options.Regex)
		if err != nil {
			return fmt.Errorf("无效的正则表达式: %v", err)
		}
	}

	// 创建通配符模式（如果提供）
	var pattern string
	if options.Name != "" {
		pattern = options.Name
	}

	// 规范化根目录路径
	root, err = filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %v", err)
	}

	// 规范化包含目录路径
	normalizedIncludeDirs := make([]string, 0, len(options.IncludeDirs))
	for _, dir := range options.IncludeDirs {
		// 移除路径中的 ./ 和 .\ 前缀
		dir = strings.TrimPrefix(dir, "./")
		dir = strings.TrimPrefix(dir, ".\\")
		// 移除末尾的斜杠
		dir = strings.TrimRight(dir, "/\\")
		normalizedIncludeDirs = append(normalizedIncludeDirs, dir)
	}

	// 遍历目录
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(output, "警告: 访问 %s 时出错: %v\n", path, err)
			return nil // 继续处理其他文件
		}

		// 计算相对于根目录的路径
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		// 将相对路径规范化（统一使用正斜杠）
		relPath = filepath.ToSlash(relPath)

		// 计算深度
		depth := 0
		if relPath != "." {
			depth = len(strings.Split(relPath, "/"))
		}

		// 检查深度限制
		if options.MinDepth > 0 && depth < options.MinDepth {
			return nil
		}
		if options.MaxDepth > 0 && depth > options.MaxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查是否应该排除此目录
		if info.IsDir() && isExcludedDir(path, options.ExcludeDirs) {
			return filepath.SkipDir
		}

		// 检查是否在指定的包含目录中
		if len(normalizedIncludeDirs) > 0 {
			// 如果是根目录，允许继续搜索
			if path == root {
				return nil
			}

			isIncluded := false

			if info.IsDir() {
				// 检查当前目录是否匹配任何包含模式
				currentDirPath := filepath.ToSlash(filepath.Clean(relPath))
				for _, includeDir := range normalizedIncludeDirs {
					includePattern := filepath.ToSlash(filepath.Clean(includeDir))
					if matched, _ := filepath.Match(includePattern, currentDirPath); matched {
						isIncluded = true
						break
					}
					// 检查是否是目标目录的父目录
					if strings.HasPrefix(includePattern, currentDirPath) {
						return nil // 继续搜索
					}
				}
				if !isIncluded {
					return filepath.SkipDir
				}
			} else {
				// 对于文件，检查其所在目录是否匹配
				dirPath := filepath.Dir(relPath)
				if dirPath == "." {
					// 检查根目录下的文件是否应该包含
					for _, includeDir := range normalizedIncludeDirs {
						if includeDir == "." {
							isIncluded = true
							break
						}
					}
				} else {
					dirPath = filepath.ToSlash(filepath.Clean(dirPath))
					for _, includeDir := range normalizedIncludeDirs {
						includePattern := filepath.ToSlash(filepath.Clean(includeDir))
						if strings.HasPrefix(dirPath, includePattern) {
							isIncluded = true
							break
						}
					}
				}
				if !isIncluded {
					return nil
				}
			}
		}

		// 检查文件类型
		if options.Type != "" {
			isMatch := false
			switch options.Type {
			case "f":
				isMatch = !info.IsDir()
			case "d":
				isMatch = info.IsDir()
			case "l":
				isMatch = info.Mode()&os.ModeSymlink != 0
			}
			if !isMatch {
				return nil
			}
		}

		// 检查文件大小
		if !info.IsDir() {
			size := info.Size()
			if options.MinSize > 0 && size < options.MinSize {
				return nil
			}
			if options.MaxSize > 0 && size > options.MaxSize {
				return nil
			}
		}

		// 检查修改时间
		modTime := info.ModTime()
		if !options.ModifiedAfter.IsZero() && modTime.Before(options.ModifiedAfter) {
			return nil
		}
		if !options.ModifiedBefore.IsZero() && modTime.After(options.ModifiedBefore) {
			return nil
		}

		// 检查文件名模式
		if pattern != "" {
			matched, err := filepath.Match(pattern, info.Name())
			if err != nil || !matched {
				return nil
			}
		}

		// 检查正则表达式
		if re != nil && !re.MatchString(info.Name()) {
			return nil
		}

		// 输出结果
		fmt.Fprintln(output, path)

		return nil
	})

	return err
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

// FormatSize 格式化文件大小
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
