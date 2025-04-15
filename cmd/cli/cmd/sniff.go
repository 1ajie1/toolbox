package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"
	"tuleaj_tools/tool-box/pkg/netdiag"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// sniffCmd 表示网络抓包命令
var sniffCmd = &cobra.Command{
	Use:   "sniff [接口名]",
	Short: "执行网络抓包",
	Long: `执行网络抓包分析，类似于tcpdump功能。
该命令可以捕获指定网络接口上的数据包，并根据过滤规则进行显示。
支持保存为pcap文件格式，可与Wireshark等工具兼容。

示例:
  %[1]s sniff eth0
  %[1]s sniff eth0 --filter "tcp and port 80"
  %[1]s sniff eth0 --output capture.txt
  %[1]s sniff eth0 --pcap capture.pcap
  %[1]s sniff --list-interfaces`,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查是否要列出接口
		listInterfaces, _ := cmd.Flags().GetBool("list-interfaces")
		if listInterfaces {
			showInterfaces()
			return
		}

		// 需要指定接口名
		if len(args) < 1 {
			fmt.Println("错误: 必须指定网络接口名称")
			fmt.Println("可以使用 --list-interfaces 查看可用的网络接口")
			cmd.Help()
			os.Exit(1)
		}

		// 获取参数
		interfaceName := args[0]
		filter, _ := cmd.Flags().GetString("filter")
		output, _ := cmd.Flags().GetString("output")
		pcapFile, _ := cmd.Flags().GetString("pcap")
		count, _ := cmd.Flags().GetInt("count")
		verbose, _ := cmd.Flags().GetBool("verbose")
		promiscuous, _ := cmd.Flags().GetBool("promiscuous")
		stats, _ := cmd.Flags().GetBool("stats")
		snaplen, _ := cmd.Flags().GetInt("snaplen")
		payloadLen, _ := cmd.Flags().GetInt("payload")
		timeout, _ := cmd.Flags().GetFloat64("timeout")

		// 执行抓包
		executeSniff(interfaceName, filter, output, pcapFile, count, verbose,
			promiscuous, stats, snaplen, payloadLen, time.Duration(timeout*float64(time.Second)))
	},
}

func init() {
	rootCmd.AddCommand(sniffCmd)

	// 添加命令行标志
	sniffCmd.Flags().StringP("filter", "f", "", "设置过滤规则，如 'tcp and port 80'")
	sniffCmd.Flags().StringP("output", "o", "", "输出捕获结果到文本文件")
	sniffCmd.Flags().StringP("pcap", "w", "", "保存捕获结果为pcap文件")
	sniffCmd.Flags().IntP("count", "c", 0, "要捕获的包数量，0表示无限制")
	sniffCmd.Flags().BoolP("verbose", "v", false, "显示详细的包信息")
	sniffCmd.Flags().BoolP("promiscuous", "p", true, "启用混杂模式")
	sniffCmd.Flags().BoolP("stats", "s", true, "显示统计信息")
	sniffCmd.Flags().BoolP("list-interfaces", "l", false, "列出可用的网络接口")
	sniffCmd.Flags().IntP("snaplen", "", 1600, "捕获的数据包大小限制")
	sniffCmd.Flags().IntP("payload", "", 64, "显示的载荷长度，0表示不显示")
	sniffCmd.Flags().Float64P("timeout", "t", 0, "捕获超时时间(秒)，0表示一直捕获直到中断")
}

// showInterfaces 显示所有可用的网络接口
func showInterfaces() {
	interfaces, err := netdiag.ListInterfaces()
	if err != nil {
		fmt.Printf("获取网络接口列表失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("可用的网络接口:")
	for i, iface := range interfaces {
		fmt.Printf("%d. %s\n", i+1, iface)
	}
}

// executeSniff 执行抓包操作
func executeSniff(interfaceName, filter, output, pcapFile string, count int, verbose,
	promiscuous, stats bool, snaplen, payloadLen int, timeout time.Duration) {

	// 使用粗体黄色打印
	boldYellow := color.New(color.FgYellow, color.Bold)
	boldYellow.Printf("开始在接口 %s 上抓包...\n", interfaceName)
	if filter != "" {
		boldYellow.Printf("过滤规则: %s\n", filter)
	}
	fmt.Println("按 Ctrl+C 停止抓包")
	fmt.Println()

	// 准备配置
	config := netdiag.SnifferConfig{
		Interface:   interfaceName,
		Filter:      filter,
		Output:      output,
		Count:       count,
		Verbose:     verbose,
		Promiscuous: promiscuous,
		Statistics:  stats,
		Snaplen:     snaplen,
		PayloadLen:  payloadLen,
		SavePcap:    pcapFile,
	}

	// 设置超时
	if timeout > 0 {
		config.Timeout = timeout
	} else {
		// 无限超时
		config.Timeout = -1 * time.Second
	}

	// 执行抓包 - 现在信号处理已在内部实现
	if err := netdiag.StartSniffer(config); err != nil {
		if !strings.Contains(err.Error(), "由于系统调用而中断") {
			fmt.Printf("\n抓包失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 打印输出信息
	if output != "" {
		fmt.Printf("\n抓包结果已保存到: %s\n", output)
	}
	if pcapFile != "" {
		fmt.Printf("PCAP文件已保存到: %s\n", pcapFile)
	}
}
