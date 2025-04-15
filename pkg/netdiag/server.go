package netdiag

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

// SpeedTestServer 表示速度测试服务器的配置
type SpeedTestServer struct {
	Port     int    // 服务器监听的端口
	Host     string // 服务器绑定的主机地址
	DataSize int    // 用于测试的数据大小（MB）
}

// 默认配置
var defaultServer = &SpeedTestServer{
	Port:     8080,
	Host:     "localhost",
	DataSize: 10, // 10MB
}

// 随机生成的测试数据
var testData []byte

// StartSpeedTestServer 启动一个本地的速度测试服务器
func StartSpeedTestServer(config *SpeedTestServer) error {
	if config == nil {
		config = defaultServer
	}

	// 初始化随机测试数据
	rand.Seed(time.Now().UnixNano())
	testData = make([]byte, config.DataSize*1024*1024) // 创建DataSize MB的数据
	rand.Read(testData)                                // 填充随机数据

	// 设置HTTP处理函数
	http.HandleFunc("/download", handleDownload)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/ping", handlePing)

	// 启动HTTP服务器
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	fmt.Printf("启动速度测试服务器在 %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// 处理下载测试请求
func handleDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
	w.Write(testData)
}

// 处理上传测试请求
func handleUpload(w http.ResponseWriter, r *http.Request) {
	// 消费请求体中的所有数据
	limitedReader := http.MaxBytesReader(w, r.Body, 100*1024*1024) // 最大100MB
	_, err := io.ReadAll(limitedReader)
	if err != nil {
		http.Error(w, "读取上传数据失败", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// 处理Ping测试请求
func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
