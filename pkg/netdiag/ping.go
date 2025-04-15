package netdiag

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// PingResult 表示ping操作的结果
type PingResult struct {
	Destination    string
	Success        bool
	AvgLatency     string
	PacketLoss     string
	Error          string
	DetailedOutput []string // 每次ping的详细输出
}

// PingOptions Ping操作的选项
type PingOptions struct {
	Count    int           // 要发送的Ping包数量
	Interval time.Duration // Ping的间隔时间，单位为秒
}

// Ping 函数执行ping操作并返回结果，支持实时输出
func Ping(host string, options PingOptions, callback func(string)) (PingResult, error) {
	result := PingResult{Destination: host}
	result.DetailedOutput = make([]string, 0)

	// 检查主机名是否有效
	_, err := net.LookupHost(host)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("无法解析主机名: %v", err)
		return result, err
	}

	// 根据操作系统选择合适的ping命令
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		intervalArgs := []string{}
		if options.Interval > 0 {
			// Windows ping的间隔以毫秒为单位
			intervalMs := int(options.Interval.Seconds() * 1000)
			intervalArgs = []string{"-w", fmt.Sprintf("%d", intervalMs)}
		}
		cmd = exec.Command("ping", append([]string{"-n", fmt.Sprintf("%d", options.Count)}, append(intervalArgs, host)...)...)
	} else {
		intervalArgs := []string{}
		if options.Interval > 0 {
			// Linux/macOS ping的间隔以秒为单位
			intervalArgs = []string{"-i", fmt.Sprintf("%.1f", options.Interval.Seconds())}
		}
		cmd = exec.Command("ping", append([]string{"-c", fmt.Sprintf("%d", options.Count)}, append(intervalArgs, host)...)...)
	}

	// 创建一个命令管道来获取输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("创建管道失败: %v", err)
		return result, err
	}

	// 启动命令
	if err = cmd.Start(); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("启动ping命令失败: %v", err)
		return result, err
	}

	// 读取输出
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		result.DetailedOutput = append(result.DetailedOutput, line)

		// 如果提供了回调函数，则调用它
		if callback != nil {
			callback(line)
		}
	}

	// 等待命令完成
	err = cmd.Wait()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("ping命令执行失败: %v", err)
		return result, err
	}

	// 解析结果
	output := strings.Join(result.DetailedOutput, "\n")
	result.Success = true

	// 提取平均延迟
	if strings.Contains(output, "Average") || strings.Contains(output, "平均") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Average") || strings.Contains(line, "平均") || strings.Contains(line, "avg") {
				// 处理不同格式的平均延迟输出
				if strings.Contains(line, "=") {
					// Windows 格式: "平均 = 28ms"
					parts := strings.Split(line, "=")
					if len(parts) > 1 {
						latencyPart := strings.TrimSpace(parts[len(parts)-1])
						if strings.Contains(latencyPart, "/") {
							// 处理包含多个值的格式 (min/avg/max)
							latencyParts := strings.Split(latencyPart, "/")
							if len(latencyParts) >= 2 {
								result.AvgLatency = strings.TrimSpace(latencyParts[1]) + "ms"
							}
						} else {
							// 单个值格式
							result.AvgLatency = latencyPart
						}
					}
				} else if strings.Contains(line, "min/avg/max") {
					// Linux 格式: "rtt min/avg/max/mdev = 27.963/28.209/28.855/0.373 ms"
					parts := strings.Split(line, "=")
					if len(parts) > 1 {
						statsParts := strings.Split(strings.TrimSpace(parts[1]), "/")
						if len(statsParts) >= 2 {
							result.AvgLatency = statsParts[1] + " ms"
						}
					}
				}
				break
			}
		}
	}

	// 提取丢包率
	if strings.Contains(output, "packet loss") || strings.Contains(output, "丢失") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "packet loss") || strings.Contains(line, "丢失") {
				// 从行中提取百分比
				percentRegex := regexp.MustCompile(`(\d+)%`)
				matches := percentRegex.FindStringSubmatch(line)
				if len(matches) > 1 {
					result.PacketLoss = matches[1] + "%"
				}
				break
			}
		}
	}

	return result, nil
}

// SimplePing 是原来Ping函数的简化版本，保持向后兼容
func SimplePing(host string, count int) (PingResult, error) {
	// 调用新的Ping函数，但不使用回调
	options := PingOptions{
		Count:    count,
		Interval: time.Duration(1) * time.Second, // 默认间隔1秒
	}
	return Ping(host, options, nil)
}
