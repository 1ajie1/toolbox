package network

import (
	"fmt"
	"tuleaj_tools/tool-box/pkg/netdiag"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ipinfoCmd 表示 ipinfo 命令
var ipinfoCmd = &cobra.Command{
	Use:   "ipinfo [IP地址]",
	Short: "获取IP地址信息",
	Long: `获取IP地址的地理位置和相关信息。

如果不提供IP地址，则获取本机的公网IP信息。
该命令使用ipinfo.io的API服务来获取IP信息。

示例:
  %[1]s network ipinfo
  %[1]s network ipinfo 8.8.8.8`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var ip string
		if len(args) > 0 {
			ip = args[0]
		}
		executeIPInfo(ip)
	},
}

func init() {
	NetworkCmd.AddCommand(ipinfoCmd)
}

// executeIPInfo 获取IP信息
func executeIPInfo(ip string) {
	if ip == "" {
		fmt.Println("正在获取本机公网IP信息...")
	} else {
		fmt.Printf("正在获取IP %s 的信息...\n", ip)
	}

	info, err := netdiag.GetIPInfo(ip)
	if err != nil {
		color.Red("获取IP信息失败: %s\n", err)
		return
	}

	color.Green("IP信息:\n")
	fmt.Printf("IP地址: %s\n", info.IP)
	fmt.Printf("位置: %s, %s, %s\n", info.City, info.Region, info.Country)
	fmt.Printf("邮政编码: %s\n", info.PostalCode)
	fmt.Printf("ISP: %s\n", info.ISP)
	fmt.Printf("时区: %s\n", info.Timezone)

	// 如果是本机IP，还显示本地网络接口信息
	if ip == "" {
		fmt.Println("\n本地网络接口信息:")
		localIPs, err := netdiag.GetLocalIPs()
		if err != nil {
			color.Red("获取本地网络接口信息失败: %s\n", err)
			return
		}

		for i, localIP := range localIPs {
			ipVersion := "IPv4"
			if !localIP.IsIPv4 {
				ipVersion = "IPv6"
			}

			fmt.Printf("[%d] 接口: %s\n", i+1, localIP.InterfaceName)
			fmt.Printf("    IP地址: %s (%s)\n", localIP.IPAddress, ipVersion)
			fmt.Printf("    MAC地址: %s\n", localIP.MACAddress)
			fmt.Printf("    状态: %s\n\n", statusText(localIP.IsUp))
		}
	}
}

// statusText 状态文本
func statusText(isUp bool) string {
	if isUp {
		return "已连接"
	}
	return "已断开"
}
