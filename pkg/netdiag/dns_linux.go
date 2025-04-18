//go:build !windows
// +build !windows

package netdiag

import "fmt"

// getWindowsDNSServers 在非Windows平台上的空实现
// 这个函数是为了让编译通过，在Linux/macOS上不会被调用
func getWindowsDNSServers() ([]string, error) {
	return nil, fmt.Errorf("不支持的平台")
}
