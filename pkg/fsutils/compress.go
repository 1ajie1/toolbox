package fsutils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsnet/compress/bzip2"
	"github.com/nwaples/rardecode"
	"github.com/saracen/go7z"
	"github.com/ulikunitz/xz"
)

// CompressFormat 定义压缩格式类型
type CompressFormat string

const (
	ZIP      CompressFormat = "zip"
	TARGZ    CompressFormat = "tar.gz"
	TARBZ2   CompressFormat = "tar.bz2"
	TARXZ    CompressFormat = "tar.xz"
	GZ       CompressFormat = "gz"
	BZ2      CompressFormat = "bz2"
	XZ       CompressFormat = "xz"
	RAR      CompressFormat = "rar" // 仅支持解压缩
	SEVENZIP CompressFormat = "7z"
)

// CompressOptions 定义压缩选项
type CompressOptions struct {
	Format       CompressFormat // 压缩格式
	Level        int            // 压缩级别（1-9，0表示默认）
	ExcludePaths []string       // 要排除的路径列表
}

// shouldExclude 检查路径是否应该被排除
func shouldExclude(path string, excludePaths []string) bool {
	if len(excludePaths) == 0 {
		return false
	}

	// 获取路径的绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, excludePath := range excludePaths {
		// 获取排除路径的绝对路径
		absExcludePath, err := filepath.Abs(excludePath)
		if err != nil {
			continue
		}

		// 检查路径是否匹配或是排除路径的子目录
		if absPath == absExcludePath || strings.HasPrefix(absPath, absExcludePath+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// Compress 压缩文件或目录
func Compress(src string, dst string, options CompressOptions) error {
	// 检查源路径是否存在
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("无法访问源文件/目录: %v", err)
	}

	// 根据不同格式调用相应的压缩函数
	switch options.Format {
	case ZIP:
		return compressZip(src, dst, srcInfo.IsDir(), options)
	case TARGZ:
		return compressTarGz(src, dst, srcInfo.IsDir(), options)
	case TARBZ2:
		return compressTarBz2(src, dst, srcInfo.IsDir(), options)
	case TARXZ:
		return compressTarXz(src, dst, srcInfo.IsDir(), options)
	case GZ:
		if srcInfo.IsDir() {
			return fmt.Errorf("gz格式不支持压缩目录")
		}
		return compressGz(src, dst)
	case BZ2:
		if srcInfo.IsDir() {
			return fmt.Errorf("bz2格式不支持压缩目录")
		}
		return compressBz2(src, dst)
	case XZ:
		if srcInfo.IsDir() {
			return fmt.Errorf("xz格式不支持压缩目录")
		}
		return compressXz(src, dst)
	case RAR:
		return fmt.Errorf("RAR格式仅支持解压缩，不支持压缩（因为是专有格式）")
	case SEVENZIP:
		return compress7z()
	default:
		return fmt.Errorf("不支持的压缩格式: %s", options.Format)
	}
}

// Decompress 解压缩文件
func Decompress(src string, dst string) error {
	// 检查源文件是否存在
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("无法访问压缩文件: %v", err)
	}

	// 创建目标目录（如果不存在）
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("无法创建目标目录: %v", err)
	}

	// 根据文件扩展名判断压缩格式
	switch {
	case strings.HasSuffix(src, ".zip"):
		return decompressZip(src, dst)
	case strings.HasSuffix(src, ".tar.gz"), strings.HasSuffix(src, ".tgz"):
		return decompressTarGz(src, dst)
	case strings.HasSuffix(src, ".tar.bz2"), strings.HasSuffix(src, ".tbz2"):
		return decompressTarBz2(src, dst)
	case strings.HasSuffix(src, ".tar.xz"), strings.HasSuffix(src, ".txz"):
		return decompressTarXz(src, dst)
	case strings.HasSuffix(src, ".gz"):
		return decompressGz(src, dst)
	case strings.HasSuffix(src, ".bz2"):
		return decompressBz2(src, dst)
	case strings.HasSuffix(src, ".xz"):
		return decompressXz(src, dst)
	case strings.HasSuffix(src, ".rar"):
		return decompressRar(src, dst)
	case strings.HasSuffix(src, ".7z"):
		return decompress7z(src, dst)
	default:
		return fmt.Errorf("无法识别的压缩格式")
	}
}

// compressZip 创建zip压缩文件
func compressZip(src, dst string, isDir bool, options CompressOptions) error {
	zipfile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	if isDir {
		// 遍历目录
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 检查是否应该排除此路径
			if shouldExclude(path, options.ExcludePaths) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// 获取相对路径
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// 设置相对路径
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(writer, file)
				if err != nil {
					return err
				}
			}
			return err
		})
	} else {
		// 压缩单个文件
		if shouldExclude(src, options.ExcludePaths) {
			return nil
		}

		file, err := os.Open(src)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := archive.Create(filepath.Base(src))
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	}
}

// compressTarGz 创建tar.gz压缩文件
func compressTarGz(src, dst string, isDir bool, options CompressOptions) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	if isDir {
		// 遍历目录
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 检查是否应该排除此路径
			if shouldExclude(path, options.ExcludePaths) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// 获取相对路径
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}

			// 创建tar头部
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(tw, file)
				if err != nil {
					return err
				}
			}
			return nil
		})
	} else {
		// 压缩单个文件
		if shouldExclude(src, options.ExcludePaths) {
			return nil
		}

		file, err := os.Open(src)
		if err != nil {
			return err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.Base(src)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		_, err = io.Copy(tw, file)
		return err
	}
}

// compressTarBz2 创建tar.bz2压缩文件
func compressTarBz2(src, dst string, isDir bool, options CompressOptions) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	bz2w, err := bzip2.NewWriter(file, nil)
	if err != nil {
		return err
	}
	defer bz2w.Close()

	tw := tar.NewWriter(bz2w)
	defer tw.Close()

	if isDir {
		// 遍历目录
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 检查是否应该排除此路径
			if shouldExclude(path, options.ExcludePaths) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// 获取相对路径
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}

			// 创建tar头部
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(tw, file)
				if err != nil {
					return err
				}
			}
			return nil
		})
	} else {
		// 压缩单个文件
		if shouldExclude(src, options.ExcludePaths) {
			return nil
		}

		file, err := os.Open(src)
		if err != nil {
			return err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.Base(src)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		_, err = io.Copy(tw, file)
		return err
	}
}

// compressTarXz 创建tar.xz压缩文件
func compressTarXz(src, dst string, isDir bool, options CompressOptions) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	xzw, err := xz.NewWriter(file)
	if err != nil {
		return err
	}
	defer xzw.Close()

	tw := tar.NewWriter(xzw)
	defer tw.Close()

	if isDir {
		// 遍历目录
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 检查是否应该排除此路径
			if shouldExclude(path, options.ExcludePaths) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// 获取相对路径
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}

			// 创建tar头部
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(tw, file)
				if err != nil {
					return err
				}
			}
			return nil
		})
	} else {
		// 压缩单个文件
		if shouldExclude(src, options.ExcludePaths) {
			return nil
		}

		file, err := os.Open(src)
		if err != nil {
			return err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.Base(src)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		_, err = io.Copy(tw, file)
		return err
	}
}

// compressGz 创建gz压缩文件
func compressGz(src, dst string) error {
	if shouldExclude(src, nil) {
		return nil
	}

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

	gzw := gzip.NewWriter(dstFile)
	defer gzw.Close()

	_, err = io.Copy(gzw, srcFile)
	return err
}

// compressBz2 创建bz2压缩文件
func compressBz2(src, dst string) error {
	if shouldExclude(src, nil) {
		return nil
	}

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

	bz2w, err := bzip2.NewWriter(dstFile, nil)
	if err != nil {
		return err
	}
	defer bz2w.Close()

	_, err = io.Copy(bz2w, srcFile)
	return err
}

// compressXz 创建xz压缩文件
func compressXz(src, dst string) error {
	if shouldExclude(src, nil) {
		return nil
	}

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

	xzw, err := xz.NewWriter(dstFile)
	if err != nil {
		return err
	}
	defer xzw.Close()

	_, err = io.Copy(xzw, srcFile)
	return err
}

// compress7z 创建7z压缩文件
func compress7z() error {
	// 目前 go7z 库不支持写入操作
	return fmt.Errorf("当前版本暂不支持创建7z文件（因为使用的库仅支持解压缩），请使用其他格式如 zip 或 tar.gz")
}

// decompressZip 解压zip文件
func decompressZip(src, dst string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// 获取目标目录的绝对路径
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		// 清理文件路径，移除开头的 / 或 ../
		cleanedPath := filepath.Clean(file.Name)
		if cleanedPath == "." || strings.HasPrefix(cleanedPath, ".."+string(os.PathSeparator)) {
			continue // 跳过可疑路径
		}

		path := filepath.Join(dst, cleanedPath)

		// 获取最终路径的绝对路径
		pathAbs, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		// 确保解压的文件路径在目标目录内
		if !strings.HasPrefix(pathAbs, dstAbs) {
			return fmt.Errorf("非法的文件路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// decompressTarGz 解压tar.gz文件
func decompressTarGz(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	return decompressTar(gzr, dst)
}

// decompressTarBz2 解压tar.bz2文件
func decompressTarBz2(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	bz2r, err := bzip2.NewReader(file, nil)
	if err != nil {
		return err
	}
	defer bz2r.Close()

	return decompressTar(bz2r, dst)
}

// decompressTarXz 解压tar.xz文件
func decompressTarXz(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	xzr, err := xz.NewReader(file)
	if err != nil {
		return err
	}

	return decompressTar(xzr, dst)
}

// decompressTar 解压tar文件
func decompressTar(reader io.Reader, dst string) error {
	tr := tar.NewReader(reader)

	// 确保目标目录存在
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// 获取目标目录的绝对路径
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 清理文件路径，移除开头的 / 或 ../
		cleanedPath := filepath.Clean(header.Name)
		if cleanedPath == "." || strings.HasPrefix(cleanedPath, ".."+string(os.PathSeparator)) {
			continue // 跳过可疑路径
		}

		path := filepath.Join(dst, cleanedPath)

		// 获取最终路径的绝对路径
		pathAbs, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		// 确保解压的文件路径在目标目录内
		if !strings.HasPrefix(pathAbs, dstAbs) {
			return fmt.Errorf("非法的文件路径: %s", header.Name)
		}

		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, tr)
		if err != nil {
			return err
		}
	}
	return nil
}

// decompressGz 解压gz文件
func decompressGz(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	gzr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gzr.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, gzr)
	return err
}

// decompressBz2 解压bz2文件
func decompressBz2(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	bz2r, err := bzip2.NewReader(srcFile, nil)
	if err != nil {
		return err
	}
	defer bz2r.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, bz2r)
	return err
}

// decompressXz 解压xz文件
func decompressXz(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	xzr, err := xz.NewReader(srcFile)
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, xzr)
	return err
}

// decompressRar 解压rar文件
func decompressRar(src, dst string) error {
	// 打开RAR文件
	rfile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer rfile.Close()

	// 创建RAR解压器
	rr, err := rardecode.NewReader(rfile, "")
	if err != nil {
		return fmt.Errorf("无法读取RAR文件: %v", err)
	}

	// 获取目标目录的绝对路径
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	// 解压每个文件
	for {
		header, err := rr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 清理文件路径，移除开头的 / 或 ../
		cleanedPath := filepath.Clean(header.Name)
		if cleanedPath == "." || strings.HasPrefix(cleanedPath, ".."+string(os.PathSeparator)) {
			continue // 跳过可疑路径
		}

		path := filepath.Join(dst, cleanedPath)

		// 获取最终路径的绝对路径
		pathAbs, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		// 确保解压的文件路径在目标目录内
		if !strings.HasPrefix(pathAbs, dstAbs) {
			return fmt.Errorf("非法的文件路径: %s", header.Name)
		}

		if header.IsDir {
			if err = os.MkdirAll(path, 0755); err != nil {
				return err
			}
			continue
		}

		// 确保父目录存在
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// 创建文件
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		// 复制文件内容
		_, err = io.Copy(file, rr)
		file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// decompress7z 解压7z文件
func decompress7z(src, dst string) error {
	// 打开源文件
	sz, err := go7z.OpenReader(src)
	if err != nil {
		return fmt.Errorf("无法读取7z文件: %v", err)
	}
	defer sz.Close()

	// 获取目标目录的绝对路径
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	// 遍历并解压所有文件
	for {
		hdr, err := sz.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 清理文件路径，移除开头的 / 或 ../
		cleanedPath := filepath.Clean(hdr.Name)
		if cleanedPath == "." || strings.HasPrefix(cleanedPath, ".."+string(os.PathSeparator)) {
			continue // 跳过可疑路径
		}

		path := filepath.Join(dst, cleanedPath)

		// 获取最终路径的绝对路径
		pathAbs, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		// 确保解压的文件路径在目标目录内
		if !strings.HasPrefix(pathAbs, dstAbs) {
			return fmt.Errorf("非法的文件路径: %s", hdr.Name)
		}

		// 如果是目录
		if strings.HasSuffix(hdr.Name, "/") {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// 创建文件
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		// 复制内容
		_, err = io.Copy(outFile, sz)
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
