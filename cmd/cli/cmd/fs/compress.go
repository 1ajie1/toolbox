package fs

import (
	"fmt"
	"os"
	"strings"
	"toolbox/pkg/fsutils"

	"github.com/spf13/cobra"
)

var compressCmd = &cobra.Command{
	Use:   "compress [源文件/目录] [目标路径]",
	Short: "压缩或解压缩文件",
	Long: `压缩或解压缩文件/目录。

模式：
  - compress:   压缩模式（默认）
  - decompress: 解压缩模式

支持的压缩格式：
  - zip:     ZIP压缩文件（支持目录）
  - tar.gz:  TAR+GZIP压缩文件（支持目录，或 .tgz）
  - tar.bz2: TAR+BZIP2压缩文件（支持目录，或 .tbz2）
  - tar.xz:  TAR+XZ压缩文件（支持目录，或 .txz）
  - gz:      GZIP压缩文件（仅支持单文件）
  - bz2:     BZIP2压缩文件（仅支持单文件）
  - xz:      XZ压缩文件（仅支持单文件）
  - 7z:      7-Zip压缩文件（支持目录）
  - rar:     RAR压缩文件（仅支持解压缩）

示例:
  # 压缩（默认模式）
  %[1]s fs compress myfile.txt myfile.txt.gz
  %[1]s fs compress mydir mydir.zip --type zip
  %[1]s fs compress mydir output.7z --type 7z
  %[1]s fs compress mydir output --type tar.gz -l 9 -k

  # 解压缩
  %[1]s fs compress myfile.txt.gz myfile.txt --mode decompress
  %[1]s fs compress mydir.zip extracted/ --mode decompress
  %[1]s fs compress mydir.7z extracted/ --mode decompress`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		dst := args[1]

		// 获取操作模式
		mode, _ := cmd.Flags().GetString("mode")
		if mode == "decompress" {
			return fsutils.Decompress(src, dst)
		}

		// 压缩模式
		// 获取压缩类型
		compressionType, _ := cmd.Flags().GetString("type")
		var format fsutils.CompressFormat

		// 如果指定了压缩类型，使用指定的类型
		if compressionType != "" {
			switch strings.ToLower(compressionType) {
			case "zip":
				format = fsutils.ZIP
			case "tar.gz", "tgz":
				format = fsutils.TARGZ
			case "tar.bz2", "tbz2":
				format = fsutils.TARBZ2
			case "tar.xz", "txz":
				format = fsutils.TARXZ
			case "gz":
				format = fsutils.GZ
			case "bz2":
				format = fsutils.BZ2
			case "xz":
				format = fsutils.XZ
			case "7z":
				format = fsutils.SEVENZIP
			default:
				return fmt.Errorf("不支持的压缩格式: %s", compressionType)
			}
		} else {
			// 否则根据目标文件扩展名自动检测
			switch {
			case strings.HasSuffix(dst, ".zip"):
				format = fsutils.ZIP
			case strings.HasSuffix(dst, ".tar.gz"), strings.HasSuffix(dst, ".tgz"):
				format = fsutils.TARGZ
			case strings.HasSuffix(dst, ".tar.bz2"), strings.HasSuffix(dst, ".tbz2"):
				format = fsutils.TARBZ2
			case strings.HasSuffix(dst, ".tar.xz"), strings.HasSuffix(dst, ".txz"):
				format = fsutils.TARXZ
			case strings.HasSuffix(dst, ".gz"):
				format = fsutils.GZ
			case strings.HasSuffix(dst, ".bz2"):
				format = fsutils.BZ2
			case strings.HasSuffix(dst, ".xz"):
				format = fsutils.XZ
			case strings.HasSuffix(dst, ".7z"):
				format = fsutils.SEVENZIP
			default:
				return fmt.Errorf("无法从文件扩展名识别压缩格式，请使用 --type 选项指定压缩格式")
			}
		}

		// 检查源路径是否为目录
		srcInfo, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("无法访问源文件/目录: %v", err)
		}

		// 检查单文件压缩格式是否用于目录
		if srcInfo.IsDir() && (format == fsutils.GZ || format == fsutils.BZ2 || format == fsutils.XZ) {
			return fmt.Errorf("%s 格式不支持压缩目录，请使用 zip、tar.gz、tar.bz2、tar.xz", format)
		}

		level, _ := cmd.Flags().GetInt("level")

		options := fsutils.CompressOptions{
			Format: format,
			Level:  level,
		}

		return fsutils.Compress(src, dst, options)
	},
}

func init() {
	compressCmd.Flags().StringP("mode", "m", "compress", "操作模式（compress 或 decompress）(解压缩额外支持rar、7z)")
	compressCmd.Flags().StringP("type", "t", "", `压缩格式（可选值：zip, tar.gz, tar.bz2, tar.xz, gz, bz2, xz）
如果不指定，将根据目标文件扩展名自动检测`)
	compressCmd.Flags().IntP("level", "l", 6, "压缩级别（1-9）")

	FsCmd.AddCommand(compressCmd)
}
