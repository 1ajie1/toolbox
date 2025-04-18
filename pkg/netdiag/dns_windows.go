//go:build windows
// +build windows

package netdiag

import (
	"fmt"

	"github.com/StackExchange/wmi"
)

// getWindowsDNSServers 使用Windows API获取DNS服务器
func getWindowsDNSServers() ([]string, error) {
	var dnsServers []string

	// 使用WMI查询获取网络适配器配置信息
	type NetworkAdapterConfiguration struct {
		DNSServerSearchOrder []string
	}

	var nacs []NetworkAdapterConfiguration

	// 使用go-wmi包查询网络适配器配置
	query := "SELECT DNSServerSearchOrder FROM Win32_NetworkAdapterConfiguration WHERE IPEnabled = TRUE"

	err := wmi.Query(query, &nacs)
	if err != nil {
		return nil, fmt.Errorf("WMI查询失败: %v", err)
	}

	// 从查询结果中提取DNS服务器信息
	for _, config := range nacs {
		if len(config.DNSServerSearchOrder) > 0 {
			for _, dns := range config.DNSServerSearchOrder {
				if dns != "" && !contains(dnsServers, dns) {
					dnsServers = append(dnsServers, dns)
				}
			}
		}
	}

	if len(dnsServers) == 0 {
		return nil, fmt.Errorf("未找到DNS服务器信息")
	}

	return dnsServers, nil
}
