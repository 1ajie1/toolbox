package network

import (
	"fmt"
	"os"
	"toolbox/pkg/netdiag"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// speedtestCmd 表示 speedtest 命令
var speedtestCmd = &cobra.Command{
	Use:   "speedtest",
	Short: "执行网络速度测试",
	Long: `执行网络速度测试，测量下载速度、上传速度和网络延迟。

该命令有两种模式：
1. 服务器模式：启动本地测试服务器
2. 测试模式：执行网络速度测试

在测试模式下，需要确保本地测试服务器（默认地址为：http://localhost:8080）已启动。

示例:
  # 启动本地测试服务器
  %[1]s network speedtest --server --port 8080

  # 执行速度测试
  %[1]s network speedtest`,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查是否以服务器模式运行
		isServer, _ := cmd.Flags().GetBool("server")
		if isServer {
			port, _ := cmd.Flags().GetInt("port")
			host, _ := cmd.Flags().GetString("host")
			dataSize, _ := cmd.Flags().GetInt("size")
			startServer(port, host, dataSize)
		} else {
			executeSpeedTest()
		}
	},
}

func init() {
	NetworkCmd.AddCommand(speedtestCmd)

	// 添加命令行标志
	speedtestCmd.Flags().BoolP("server", "s", false, "以服务器模式运行")
	speedtestCmd.Flags().IntP("port", "p", 8080, "服务器监听的端口")
	speedtestCmd.Flags().StringP("host", "H", "localhost", "服务器绑定的主机地址")
	speedtestCmd.Flags().IntP("size", "S", 10, "用于测试的数据大小(MB)")
}

// executeSpeedTest 执行网络速度测试
func executeSpeedTest() {
	fmt.Println("正在进行网络速度测试...")

	result := netdiag.RunSpeedTest()

	if result.Error != "" {
		color.Red("速度测试失败: %s\n", result.Error)
		return
	}

	color.Green("速度测试完成(服务器: %s):\n", result.ServerName)
	fmt.Printf("下载速度: %.2f Mbps\n", result.DownloadSpeed)
	fmt.Printf("上传速度: %.2f Mbps\n", result.UploadSpeed)
	fmt.Printf("延迟: %.0f ms\n", result.Latency)
}

// startServer 启动速度测试服务器
func startServer(port int, host string, dataSize int) {
	fmt.Printf("正在启动速度测试服务器 %s:%d...\n", host, port)

	config := &netdiag.SpeedTestServer{
		Port:     port,
		Host:     host,
		DataSize: dataSize,
	}

	err := netdiag.StartSpeedTestServer(config)
	if err != nil {
		fmt.Printf("启动服务器失败: %v\n", err)
		os.Exit(1)
	}
}
