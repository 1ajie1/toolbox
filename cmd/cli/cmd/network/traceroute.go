package network

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
  %[1]s network traceroute example.com
  %[1]s network traceroute 8.8.8.8 --max-hops 20`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		maxHops, _ := cmd.Flags().GetInt("max-hops")
		timeout, _ := cmd.Flags().GetDuration("timeout")
		packetSize, _ := cmd.Flags().GetInt("packet-size")
		noColor, _ := cmd.Flags().GetBool("no-color")

		executeTraceroute(host, maxHops, timeout, packetSize, !noColor)
	},
}

func init() {
	NetworkCmd.AddCommand(tracerouteCmd)

	// 添加命令行标志
	tracerouteCmd.Flags().IntP("max-hops", "m", 30, "最大跳数")
	tracerouteCmd.Flags().DurationP("timeout", "t", 3*time.Second, "超时时间")
	tracerouteCmd.Flags().IntP("packet-size", "s", 60, "数据包大小(字节)")
	tracerouteCmd.Flags().Bool("no-color", false, "禁用彩色输出")
}

// executeTraceroute 执行路由跟踪
func executeTraceroute(host string, maxHops int, timeout time.Duration, packetSize int, useColor bool) {
	// 如果不使用彩色输出，禁用color库的颜色功能
	color.NoColor = !useColor

	// 创建彩色输出对象
	titleColor := color.New(color.FgHiWhite, color.Bold)
	headerColor := color.New(color.FgCyan, color.Bold)
	numberColor := color.New(color.FgYellow)
	ipColor := color.New(color.FgGreen)
	hostnameColor := color.New(color.FgHiCyan)
	timeoutColor := color.New(color.FgRed)
	rttColor := color.New(color.FgMagenta)

	titleColor.Printf("正在执行到 %s 的路由跟踪 (最大跳数: %d)...\n\n", host, maxHops)

	// 打印表头
	headerColor.Println("Traceroute 路由跟踪")
	fmt.Printf("%s %s %s %s\n",
		headerColor.Sprint(fmt.Sprintf("%-5s", "跳数")),
		headerColor.Sprint(fmt.Sprintf("%-40s", "主机名")),
		headerColor.Sprint(fmt.Sprintf("%-15s", "IP地址")),
		headerColor.Sprint("延迟"))
	fmt.Println(fmt.Sprintf("%s", color.New(color.Faint).Sprint(
		"--------------------------------------------------------------------------------")))

	options := netdiag.TracerouteOptions{
		MaxHops:    maxHops,
		Timeout:    timeout,
		PacketSize: packetSize,
		RealTimeCallback: func(hop netdiag.HopInfo) {
			// 实时回调函数，当每一跳有结果时会调用此函数

			// 格式化跳数
			numStr := numberColor.Sprintf("%-5d", hop.Number)

			// 格式化主机名
			hostStr := "*"
			if hop.Name != "*" {
				hostStr = hostnameColor.Sprint(hop.Name)
			} else {
				hostStr = timeoutColor.Sprint("*")
			}
			hostStr = fmt.Sprintf("%-40s", hostStr)

			// 格式化IP地址
			ipStr := "*"
			if hop.IP != "*" {
				ipStr = ipColor.Sprint(hop.IP)
			} else {
				ipStr = timeoutColor.Sprint("*")
			}
			ipStr = fmt.Sprintf("%-15s", ipStr)

			// 格式化延迟时间
			latencyStr := "*"
			if len(hop.RTT) > 0 && hop.RTT[0] != "*" {
				latencyStr = rttColor.Sprint(hop.RTT[0])
			} else {
				latencyStr = timeoutColor.Sprint("*")
			}

			// 输出当前跳的信息
			fmt.Printf("%s %s %s %s\n", numStr, hostStr, ipStr, latencyStr)
		},
	}

	// 开始执行traceroute，这次不会收集所有结果后统一输出，而是通过回调函数实时输出
	result, err := netdiag.Traceroute(host, options)
	if err != nil {
		color.Red("错误: %v\n", err)
		os.Exit(1)
	}

	if result.Error != "" {
		color.Red("错误: %s\n", result.Error)
		os.Exit(1)
	}

	// 输出完成信息
	if len(result.Hops) > 0 {
		lastHop := result.Hops[len(result.Hops)-1]
		if lastHop.IP != "*" && lastHop.IP == result.TargetIP {
			titleColor.Printf("\n路由跟踪完成: 共经过 %d 跳到达目标 %s\n", len(result.Hops), host)
		} else {
			color.Yellow("\n路由跟踪未能到达目标，已达到最大跳数限制: %d\n", maxHops)
		}
	} else {
		color.Red("\n路由跟踪失败，未获取到任何路由信息\n")
	}
}
