package network

import (
	"context"
	"fmt"
	"sync"
	"time"
	"toolbox/pkg/netdiag"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// NetworkService GUI网络服务结构体
type NetworkService struct {
	ctx context.Context
	stopChan chan struct{}
	mutex sync.Mutex
}

// NewNetworkService 创建新的网络服务实例
func NewNetworkService(ctx context.Context) *NetworkService {
	return &NetworkService{
		ctx: ctx,
		stopChan: make(chan struct{}),
	}
}

// PingResult GUI展示用的Ping结果结构
type PingResult struct {
	Success     bool     `json:"success"`
	AvgLatency  string   `json:"avgLatency"`
	PacketLoss  string   `json:"packetLoss"`
	Error       string   `json:"error"`
	OutputLines []string `json:"outputLines"`
}

// DNSResult GUI展示用的DNS查询结果结构
type DNSResult struct {
	Success    bool     `json:"success"`
	Domain     string   `json:"domain"`
	IPList     []string `json:"ipList"`
	Error      string   `json:"error"`
	ServerUsed string   `json:"serverUsed"`
}

// SpeedTestConfig GUI配置速度测试服务器的结构
type SpeedTestConfig struct {
	Port     int    `json:"port"`
	Host     string `json:"host"`
	DataSize int    `json:"dataSize"`
}

// Ping 执行ping操作并返回GUI友好的结果
func (s *NetworkService) Ping(host string, count int, interval time.Duration) (*PingResult, error) {
	s.mutex.Lock()
	// 关闭之前的stopChan（如果存在）
	if s.stopChan != nil {
		close(s.stopChan)
	}
	s.stopChan = make(chan struct{})
	currentStopChan := s.stopChan
	s.mutex.Unlock()

	options := netdiag.PingOptions{
		Count:    count,
		Interval: interval,
		StopChan: currentStopChan, // 传递停止通道到ping选项中
	}

	var outputLines []string
	result, err := netdiag.Ping(host, options, func(output string) {
		select {
		case <-currentStopChan:
			return
		default:
			outputLines = append(outputLines, output)
			runtime.EventsEmit(s.ctx, "ping:output", output)
		}
	})

	// 检查是否是由于停止信号而退出
	select {
	case <-currentStopChan:
		return &PingResult{
			Success: true,
			Error:   "操作已停止",
			OutputLines: outputLines,
		}, nil
	default:
		if err != nil {
			return &PingResult{
				Success: false,
				Error:   err.Error(),
				OutputLines: outputLines,
			}, nil
		}
	}

	return &PingResult{
		Success:     true,
		AvgLatency:  result.AvgLatency,
		PacketLoss:  result.PacketLoss,
		OutputLines: outputLines,
	}, nil
}

// StopPing 停止当前的ping操作
func (s *NetworkService) StopPing() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.stopChan != nil {
		close(s.stopChan)
		s.stopChan = make(chan struct{})
	}
	return nil
}

// DNSLookup 执行DNS查询并返回GUI友好的结果
func (s *NetworkService) DNSLookup(domain string) (*DNSResult, error) {
	result, err := netdiag.LookupIP(domain, "")
	if err != nil {
		return &DNSResult{
			Success: false,
			Domain:  domain,
			Error:   err.Error(),
		}, nil
	}

	var ipList []string
	for _, record := range result.Records {
		ipList = append(ipList, record.Value)
	}

	return &DNSResult{
		Success:    true,
		Domain:     domain,
		IPList:     ipList,
		ServerUsed: result.ServerUsed,
	}, nil
}

// GetSystemDNSServers 获取系统DNS服务器列表
func (s *NetworkService) GetSystemDNSServers() []string {
	return netdiag.GetSystemDNSServers()
}

// StartSpeedTest 启动速度测试服务器
func (s *NetworkService) StartSpeedTest(config *SpeedTestConfig) error {
	serverConfig := &netdiag.SpeedTestServer{
		Port:     config.Port,
		Host:     config.Host,
		DataSize: config.DataSize,
	}
	return netdiag.StartSpeedTestServer(serverConfig)
}

// StopSpeedTest 停止速度测试服务器
func (s *NetworkService) StopSpeedTest() error {
	// TODO: 实现停止服务器的逻辑
	return fmt.Errorf("未实现")
} 