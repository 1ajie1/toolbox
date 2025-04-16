package netdiag

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
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
	log.Printf("开始扫描主机 %s 的端口 %d", host, port)
	result := PortStatus{
		Port: port,
		Open: false,
	}

	// 正确处理IPv6地址格式
	var address string
	if strings.Contains(host, ":") && !strings.Contains(host, "[") {
		// IPv6地址需要用方括号括起来
		address = fmt.Sprintf("[%s]:%d", host, port)
	} else {
		address = fmt.Sprintf("%s:%d", host, port)
	}

	conn, err := net.DialTimeout("tcp", address, timeout)

	if err != nil {
		log.Printf("扫描主机 %s 的端口 %d 失败: %v", host, port, err)
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
func ScanPorts(host string, startPort, endPort int, timeout time.Duration, concurrency int) PortScanResult {
	result := PortScanResult{
		Host:  host,
		Ports: []PortStatus{},
	}

	// 检查主机名是否有效
	_, err := net.LookupHost(host)
	if err != nil {
		log.Printf("无法解析主机名 %s: %v", host, err)
		result.Error = fmt.Sprintf("无法解析主机名: %v", err)
		return result
	}

	var wg sync.WaitGroup
	results := make(chan PortStatus, endPort-startPort+1)
	sem := make(chan struct{}, concurrency)

	for port := startPort; port <= endPort; port++ {
		wg.Add(1)
		go func(p int) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()
			results <- ScanPort(host, p, timeout)
		}(port)
	}

	wg.Wait()
	close(results)

	for status := range results {
		if status.Open {
			result.Ports = append(result.Ports, status)
		}
	}

	log.Printf("完成扫描主机 %s 从端口 %d 到 %d，共发现 %d 个开放端口", host, startPort, endPort, len(result.Ports))
	return result
}

// ScanCommonPorts 扫描主机的常用端口
func ScanCommonPorts(host string, timeout time.Duration, concurrency int) PortScanResult {
	result := PortScanResult{
		Host:  host,
		Ports: []PortStatus{},
	}

	// 检查主机名是否有效
	_, err := net.LookupHost(host)
	if err != nil {
		log.Printf("无法解析主机名 %s: %v", host, err)
		result.Error = fmt.Sprintf("无法解析主机名: %v", err)
		return result
	}

	var wg sync.WaitGroup
	results := make(chan PortStatus, len(commonPorts))
	sem := make(chan struct{}, concurrency)

	for port := range commonPorts {
		wg.Add(1)
		go func(p int) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()
			results <- ScanPort(host, p, timeout)
		}(port)
	}

	wg.Wait()
	close(results)

	for status := range results {
		if status.Open {
			result.Ports = append(result.Ports, status)
		}
	}

	log.Printf("完成扫描主机 %s 的常用端口，共发现 %d 个开放端口", host, len(result.Ports))
	return result
}

// ScanSpecificPorts 扫描主机的指定端口列表
func ScanSpecificPorts(host string, ports []int, timeout time.Duration, concurrency int) PortScanResult {
	result := PortScanResult{
		Host:  host,
		Ports: []PortStatus{},
	}

	// 检查主机名是否有效
	_, err := net.LookupHost(host)
	if err != nil {
		log.Printf("无法解析主机名 %s: %v", host, err)
		result.Error = fmt.Sprintf("无法解析主机名: %v", err)
		return result
	}

	var wg sync.WaitGroup
	results := make(chan PortStatus, len(ports))
	sem := make(chan struct{}, concurrency)

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()
			results <- ScanPort(host, p, timeout)
		}(port)
	}

	wg.Wait()
	close(results)

	for status := range results {
		if status.Open {
			result.Ports = append(result.Ports, status)
		}
	}

	log.Printf("完成扫描主机 %s 的指定端口列表，共发现 %d 个开放端口", host, len(result.Ports))
	return result
}
