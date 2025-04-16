package network

import (
	"github.com/spf13/cobra"
)

// networkCmd 表示 network 命令组
var NetworkCmd = &cobra.Command{
	Use:   "network",
	Short: "网络诊断工具集",
	Long: `网络诊断工具集，包含多种网络相关命令，如ping、端口扫描、DNS查询等。

示例:
  %[1]s network ping example.com
  %[1]s network portscan example.com --start-port 80 --end-port 100
  %[1]s network dns example.com --type mx
  %[1]s network traceroute example.com
  %[1]s network speedtest
  %[1]s network ipinfo 8.8.8.8
  %[1]s network sniff eth0 --filter "tcp and port 80"
  %[1]s network sniff --list-interfaces`,
}

func init() {
	// 子命令将在各自的init函数中添加到NetworkCmd
}
