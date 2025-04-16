package network

import (
	"fmt"
	"strings"
	"tuleaj_tools/tool-box/pkg/netdiag"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// dnsCmd 表示 dns 命令
var dnsCmd = &cobra.Command{
	Use:   "dns [域名]",
	Short: "执行DNS查询",
	Long: `查询指定域名的DNS记录。

可以指定要查询的DNS记录类型，如A/AAAA(IP)、MX、NS、TXT等。
默认查询A和AAAA记录（IP地址）。

可以指定使用哪个DNS服务器进行查询，格式为IP:端口，如8.8.8.8:53。
如果不指定DNS服务器，则使用系统默认的DNS解析方式。

示例:
  %[1]s network dns example.com
  %[1]s network dns example.com --type mx
  %[1]s network dns example.com --type ns
  %[1]s network dns example.com --dns-server 8.8.8.8
  %[1]s network dns example.com --dns-server 8.8.8.8:53 --type all`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		recordType, _ := cmd.Flags().GetString("type")
		dnsServer, _ := cmd.Flags().GetString("dns-server")

		executeDNSQuery(domain, recordType, dnsServer)
	},
}

func init() {
	NetworkCmd.AddCommand(dnsCmd)

	// 添加命令行标志
	dnsCmd.Flags().StringP("type", "t", "ip", "DNS记录类型 (ip, mx, ns, txt, all)")
	dnsCmd.Flags().StringP("dns-server", "d", "", "指定DNS服务器 (例如: 8.8.8.8 或 8.8.8.8:53)")
}

// executeDNSQuery 执行DNS查询
func executeDNSQuery(domain string, recordType string, dnsServer string) {
	fmt.Printf("正在查询 %s 的DNS记录...\n", domain)
	if dnsServer != "" {
		fmt.Printf("使用DNS服务器: %s\n", dnsServer)
	}

	recordType = strings.ToLower(recordType)

	if recordType == "all" {
		// 查询所有类型的记录
		results := netdiag.QueryDNS(domain, dnsServer)

		for recordType, result := range results {
			if result.Error != "" {
				color.Red("%s记录查询失败: %s\n", recordType, result.Error)
				continue
			}

			if len(result.Records) == 0 {
				color.Yellow("未找到%s记录。\n", recordType)
				continue
			}

			color.Green("%s记录 (查询方式: %s):\n", recordType, getQueryMethodText(result))
			for _, record := range result.Records {
				fmt.Printf("类型: %s, 值: %s\n", record.Type, record.Value)
			}
			fmt.Println()
		}
	} else {
		// 查询指定类型的记录
		var result netdiag.DNSQueryResult
		var err error

		switch recordType {
		case "ip":
			result, err = netdiag.LookupIP(domain, dnsServer)
		case "mx":
			result, err = netdiag.LookupMX(domain, dnsServer)
		case "ns":
			result, err = netdiag.LookupNS(domain, dnsServer)
		case "txt":
			result, err = netdiag.LookupTXT(domain, dnsServer)
		default:
			fmt.Printf("不支持的DNS记录类型: %s\n", recordType)
			return
		}

		if err != nil {
			color.Red("DNS查询失败: %s\n", err)
			return
		}

		if len(result.Records) == 0 {
			color.Yellow("未找到%s记录。\n", recordType)
			return
		}

		color.Green("%s记录 (查询方式: %s):\n", strings.ToUpper(recordType), getQueryMethodText(result))
		for _, record := range result.Records {
			fmt.Printf("类型: %s, 值: %s\n", record.Type, record.Value)
		}
	}
}

// getQueryMethodText 获取查询方式的文本描述
func getQueryMethodText(result netdiag.DNSQueryResult) string {
	if result.Method == "host" {
		if strings.Contains(result.ServerUsed, "系统DNS") {
			return "系统DNS"
		}
		return fmt.Sprintf("系统DNS服务器 (%s)", result.ServerUsed)
	} else if result.Method == "dns" {
		return fmt.Sprintf("自定义DNS服务器 (%s)", result.ServerUsed)
	}
	return "未知"
}
