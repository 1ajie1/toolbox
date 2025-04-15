package netdiag

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

// IPInfo 表示IP地址相关信息
type IPInfo struct {
	IP         string `json:"ip"`
	City       string `json:"city"`
	Region     string `json:"region"`
	Country    string `json:"country"`
	Location   string `json:"loc"`
	ISP        string `json:"org"`
	PostalCode string `json:"postal"`
	Timezone   string `json:"timezone"`
}

// LocalIPInfo 表示本地IP地址信息
type LocalIPInfo struct {
	InterfaceName string
	IPAddress     string
	MACAddress    string
	IsIPv4        bool
	IsUp          bool
}

// GetPublicIP 获取公共IP地址
func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetIPInfo 获取IP地址的详细信息
func GetIPInfo(ip string) (IPInfo, error) {
	var info IPInfo

	// 如果IP为空，则获取本机公网IP
	if ip == "" {
		var err error
		ip, err = GetPublicIP()
		if err != nil {
			return info, err
		}
	}

	// 使用ipinfo.io API获取IP详细信息
	url := fmt.Sprintf("https://ipinfo.io/%s/json", ip)
	resp, err := http.Get(url)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

// GetLocalIPs 获取本地所有网络接口的IP地址
func GetLocalIPs() ([]LocalIPInfo, error) {
	var results []LocalIPInfo

	interfaces, err := net.Interfaces()
	if err != nil {
		return results, err
	}

	for _, iface := range interfaces {
		// 跳过禁用的接口
		if (iface.Flags & net.FlagUp) == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 跳过回环地址
			if ip.IsLoopback() {
				continue
			}

			info := LocalIPInfo{
				InterfaceName: iface.Name,
				IPAddress:     ip.String(),
				MACAddress:    iface.HardwareAddr.String(),
				IsIPv4:        ip.To4() != nil,
				IsUp:          (iface.Flags & net.FlagUp) != 0,
			}

			results = append(results, info)
		}
	}

	return results, nil
}

// IsPrivateIP 检查IP是否为私有IP地址
func IsPrivateIP(ipStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("无效的IP地址: %s", ipStr)
	}

	// 检查是否为私有IP地址
	return ip.IsPrivate(), nil
}
