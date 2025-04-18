package network

import (
	"fmt"
	"os"
	"time"
	"toolbox/pkg/netdiag"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// pingCmd 表示ping命令
var pingCmd = &cobra.Command{
	Use:   "ping [主机名或IP]",
	Short: "执行Ping测试",
	Long: `执行Ping测试来检查网络连通性和测量延迟。
该命令将向指定的主机发送ICMP echo请求包，并显示结果。

示例:
  %[1]s network ping example.com
  %[1]s network ping 8.8.8.8 --count 10
  %[1]s network ping example.com --interval 2`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		count, _ := cmd.Flags().GetInt("count")
		interval, _ := cmd.Flags().GetFloat64("interval")

		executePing(host, count, time.Duration(interval*float64(time.Second)))
	},
}

func init() {
	NetworkCmd.AddCommand(pingCmd)

	// 添加命令行标志
	pingCmd.Flags().IntP("count", "c", 4, "要发送的Ping包数量")
	pingCmd.Flags().Float64P("interval", "i", 1.0, "Ping的间隔时间(秒)")
}

// executePing 执行Ping命令
func executePing(host string, count int, interval time.Duration) {
	fmt.Printf("正在Ping %s (%d次，间隔%.1f秒)...\n\n", host, count, interval.Seconds())

	// 创建颜色对象
	successColor := color.New(color.FgGreen)
	errorColor := color.New(color.FgRed)

	// 创建选项
	options := netdiag.PingOptions{
		Count:    count,
		Interval: interval,
	}

	// 回调函数，用于实时显示ping结果
	pingCallback := func(line string) {
		if line != "" {
			fmt.Println(line)
		}
	}

	// 执行ping操作
	result, err := netdiag.Ping(host, options, pingCallback)
	if err != nil {
		fmt.Println("\n错误:", err)
		os.Exit(1)
	}

	if !result.Success {
		errorColor.Printf("\nPing %s 失败: %s\n", host, result.Error)
		os.Exit(1)
	}

	// 显示统计信息
	fmt.Println("\n---- Ping 统计信息 ----")
	successColor.Printf("Ping %s 成功:\n", host)
}
