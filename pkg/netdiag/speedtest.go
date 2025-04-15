package netdiag

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SpeedTestResult 表示网络速度测试结果
type SpeedTestResult struct {
	DownloadSpeed float64 // 单位: Mbps
	UploadSpeed   float64 // 单位: Mbps
	Latency       float64 // 单位: ms
	ServerName    string
	Error         string
}

// 默认测试服务器URL
const (
	defaultDownloadURL = "http://localhost:8080/download" // 本地下载测试URL
	defaultUploadURL   = "http://localhost:8080/upload"   // 本地上传测试URL
	defaultPingURL     = "http://localhost:8080/ping"     // 本地Ping测试URL
)

// TestDownloadSpeed 测试下载速度
func TestDownloadSpeed(url string) (float64, error) {
	if url == "" {
		url = defaultDownloadURL
	}

	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// 将响应体读取到内存中以计算下载速度
	// 使用io.Copy到io.Discard避免额外的内存分配
	bytes, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return 0, err
	}

	duration := time.Since(start).Seconds()

	// 计算Mbps (兆比特每秒)
	// bytes * 8 为比特数，除以1000000为兆比特，除以耗时为每秒
	mbps := (float64(bytes) * 8) / 1000000 / duration

	return mbps, nil
}

// TestUploadSpeed 测试上传速度
func TestUploadSpeed(url string, sizeMB int) (float64, error) {
	if url == "" {
		url = defaultUploadURL
	}

	if sizeMB <= 0 {
		sizeMB = 10 // 默认上传10MB数据
	}

	// 生成要上传的数据
	data := make([]byte, sizeMB*1000000) // 生成sizeMB兆字节的数据

	start := time.Now()

	// 创建一个简单的Reader来读取数据
	dataReader := io.NopCloser(bytes.NewReader(data))

	// 执行POST请求进行上传
	resp, err := http.Post(url, "application/octet-stream", dataReader)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// 读取响应以确保请求完成
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return 0, err
	}

	duration := time.Since(start).Seconds()

	// 计算Mbps (兆比特每秒)
	mbps := (float64(len(data)) * 8) / 1000000 / duration

	return mbps, nil
}

// TestLatency 测试网络延迟
func TestLatency(url string, count int) (float64, error) {
	if url == "" {
		url = defaultPingURL
	}

	if count <= 0 {
		count = 5 // 默认测试5次取平均值
	}

	var totalLatency float64

	for i := 0; i < count; i++ {
		start := time.Now()

		resp, err := http.Get(url)
		if err != nil {
			return 0, err
		}

		// 读取响应以确保请求完成
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			resp.Body.Close()
			return 0, err
		}
		resp.Body.Close()

		latency := time.Since(start).Milliseconds()
		totalLatency += float64(latency)

		// 等待一小段时间再进行下一次测试
		time.Sleep(100 * time.Millisecond)
	}

	// 计算平均延迟
	avgLatency := totalLatency / float64(count)

	return avgLatency, nil
}

// RunSpeedTest 执行完整的网络速度测试
func RunSpeedTest() SpeedTestResult {
	result := SpeedTestResult{
		ServerName: "本地测试服务器",
	}

	// 测试延迟
	latency, err := TestLatency("", 5)
	if err != nil {
		result.Error = fmt.Sprintf("测试延迟失败: %v", err)
		return result
	}
	result.Latency = latency

	// 测试下载速度
	downloadSpeed, err := TestDownloadSpeed("")
	if err != nil {
		result.Error = fmt.Sprintf("测试下载速度失败: %v", err)
		return result
	}
	result.DownloadSpeed = downloadSpeed

	// 测试上传速度
	uploadSpeed, err := TestUploadSpeed("", 5)
	if err != nil {
		result.Error = fmt.Sprintf("测试上传速度失败: %v", err)
		return result
	}
	result.UploadSpeed = uploadSpeed

	return result
}
