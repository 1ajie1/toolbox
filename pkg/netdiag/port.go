package netdiag

import (
	"fmt"
	"net"
	"time"
)

// PortStatus 表示端口状态
type PortStatus struct {
	Port    int
	Open    bool
	Service string
}

// PortScanResult 表示端口扫描结果
type PortScanResult struct {
	Host  string
	Ports []PortStatus
	Error string
}

// 常见端口及其服务
var commonPorts = map[int]string{
	21:    "FTP",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	110:   "POP3",
	143:   "IMAP",
	443:   "HTTPS",
	3306:  "MySQL",
	5432:  "PostgreSQL",
	6379:  "Redis",
	8080:  "HTTP Proxy",
	9000:  "Prometheus",
	9090:  "Prometheus",
	27017: "MongoDB",
}

// ScanPort 检测指定主机的单个端口是否开放
func ScanPort(host string, port int, timeout time.Duration) PortStatus {
	result := PortStatus{
		Port: port,
		Open: false,
	}

	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)

	if err != nil {
		return result
	}

	defer conn.Close()
	result.Open = true

	// 尝试识别服务
	if service, ok := commonPorts[port]; ok {
		result.Service = service
	} else {
		result.Service = "未知服务"
	}

	return result
}

// ScanPorts 扫描主机的多个端口
func ScanPorts(host string, startPort, endPort int, timeout time.Duration) PortScanResult {
	result := PortScanResult{
		Host:  host,
		Ports: []PortStatus{},
	}

	// 检查主机名是否有效
	_, err := net.LookupHost(host)
	if err != nil {
		result.Error = fmt.Sprintf("无法解析主机名: %v", err)
		return result
	}

	for port := startPort; port <= endPort; port++ {
		status := ScanPort(host, port, timeout)
		if status.Open {
			result.Ports = append(result.Ports, status)
		}
	}

	return result
}

// ScanCommonPorts 扫描主机的常用端口
func ScanCommonPorts(host string, timeout time.Duration) PortScanResult {
	result := PortScanResult{
		Host:  host,
		Ports: []PortStatus{},
	}

	// 检查主机名是否有效
	_, err := net.LookupHost(host)
	if err != nil {
		result.Error = fmt.Sprintf("无法解析主机名: %v", err)
		return result
	}

	for port := range commonPorts {
		status := ScanPort(host, port, timeout)
		if status.Open {
			result.Ports = append(result.Ports, status)
		}
	}

	return result
}
