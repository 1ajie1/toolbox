package version

import (
	"fmt"
)

// 硬编码的版本信息
const (
	version   = "1.1.0"
	developer = "tuleaj"
	about     = "Toolbox 是一款用go语言编写的多功能命令行工具，提供多种实用工具以完成各种任务。"
)

// GetVersionInfo 返回完整的版本信息
func GetVersionInfo() string {
	return fmt.Sprintf(`Toolbox Version Information:
Version:    %s
Developer:  %s
About:      %s
`, version, developer, about)
}
