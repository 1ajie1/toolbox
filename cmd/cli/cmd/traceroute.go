package cmd

import (
	"fmt"
	"os"
	"time"
	"tuleaj_tools/tool-box/pkg/netdiag"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// tracerouteCmd 表示 traceroute 命令
var tracerouteCmd = &cobra.Command{
	Use:   "traceroute [主机名或IP]",
	Short: "执行路由跟踪",
	Long: `执行路由跟踪，显示数据包从本地到目标主机的路径。

该命令会显示数据包经过的每个路由节点，包括IP地址、主机名和延迟。

示例:
  %[1]s traceroute example.com
  %[1]s traceroute 8.8.8.8 --max-hops 20`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		maxHops, _ := cmd.Flags().GetInt("max-hops")
		timeout, _ := cmd.Flags().GetDuration("timeout")
		packetSize, _ := cmd.Flags().GetInt("packet-size")

		executeTraceroute(host, maxHops, timeout, packetSize)
	},
}

func init() {
	rootCmd.AddCommand(tracerouteCmd)

	// 添加命令行标志
	tracerouteCmd.Flags().IntP("max-hops", "m", 30, "最大跳数")
	tracerouteCmd.Flags().DurationP("timeout", "t", 3*time.Second, "超时时间")
	tracerouteCmd.Flags().IntP("packet-size", "s", 60, "数据包大小(字节)")
}

// executeTraceroute 执行路由跟踪
func executeTraceroute(host string, maxHops int, timeout time.Duration, packetSize int) {
	fmt.Printf("正在执行到 %s 的路由跟踪(最大跳数: %d)...\n", host, maxHops)

	options := netdiag.TracerouteOptions{
		MaxHops:    maxHops,
		Timeout:    timeout,
		PacketSize: packetSize,
	}

	result, err := netdiag.Traceroute(host, options)
	if err != nil {
		color.Red("错误: %v\n", err)
		os.Exit(1)
	}

	if result.Error != "" {
		color.Red("错误: %s\n", result.Error)
		os.Exit(1)
	}

	// 显示结果
	fmt.Println("\n路由跟踪结果:")
	fmt.Printf("%-5s %-40s %-15s %s\n", "跳数", "主机名", "IP地址", "延迟")
	fmt.Println("----------------------------------------------------------------")

	for _, hop := range result.Hops {
		// 获取第一个RTT值作为延迟
		latency := "*"
		if len(hop.RTT) > 0 && hop.RTT[0] != "*" {
			latency = hop.RTT[0]
		}

		if hop.IP == "*" {
			fmt.Printf("%-5d %-40s %-15s %s\n", hop.Number, "*", "*", "*")
		} else {
			fmt.Printf("%-5d %-40s %-15s %s\n", hop.Number, hop.Name, hop.IP, latency)
		}
	}
}
