package main

import (
	"context"
	"fmt"
	"gui/internal/network"
	"time"
)

// App struct
type App struct {
	ctx     context.Context
	Network *network.NetworkService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		Network: network.NewNetworkService(nil),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Network = network.NewNetworkService(ctx)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// NetworkPing 执行ping操作
func (a *App) NetworkPing(host string, count int, interval float64) (*network.PingResult, error) {
	return a.Network.Ping(host, count, time.Duration(interval*float64(time.Second)))
}

// NetworkStopPing 停止ping操作
func (a *App) NetworkStopPing() error {
	return a.Network.StopPing()
}

// NetworkDNSLookup 执行DNS查询
func (a *App) NetworkDNSLookup(domain string) (*network.DNSResult, error) {
	return a.Network.DNSLookup(domain)
}

// NetworkGetSystemDNSServers 获取系统DNS服务器列表
func (a *App) NetworkGetSystemDNSServers() []string {
	return a.Network.GetSystemDNSServers()
}

// NetworkStartSpeedTest 启动速度测试服务器
func (a *App) NetworkStartSpeedTest(config *network.SpeedTestConfig) error {
	return a.Network.StartSpeedTest(config)
}

// NetworkStopSpeedTest 停止速度测试服务器
func (a *App) NetworkStopSpeedTest() error {
	return a.Network.StopSpeedTest()
}
