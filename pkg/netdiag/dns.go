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
	// 常见的resolv.conf文件位置
	switch runtime.GOOS {
	case "windows":
		// 在Windows系统下使用ipconfig命令获取DNS服务器
		cmd := exec.Command("ipconfig", "/all")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for i, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "DNS Servers") || strings.Contains(line, "DNS 服务器") {
					if i+1 < len(lines) {
						dnsLine := strings.TrimSpace(lines[i+1])
						if dnsLine != "" && !strings.Contains(dnsLine, ":") {
							dnsServers = append(dnsServers, dnsLine)
						}
					}
				}
			}
		}
	default:
		// Linux/Unix/macOS系统
		// 常见的resolv.conf文件位置
		resolvConfPaths := []string{
			"/etc/resolv.conf",         // Linux/Unix
			"/private/etc/resolv.conf", // macOS
		}

		// 尝试从resolv.conf文件获取DNS服务器
		for _, path := range resolvConfPaths {
			content, err := os.ReadFile(path)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "nameserver") {
						parts := strings.Fields(line)
						if len(parts) >= 2 {
							dnsServers = append(dnsServers, parts[1])
						}
					}
				}
				// 如果找到了DNS服务器，就跳出循环
				if len(dnsServers) > 0 {
					break
				}
			}
		}
	}

	// 如果上面的方法没有找到DNS服务器，则使用一些常见的默认DNS
	if len(dnsServers) == 0 {
		// 尝试使用常见的DNS服务器进行简单的连通性测试
		commonDNS := []string{"8.8.8.8", "8.8.4.4", "1.1.1.1", "114.114.114.114"}
		for _, dns := range commonDNS {
			d := net.Dialer{Timeout: 500 * time.Millisecond}
			conn, err := d.Dial("udp", dns+":53")
			if err == nil {
				conn.Close()
				dnsServers = append(dnsServers, dns)
				break
			}
		}
	}

	// 如果还是没有找到可用的DNS服务器，添加一个回退选项
	if len(dnsServers) == 0 {
		dnsServers = append(dnsServers, "未知DNS服务器")
	}

	return dnsServers
}

// 创建自定义DNS解析器
func createResolver(dnsServer string) *net.Resolver {
	if dnsServer == "" {
		return net.DefaultResolver
	}

	// 确保DNS服务器包含端口
	if _, _, err := net.SplitHostPort(dnsServer); err != nil {
		dnsServer = dnsServer + ":53"
	}

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 10,
			}
			return d.DialContext(ctx, "udp", dnsServer)
		},
	}
}

// LookupIP 查询域名对应的IP地址
func LookupIP(domain string, dnsServer string) (DNSQueryResult, error) {
	result := DNSQueryResult{
		Domain:  domain,
		Records: []DNSRecord{},
	}

	var ips []net.IP
	var err error

	if dnsServer == "" {
		// 使用系统默认DNS，获取系统DNS服务器信息
		systemDNS := GetSystemDNSServers()
		result.Method = "host"
		if len(systemDNS) > 0 {
			result.ServerUsed = strings.Join(systemDNS, ", ")
		} else {
			result.ServerUsed = "系统DNS"
		}
		ips, err = net.LookupIP(domain)
	} else {
		// 使用指定的DNS服务器
		result.Method = "dns"
		result.ServerUsed = dnsServer
		resolver := createResolver(dnsServer)
		ips, err = resolver.LookupIP(context.Background(), "ip", domain)
	}

	if err != nil {
		result.Error = fmt.Sprintf("IP地址查询失败: %v", err)
		return result, err
	}

	for _, ip := range ips {
		ipType := "IPv4"
		if ip.To4() == nil {
			ipType = "IPv6"
		}
		result.Records = append(result.Records, DNSRecord{
			Type:  ipType,
			Value: ip.String(),
		})
	}

	return result, nil
}

// LookupMX 查询域名的MX记录
func LookupMX(domain string, dnsServer string) (DNSQueryResult, error) {
	result := DNSQueryResult{
		Domain:  domain,
		Records: []DNSRecord{},
	}

	var mxs []*net.MX
	var err error

	if dnsServer == "" {
		// 使用系统默认DNS，获取系统DNS服务器信息
		systemDNS := GetSystemDNSServers()
		result.Method = "host"
		if len(systemDNS) > 0 {
			result.ServerUsed = strings.Join(systemDNS, ", ")
		} else {
			result.ServerUsed = "系统DNS"
		}
		mxs, err = net.LookupMX(domain)
	} else {
		// 使用指定的DNS服务器
		result.Method = "dns"
		result.ServerUsed = dnsServer
		resolver := createResolver(dnsServer)
		mxs, err = resolver.LookupMX(context.Background(), domain)
	}

	if err != nil {
		result.Error = fmt.Sprintf("MX记录查询失败: %v", err)
		return result, err
	}

	for _, mx := range mxs {
		result.Records = append(result.Records, DNSRecord{
			Type:  "MX",
			Value: fmt.Sprintf("%s (优先级: %d)", mx.Host, mx.Pref),
		})
	}

	return result, nil
}

// LookupNS 查询域名的NS记录
func LookupNS(domain string, dnsServer string) (DNSQueryResult, error) {
	result := DNSQueryResult{
		Domain:  domain,
		Records: []DNSRecord{},
	}

	var nss []*net.NS
	var err error

	if dnsServer == "" {
		// 使用系统默认DNS，获取系统DNS服务器信息
		systemDNS := GetSystemDNSServers()
		result.Method = "host"
		if len(systemDNS) > 0 {
			result.ServerUsed = strings.Join(systemDNS, ", ")
		} else {
			result.ServerUsed = "系统DNS"
		}
		nss, err = net.LookupNS(domain)
	} else {
		// 使用指定的DNS服务器
		result.Method = "dns"
		result.ServerUsed = dnsServer
		resolver := createResolver(dnsServer)
		nss, err = resolver.LookupNS(context.Background(), domain)
	}

	if err != nil {
		result.Error = fmt.Sprintf("NS记录查询失败: %v", err)
		return result, err
	}

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
		Domain:  domain,
		Records: []DNSRecord{},
	}

	var txts []string
	var err error

	if dnsServer == "" {
		// 使用系统默认DNS，获取系统DNS服务器信息
		systemDNS := GetSystemDNSServers()
		result.Method = "host"
		if len(systemDNS) > 0 {
			result.ServerUsed = strings.Join(systemDNS, ", ")
		} else {
			result.ServerUsed = "系统DNS"
		}
		txts, err = net.LookupTXT(domain)
	} else {
		// 使用指定的DNS服务器
		result.Method = "dns"
		result.ServerUsed = dnsServer
		resolver := createResolver(dnsServer)
		txts, err = resolver.LookupTXT(context.Background(), domain)
	}

	if err != nil {
		result.Error = fmt.Sprintf("TXT记录查询失败: %v", err)
		return result, err
	}

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
