package fs

import (
	"fmt"
	"path/filepath"
	"strings"
	"toolbox/pkg/fsutils"

	"github.com/spf13/cobra"
)

var splitCmd = &cobra.Command{
	Use:   "split [目录]",
	Short: "将目录打包并分片",
	Long: `将目录打包并分片，支持以下功能：
1. 将目录内容打包（支持多种压缩格式）
2. 将打包后的文件分割成指定大小的分片
3. 支持多线程并发处理
4. 支持合并分片还原文件

示例:
  # 使用默认设置分片（100M，zip格式）
  %[1]s fs split ./mydir

  # 指定分片大小和压缩格式
  %[1]s fs split ./mydir --size 1G --format tar.gz

  # 指定输出目录和线程数
  %[1]s fs split ./mydir --output ./chunks --threads 4

  # 合并分片
  %[1]s fs split ./mydir_chunks --merge mydir.zip`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		merge, _ := cmd.Flags().GetBool("merge")

		if merge {
			// 合并模式
			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				// 如果未指定输出文件，使用目录名推测
				dir := filepath.Clean(path)
				base := filepath.Base(dir)
				if strings.HasSuffix(base, "_chunks") {
					base = strings.TrimSuffix(base, "_chunks")
				}
				// 尝试确定文件扩展名
				files, err := filepath.Glob(filepath.Join(dir, "chunk_001*"))
				if err == nil && len(files) > 0 {
					ext := filepath.Ext(files[0])
					if ext != "" {
						base += ext
					}
				}
				output = base
			}

			if err := fsutils.MergeChunks(path, output, false); err != nil {
				return fmt.Errorf("合并分片失败: %v", err)
			}
			fmt.Printf("分片已合并到：%s\n", output)
			return nil
		}

		// 分片模式
		size, _ := cmd.Flags().GetString("size")
		format, _ := cmd.Flags().GetString("format")
		threads, _ := cmd.Flags().GetInt("threads")
		output, _ := cmd.Flags().GetString("output")
		remove, _ := cmd.Flags().GetBool("remove")

		// 解析分片大小
		var chunkSize int64 = 100 * 1024 * 1024 // 默认100M
		if size != "" {
			var multiplier int64 = 1024 * 1024 // 默认单位为M
			size = strings.ToUpper(size)
			if strings.HasSuffix(size, "G") {
				multiplier = 1024 * 1024 * 1024
				size = strings.TrimSuffix(size, "G")
			} else if strings.HasSuffix(size, "M") {
				size = strings.TrimSuffix(size, "M")
			}
			fmt.Sscanf(size, "%d", &chunkSize)
			chunkSize *= multiplier
		}

		// 解析压缩格式
		compressType := fsutils.ZIP // 默认使用ZIP
		switch strings.ToLower(format) {
		case "zip":
			compressType = fsutils.ZIP
		case "tar.gz", "tgz":
			compressType = fsutils.TARGZ
		case "tar.bz2", "tbz2":
			compressType = fsutils.TARBZ2
		case "tar.xz", "txz":
			compressType = fsutils.TARXZ
		default:
			return fmt.Errorf("不支持的压缩格式：%s（支持的格式：zip, tar.gz/tgz, tar.bz2/tbz2, tar.xz/txz）", format)
		}

		// 准备选项
		opts := fsutils.SplitOptions{
			SourceDir:    path,
			OutputDir:    output,
			ChunkSize:    chunkSize,
			CompressType: compressType,
			ThreadCount:  threads,
			DeleteSource: remove,
		}

		// 执行分片
		if err := fsutils.SplitArchive(&opts); err != nil {
			return fmt.Errorf("分片失败: %v", err)
		}

		fmt.Printf("分片完成，输出目录：%s\n", opts.OutputDir)
		return nil
	},
}

func init() {
	splitCmd.Flags().StringP("size", "s", "100M", "分片大小（例如：100M, 1G）")
	splitCmd.Flags().StringP("format", "f", "zip", "压缩格式（zip, tar.gz/tgz, tar.bz2/tbz2, tar.xz/txz）")
	splitCmd.Flags().StringP("output", "o", "", "输出目录（默认为源目录名_chunks）")
	splitCmd.Flags().IntP("threads", "t", 0, "线程数（默认为CPU核心数）")
	splitCmd.Flags().BoolP("remove", "r", false, "完成后删除源目录")
	splitCmd.Flags().Bool("merge", false, "合并模式（将指定目录中的分片合并）")

	FsCmd.AddCommand(splitCmd)
}
