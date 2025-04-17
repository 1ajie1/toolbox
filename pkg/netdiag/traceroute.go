package netdiag

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// TracerouteResult 表示路由跟踪的结果
type TracerouteResult struct {
	Hops     []HopInfo // 路由跳数
	Error    string
	TargetIP string // 目标IP地址
}

// HopInfo 表示路由中的一跳
type HopInfo struct {
	Number int      // 跳数
	IP     string   // IP地址
	Name   string   // 主机名
	RTT    []string // 往返时间
}

// RealTimeHopCallback 定义实时回调函数类型，用于在获取每一跳信息时立即返回结果
type RealTimeHopCallback func(hop HopInfo)

// TracerouteOptions 表示路由跟踪的选项
type TracerouteOptions struct {
	MaxHops          int                 // 最大跳数
	Timeout          time.Duration       // 超时时间
	PacketSize       int                 // 数据包大小
	RealTimeCallback RealTimeHopCallback // 实时回调，每个hop有结果就立即调用
}

// Traceroute 执行路由跟踪
func Traceroute(host string, options TracerouteOptions) (TracerouteResult, error) {
	// 根据平台选择不同的实现
	if runtime.GOOS == "windows" {
		return windowsTracerouteImpl(host, options)
	} else {
		return unixTracerouteImpl(host, options)
	}
}

// unixTracerouteImpl 在Unix/Linux平台上实现traceroute
func unixTracerouteImpl(host string, options TracerouteOptions) (TracerouteResult, error) {
	result := TracerouteResult{
		Hops: make([]HopInfo, 0),
	}

	// 解析目标主机
	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		result.Error = fmt.Sprintf("无法解析主机名: %v", err)
		return result, err
	}

	// 设置目标IP
	result.TargetIP = ipAddr.String()

	// 设置默认选项
	if options.MaxHops <= 0 {
		options.MaxHops = 30
	}
	if options.Timeout <= 0 {
		options.Timeout = 3 * time.Second
	}
	if options.PacketSize <= 0 {
		options.PacketSize = 60
	}

	// 创建原始套接字
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, IPPROTO_ICMP)
	if err != nil {
		result.Error = fmt.Sprintf("创建原始套接字失败: %v", err)
		return result, err
	}
	defer syscall.Close(fd)

	// 设置TTL
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_TTL, options.MaxHops)
	if err != nil {
		result.Error = fmt.Sprintf("设置TTL失败: %v", err)
		return result, err
	}

	// 创建ICMP数据包
	msg := make([]byte, options.PacketSize)
	msg[0] = 8 // ICMP Echo Request
	msg[1] = 0 // Code
	msg[2] = 0 // Checksum
	msg[3] = 0 // Checksum
	msg[4] = 0 // Identifier
	msg[5] = 0 // Identifier
	msg[6] = 0 // Sequence Number
	msg[7] = 0 // Sequence Number

	// 计算校验和
	checkSum := checkSum(msg)
	msg[2] = byte(checkSum >> 8)
	msg[3] = byte(checkSum & 0xff)

	// 逐跳测试
	for ttl := 1; ttl <= options.MaxHops; ttl++ {
		// 设置TTL
		err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_TTL, ttl)
		if err != nil {
			result.Error = fmt.Sprintf("设置TTL失败: %v", err)
			return result, err
		}

		// 记录开始时间
		start := time.Now()

		// 发送ICMP包
		addr := &syscall.SockaddrInet4{
			Port: 0,
			Addr: [4]byte{ipAddr.IP[0], ipAddr.IP[1], ipAddr.IP[2], ipAddr.IP[3]},
		}
		err = syscall.Sendto(fd, msg, 0, addr)
		if err != nil {
			result.Error = fmt.Sprintf("发送ICMP包失败: %v", err)
			return result, err
		}

		// 设置超时
		tv := syscall.NsecToTimeval(options.Timeout.Nanoseconds())
		err = syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, SO_RCVTIMEO, &tv)
		if err != nil {
			result.Error = fmt.Sprintf("设置超时失败: %v", err)
			return result, err
		}

		// 接收响应
		reply := make([]byte, 1500)
		_, _, err = syscall.Recvfrom(fd, reply, 0)
		if err != nil {
			// 超时或错误，记录为超时
			hop := HopInfo{
				Number: ttl,
				IP:     "*",
				Name:   "*",
				RTT:    []string{"*"},
			}

			// 调用实时回调（如果有）
			if options.RealTimeCallback != nil {
				options.RealTimeCallback(hop)
			}

			result.Hops = append(result.Hops, hop)
			continue
		}

		// 计算延迟
		latency := time.Since(start)
		rtt := fmt.Sprintf("%.2f ms", float64(latency.Microseconds())/1000.0)

		// 获取响应IP
		replyIP := net.IP(reply[12:16]).String()

		// 尝试获取主机名
		hostname := "*"
		names, err := net.LookupAddr(replyIP)
		if err == nil && len(names) > 0 {
			hostname = names[0]
		}

		// 记录这一跳
		hop := HopInfo{
			Number: ttl,
			IP:     replyIP,
			Name:   hostname,
			RTT:    []string{rtt},
		}

		// 调用实时回调（如果有）
		if options.RealTimeCallback != nil {
			options.RealTimeCallback(hop)
		}

		result.Hops = append(result.Hops, hop)

		// 如果到达目标，结束
		if replyIP == ipAddr.String() {
			break
		}
	}

	return result, nil
}

// windowsTracerouteImpl Windows平台下的traceroute实现
func windowsTracerouteImpl(host string, options TracerouteOptions) (TracerouteResult, error) {
	result := TracerouteResult{
		Hops: make([]HopInfo, 0),
	}

	// 解析目标主机
	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		result.Error = fmt.Sprintf("无法解析主机名: %v", err)
		return result, err
	}

	// 设置目标IP
	result.TargetIP = ipAddr.String()

	// 设置默认选项
	if options.MaxHops <= 0 {
		options.MaxHops = 30
	}
	if options.Timeout <= 0 {
		options.Timeout = 3 * time.Second
	}
	if options.PacketSize <= 0 {
		options.PacketSize = 60
	}

	// 使用统一方法创建连接
	conn, err := net.DialIP("ip4:icmp", nil, ipAddr)
	if err != nil {
		result.Error = fmt.Sprintf("创建连接失败: %v", err)
		return result, err
	}
	defer conn.Close()

	// 设置TTL
	ipConn := ipv4.NewConn(conn)

	// 创建ICMP消息
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: make([]byte, options.PacketSize),
		},
	}

	// 将消息序列化为字节
	wb, err := wm.Marshal(nil)
	if err != nil {
		result.Error = fmt.Sprintf("序列化ICMP消息失败: %v", err)
		return result, err
	}

	// 逐跳测试
	for ttl := 1; ttl <= options.MaxHops; ttl++ {
		// 设置当前TTL
		if err := ipConn.SetTTL(ttl); err != nil {
			result.Error = fmt.Sprintf("设置TTL失败: %v", err)
			return result, err
		}

		// 设置超时
		conn.SetDeadline(time.Now().Add(options.Timeout))

		// 记录开始时间
		start := time.Now()

		// 发送ICMP包
		if _, err := conn.Write(wb); err != nil {
			result.Error = fmt.Sprintf("发送ICMP包失败: %v", err)
			return result, err
		}

		// 接收响应
		rb := make([]byte, 1500)
		n, err := conn.Read(rb)
		if err != nil {
			// 超时或错误，记录为超时
			hop := HopInfo{
				Number: ttl,
				IP:     "*",
				Name:   "*",
				RTT:    []string{"*"},
			}

			// 调用实时回调（如果有）
			if options.RealTimeCallback != nil {
				options.RealTimeCallback(hop)
			}

			result.Hops = append(result.Hops, hop)
			continue
		}

		// 计算延迟
		latency := time.Since(start)
		rtt := fmt.Sprintf("%.2f ms", float64(latency.Microseconds())/1000.0)

		// 解析响应
		_, err = icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rb[:n])

		// 获取远程IP (从ICMP响应头提取)
		remoteIP := net.IP(rb[12:16]).String()

		// 尝试获取主机名
		hostname := "*"
		names, err := net.LookupAddr(remoteIP)
		if err == nil && len(names) > 0 {
			hostname = names[0]
		}

		// 记录这一跳
		hop := HopInfo{
			Number: ttl,
			IP:     remoteIP,
			Name:   hostname,
			RTT:    []string{rtt},
		}

		// 调用实时回调（如果有）
		if options.RealTimeCallback != nil {
			options.RealTimeCallback(hop)
		}

		result.Hops = append(result.Hops, hop)

		// 如果到达目标，结束
		if remoteIP == ipAddr.String() {
			break
		}
	}

	return result, nil
}

// checkSum 计算ICMP校验和
func checkSum(msg []byte) uint16 {
	sum := 0
	for i := 0; i < len(msg); i += 2 {
		sum += int(msg[i])<<8 | int(msg[i+1])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += sum >> 16
	return uint16(^sum)
}
