package version

import (
	"fmt"

	"toolbox/pkg/version"

	"github.com/spf13/cobra"
)

// VersionCmd 表示 version 命令
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示工具箱的版本信息，包括版本号、构建信息和开发者信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.GetVersionInfo())
	},
}
