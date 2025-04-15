package cmd

import (
	"fmt"
	"time"
	"tuleaj_tools/tool-box/pkg/netdiag"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// portScanCmd 表示 portscan 命令
var portScanCmd = &cobra.Command{
	Use:   "portscan [主机名或IP]",
	Short: "执行端口扫描",
	Long: `对指定主机执行端口扫描，检测开放的端口和服务。

可以指定要扫描的端口范围或选择只扫描常见端口。

示例:
  %[1]s portscan example.com
  %[1]s portscan example.com --start-port 80 --end-port 100
  %[1]s portscan example.com --common-ports`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		startPort, _ := cmd.Flags().GetInt("start-port")
		endPort, _ := cmd.Flags().GetInt("end-port")
		commonPorts, _ := cmd.Flags().GetBool("common-ports")
		timeout, _ := cmd.Flags().GetInt("timeout")

		timeoutDuration := time.Duration(timeout) * time.Millisecond
		executePortScan(host, startPort, endPort, commonPorts, timeoutDuration)
	},
}

func init() {
	rootCmd.AddCommand(portScanCmd)

	// 添加命令行标志
	portScanCmd.Flags().IntP("start-port", "s", 1, "起始端口号")
	portScanCmd.Flags().IntP("end-port", "e", 1024, "结束端口号")
	portScanCmd.Flags().BoolP("common-ports", "c", false, "仅扫描常见端口")
	portScanCmd.Flags().IntP("timeout", "t", 1000, "连接超时(毫秒)")
}

// executePortScan 执行端口扫描
func executePortScan(host string, startPort, endPort int, commonPorts bool, timeout time.Duration) {
	fmt.Printf("正在扫描 %s 的端口...\n", host)

	var result netdiag.PortScanResult

	if commonPorts {
		fmt.Println("仅扫描常见端口...")
		result = netdiag.ScanCommonPorts(host, timeout)
	} else {
		fmt.Printf("扫描端口范围: %d-%d...\n", startPort, endPort)
		result = netdiag.ScanPorts(host, startPort, endPort, timeout)
	}

	if result.Error != "" {
		color.Red("端口扫描失败: %s\n", result.Error)
		return
	}

	if len(result.Ports) == 0 {
		color.Yellow("未发现开放的端口。\n")
		return
	}

	color.Green("发现 %d 个开放的端口:\n", len(result.Ports))
	fmt.Println("端口\t状态\t服务")
	fmt.Println("----\t----\t----")

	for _, port := range result.Ports {
		fmt.Printf("%d\t%s\t%s\n", port.Port, "开放", port.Service)
	}
}
