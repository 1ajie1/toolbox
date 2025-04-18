package network

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"toolbox/pkg/netdiag"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// portScanCmd 表示 portscan 命令
var portScanCmd = &cobra.Command{
	Use:   "portscan [主机名或IP]",
	Short: "执行端口扫描",
	Long: `对指定主机执行端口扫描，检测开放的端口和服务。

可以指定要扫描的端口范围或选择只扫描常见端口。
也可以指定一组非连续的端口进行扫描，用逗号分隔。

示例:
  %[1]s network portscan example.com
  %[1]s network portscan example.com --start-port 80 --end-port 100
  %[1]s network portscan example.com --common-ports
  %[1]s network portscan example.com --ports 22,80,443,3306,8080`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		startPort, _ := cmd.Flags().GetInt("start-port")
		endPort, _ := cmd.Flags().GetInt("end-port")
		commonPorts, _ := cmd.Flags().GetBool("common-ports")
		portList, _ := cmd.Flags().GetString("ports")
		timeout, _ := cmd.Flags().GetInt("timeout")
		concurrency, _ := cmd.Flags().GetInt("concurrency")

		timeoutDuration := time.Duration(timeout) * time.Millisecond
		executePortScan(host, startPort, endPort, commonPorts, portList, timeoutDuration, concurrency)
	},
}

func init() {
	NetworkCmd.AddCommand(portScanCmd)

	// 添加命令行标志
	portScanCmd.Flags().IntP("start-port", "s", 1, "起始端口号")
	portScanCmd.Flags().IntP("end-port", "e", 1024, "结束端口号")
	portScanCmd.Flags().BoolP("common-ports", "c", false, "仅扫描常见端口")
	portScanCmd.Flags().StringP("ports", "p", "", "一组非连续的端口，用逗号分隔")
	portScanCmd.Flags().IntP("timeout", "t", 1000, "连接超时(毫秒)")
	portScanCmd.Flags().IntP("concurrency", "C", 100, "并发连接数")
}

// executePortScan 执行端口扫描
func executePortScan(host string, startPort, endPort int, commonPorts bool, portList string, timeout time.Duration, concurrency int) {
	fmt.Printf("正在扫描 %s 的端口...\n", host)

	var result netdiag.PortScanResult

	if portList != "" {
		// 扫描指定的端口列表
		fmt.Println("扫描指定的端口列表...")
		ports, err := parsePortList(portList)
		if err != nil {
			color.Red("解析端口列表失败: %s\n", err)
			return
		}
		result = netdiag.ScanSpecificPorts(host, ports, timeout, concurrency)
	} else if commonPorts {
		// 扫描常见端口
		fmt.Println("仅扫描常见端口...")
		result = netdiag.ScanCommonPorts(host, timeout, concurrency)
	} else {
		// 扫描端口范围
		fmt.Printf("扫描端口范围: %d-%d...\n", startPort, endPort)
		result = netdiag.ScanPorts(host, startPort, endPort, timeout, concurrency)
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

// parsePortList 解析端口列表字符串
func parsePortList(portList string) ([]int, error) {
	var ports []int

	// 分割端口列表
	portStrs := strings.Split(portList, ",")

	for _, portStr := range portStrs {
		// 去除空格
		portStr = strings.TrimSpace(portStr)
		if portStr == "" {
			continue
		}

		// 尝试将字符串转换为整数
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("端口 '%s' 格式无效: %v", portStr, err)
		}

		// 检查端口范围
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("端口 %d 超出有效范围 (1-65535)", port)
		}

		ports = append(ports, port)
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("未提供有效的端口")
	}

	return ports, nil
}
