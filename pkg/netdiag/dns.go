package netdiag

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// DNSRecord 表示DNS记录
type DNSRecord struct {
	Type  string
	Value string
}

// DNSQueryResult 表示DNS查询结果
type DNSQueryResult struct {
	Domain     string
	Records    []DNSRecord
	Error      string
	Method     string // 查询方式: "host" 或 "dns"
	ServerUsed string // 如果使用DNS服务器，记录使用的服务器
}

// GetSystemDNSServers 获取系统当前使用的DNS服务器
func GetSystemDNSServers() []string {
	var dnsServers []string

	// 根据操作系统选择不同的实现
	switch runtime.GOOS {
	case "windows":
		// 使用Windows特定的函数
		servers, err := getWindowsDNSServers()
		if err == nil && len(servers) > 0 {
			return servers
		}

		// 如果API方法失败，回退到使用ipconfig命令
		cmd := exec.Command("ipconfig", "/all")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "DNS Servers") || strings.Contains(line, "DNS 服务器") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						servers := strings.TrimSpace(parts[1])
						if servers != "" {
							dnsServers = append(dnsServers, servers)
						}
					}
				}
			}
		}
	case "linux", "darwin":
		// Linux/macOS: 解析resolv.conf文件
		resolvConfPaths := []string{"/etc/resolv.conf"}
		if runtime.GOOS == "darwin" {
			// macOS有时会使用不同的resolv.conf位置
			resolvConfPaths = append(resolvConfPaths, "/private/etc/resolv.conf")
		}

		// 尝试所有可能的文件位置
		for _, path := range resolvConfPaths {
			if _, err := os.Stat(path); err == nil {
				// 文件存在
				data, err := os.ReadFile(path)
				if err == nil {
					lines := strings.Split(string(data), "\n")
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if strings.HasPrefix(line, "nameserver") {
							parts := strings.Fields(line)
							if len(parts) >= 2 {
								dnsServers = append(dnsServers, parts[1])
							}
						}
					}
				}
				// 如果找到了DNS服务器，停止搜索
				if len(dnsServers) > 0 {
					break
				}
			}
		}

		// 如果没有找到DNS服务器，尝试使用systemd-resolve命令
		if len(dnsServers) == 0 {
			if runtime.GOOS == "linux" {
				cmd := exec.Command("systemd-resolve", "--status")
				output, err := cmd.Output()
				if err == nil {
					lines := strings.Split(string(output), "\n")
					inDNSSection := false
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if strings.Contains(line, "DNS Servers:") {
							inDNSSection = true
							continue
						}
						if inDNSSection && strings.HasPrefix(line, "        ") {
							// DNS服务器行通常会有缩进
							dnsServers = append(dnsServers, strings.TrimSpace(line))
						} else if inDNSSection && line == "" {
							inDNSSection = false
						}
					}
				}
			}
		}
	}

	// 如果所有方法都失败，使用公共DNS服务器作为回退
	if len(dnsServers) == 0 {
		// 使用Google和Cloudflare DNS作为备选
		dnsServers = []string{"8.8.8.8", "1.1.1.1"}
	}

	return dnsServers
}

// 创建自定义解析器
func createResolver(dnsServer string) *net.Resolver {
	if dnsServer == "" {
		return net.DefaultResolver
	}

	// 检查dnsServer是否包含端口号
	if !strings.Contains(dnsServer, ":") {
		dnsServer = dnsServer + ":53"
	}

	// 创建一个自定义解析器，指向特定的DNS服务器
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := net.Dialer{
				Timeout: time.Second * 10,
			}
			return dialer.DialContext(ctx, "udp", dnsServer)
		},
	}
	return r
}

// LookupIP 查询域名的A和AAAA记录
func LookupIP(domain string, dnsServer string) (DNSQueryResult, error) {
	result := DNSQueryResult{
		Domain: domain,
	}

	// 创建解析器
	resolver := createResolver(dnsServer)
	if dnsServer != "" {
		result.ServerUsed = dnsServer
		result.Method = "dns"
	} else {
		result.Method = "host"
	}

	// 查询IP地址
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ips, err := resolver.LookupIP(ctx, "ip", domain)
	if err != nil {
		result.Error = fmt.Sprintf("查询失败: %v", err)
		return result, err
	}

	// 将IP地址添加到结果中
	for _, ip := range ips {
		recordType := "A"
		if ip.To4() == nil {
			recordType = "AAAA"
		}
		result.Records = append(result.Records, DNSRecord{
			Type:  recordType,
			Value: ip.String(),
		})
	}

	return result, nil
}

// LookupMX 查询域名的MX记录
func LookupMX(domain string, dnsServer string) (DNSQueryResult, error) {
	result := DNSQueryResult{
		Domain: domain,
	}

	// 创建解析器
	resolver := createResolver(dnsServer)
	if dnsServer != "" {
		result.ServerUsed = dnsServer
		result.Method = "dns"
	} else {
		result.Method = "host"
	}

	// 查询MX记录
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mxs, err := resolver.LookupMX(ctx, domain)
	if err != nil {
		result.Error = fmt.Sprintf("查询失败: %v", err)
		return result, err
	}

	// 将MX记录添加到结果中
	for _, mx := range mxs {
		result.Records = append(result.Records, DNSRecord{
			Type:  "MX",
			Value: fmt.Sprintf("%d %s", mx.Pref, mx.Host),
		})
	}

	return result, nil
}

// LookupNS 查询域名的NS记录
func LookupNS(domain string, dnsServer string) (DNSQueryResult, error) {
	result := DNSQueryResult{
		Domain: domain,
	}

	// 创建解析器
	resolver := createResolver(dnsServer)
	if dnsServer != "" {
		result.ServerUsed = dnsServer
		result.Method = "dns"
	} else {
		result.Method = "host"
	}

	// 查询NS记录
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nss, err := resolver.LookupNS(ctx, domain)
	if err != nil {
		result.Error = fmt.Sprintf("查询失败: %v", err)
		return result, err
	}

	// 将NS记录添加到结果中
	for _, ns := range nss {
		result.Records = append(result.Records, DNSRecord{
			Type:  "NS",
			Value: ns.Host,
		})
	}

	return result, nil
}

// LookupTXT 查询域名的TXT记录
func LookupTXT(domain string, dnsServer string) (DNSQueryResult, error) {
	result := DNSQueryResult{
		Domain: domain,
	}

	// 创建解析器
	resolver := createResolver(dnsServer)
	if dnsServer != "" {
		result.ServerUsed = dnsServer
		result.Method = "dns"
	} else {
		result.Method = "host"
	}

	// 查询TXT记录
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txts, err := resolver.LookupTXT(ctx, domain)
	if err != nil {
		result.Error = fmt.Sprintf("查询失败: %v", err)
		return result, err
	}

	// 将TXT记录添加到结果中
	for _, txt := range txts {
		result.Records = append(result.Records, DNSRecord{
			Type:  "TXT",
			Value: txt,
		})
	}

	return result, nil
}

// QueryDNS 查询域名的所有DNS记录
func QueryDNS(domain string, dnsServer string) map[string]DNSQueryResult {
	results := make(map[string]DNSQueryResult)

	// 查询A和AAAA记录
	ipResult, _ := LookupIP(domain, dnsServer)
	results["IP"] = ipResult

	// 查询MX记录
	mxResult, _ := LookupMX(domain, dnsServer)
	results["MX"] = mxResult

	// 查询NS记录
	nsResult, _ := LookupNS(domain, dnsServer)
	results["NS"] = nsResult

	// 查询TXT记录
	txtResult, _ := LookupTXT(domain, dnsServer)
	results["TXT"] = txtResult

	return results
}

// contains 检查字符串slice是否包含特定值
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
