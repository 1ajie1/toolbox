package netdiag

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

// SnifferConfig 配置网络抓包参数
type SnifferConfig struct {
	Interface   string
	Filter      string
	Timeout     time.Duration
	Output      string
	Snaplen     int    // 捕获的数据包大小
	Promiscuous bool   // 是否开启混杂模式
	Count       int    // 捕获的包数量，0表示无限制
	Verbose     bool   // 是否显示详细信息
	SavePcap    string // 保存为pcap文件
	Statistics  bool   // 是否显示统计信息
	PayloadLen  int    // 显示的载荷长度，0表示不显示
}

// PacketStats 网络包统计信息
type PacketStats struct {
	PacketCount int
	TotalBytes  int64
	StartTime   time.Time
	EndTime     time.Time
	ProtocolMap map[string]int
	SourceIPs   map[string]int
	DestIPs     map[string]int
	SourcePorts map[uint16]int
	DestPorts   map[uint16]int
	mutex       sync.Mutex
}

// NewPacketStats 创建统计对象
func NewPacketStats() *PacketStats {
	return &PacketStats{
		StartTime:   time.Now(),
		ProtocolMap: make(map[string]int),
		SourceIPs:   make(map[string]int),
		DestIPs:     make(map[string]int),
		SourcePorts: make(map[uint16]int),
		DestPorts:   make(map[uint16]int),
	}
}

// AddPacket 添加数据包统计
func (ps *PacketStats) AddPacket(packet gopacket.Packet) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.PacketCount++
	ps.TotalBytes += int64(packet.Metadata().Length)

	// 统计协议
	for _, layer := range packet.Layers() {
		ps.ProtocolMap[layer.LayerType().String()]++
	}

	// 统计IP地址
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		ps.SourceIPs[ip.SrcIP.String()]++
		ps.DestIPs[ip.DstIP.String()]++
	} else if ipLayer := packet.Layer(layers.LayerTypeIPv6); ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv6)
		ps.SourceIPs[ip.SrcIP.String()]++
		ps.DestIPs[ip.DstIP.String()]++
	}

	// 统计端口
	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		ps.SourcePorts[uint16(tcp.SrcPort)]++
		ps.DestPorts[uint16(tcp.DstPort)]++
	} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		ps.SourcePorts[uint16(udp.SrcPort)]++
		ps.DestPorts[uint16(udp.DstPort)]++
	}
}

// PrintStats 打印统计信息
func (ps *PacketStats) PrintStats() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.EndTime = time.Now()
	duration := ps.EndTime.Sub(ps.StartTime)

	fmt.Println("\n==== 网络抓包统计信息 ====")
	fmt.Printf("捕获时间: %s\n", duration.Round(time.Millisecond))
	fmt.Printf("数据包总数: %d\n", ps.PacketCount)
	fmt.Printf("总字节数: %d bytes\n", ps.TotalBytes)
	if duration.Seconds() > 0 {
		fmt.Printf("平均速率: %.2f 包/秒, %.2f bytes/秒\n",
			float64(ps.PacketCount)/duration.Seconds(),
			float64(ps.TotalBytes)/duration.Seconds())
	}

	// 打印协议分布
	fmt.Println("\n协议分布:")
	for proto, count := range ps.ProtocolMap {
		fmt.Printf("  %s: %d (%.1f%%)\n", proto, count, float64(count)*100/float64(ps.PacketCount))
	}

	// 打印最活跃的源IP (top 5)
	fmt.Println("\n最活跃的源IP地址:")
	printTopItems(ps.SourceIPs, 5)

	// 打印最活跃的目标IP (top 5)
	fmt.Println("\n最活跃的目标IP地址:")
	printTopItems(ps.DestIPs, 5)

	// 打印最活跃的端口 (top 5)
	fmt.Println("\n最活跃的端口:")
	printTopItemsUint16(ps.SourcePorts, 5)
}

// 辅助函数，打印top N的项目
func printTopItems(items map[string]int, n int) {
	// 转换为列表并排序
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range items {
		sorted = append(sorted, kv{k, v})
	}

	// 冒泡排序 (简单实现，生产环境应使用更高效排序)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Value < sorted[j+1].Value {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	// 打印Top N
	count := n
	if len(sorted) < count {
		count = len(sorted)
	}
	for i := 0; i < count; i++ {
		fmt.Printf("  %s: %d\n", sorted[i].Key, sorted[i].Value)
	}
}

// 打印top N的端口项目
func printTopItemsUint16(items map[uint16]int, n int) {
	// 转换为列表并排序
	type kv struct {
		Key   uint16
		Value int
	}
	var sorted []kv
	for k, v := range items {
		sorted = append(sorted, kv{k, v})
	}

	// 冒泡排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Value < sorted[j+1].Value {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	// 打印Top N
	count := n
	if len(sorted) < count {
		count = len(sorted)
	}
	for i := 0; i < count; i++ {
		fmt.Printf("  %d: %d\n", sorted[i].Key, sorted[i].Value)
	}
}

// StartSniffer 开始网络抓包
func StartSniffer(config SnifferConfig) error {
	// 设置默认值
	if config.Snaplen <= 0 {
		config.Snaplen = 1600
	}

	// 打开网络接口
	handle, err := pcap.OpenLive(config.Interface, int32(config.Snaplen), config.Promiscuous, config.Timeout)
	if err != nil {
		return fmt.Errorf("打开网络接口失败: %v", err)
	}
	defer handle.Close()

	// 设置过滤器
	if config.Filter != "" {
		if err := handle.SetBPFFilter(config.Filter); err != nil {
			return fmt.Errorf("设置过滤器失败: %v", err)
		}
	}

	// 创建输出文件
	var outFile *os.File
	if config.Output != "" {
		outFile, err = os.Create(config.Output)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %v", err)
		}
		defer outFile.Close()
	}

	// 创建pcap文件写入器
	var pcapWriter *pcapgo.Writer
	if config.SavePcap != "" {
		pcapFile, err := os.Create(config.SavePcap)
		if err != nil {
			return fmt.Errorf("创建pcap文件失败: %v", err)
		}
		defer pcapFile.Close()

		pcapWriter = pcapgo.NewWriter(pcapFile)
		if err := pcapWriter.WriteFileHeader(uint32(config.Snaplen), handle.LinkType()); err != nil {
			return fmt.Errorf("写入pcap文件头失败: %v", err)
		}
	}

	// 统计信息
	var stats *PacketStats
	if config.Statistics {
		stats = NewPacketStats()
	}

	// 创建信号通道，用于捕获中断信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// 创建停止通道，用于通知抓包循环退出
	stopChan := make(chan struct{})

	// 开始抓包
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetChan := packetSource.Packets()
	log.Printf("开始抓包，接口: %s, 过滤器: %s\n", config.Interface, config.Filter)

	// 启动goroutine监听中断信号
	go func() {
		<-signalChan
		log.Println("收到中断信号，正在停止抓包...")
		close(stopChan) // 通知抓包循环退出
		signal.Stop(signalChan)
	}()

	count := 0
	// 使用可中断的抓包循环
loop:
	for {
		select {
		case packet, ok := <-packetChan:
			if !ok {
				// 通道已关闭
				break loop
			}

			// 解析并显示数据包信息
			printPacketInfo(packet, config.Verbose, outFile, config.PayloadLen)

			// 写入pcap文件
			if pcapWriter != nil {
				if err := pcapWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
					log.Printf("写入pcap文件失败: %v", err)
				}
			}

			// 统计
			if stats != nil {
				stats.AddPacket(packet)
			}

			count++
			if config.Count > 0 && count >= config.Count {
				break loop
			}

		case <-stopChan:
			// 收到停止信号
			log.Println("停止抓包...")
			break loop
		}
	}

	// 打印统计信息
	if stats != nil {
		stats.PrintStats()
	}

	return nil
}

// printPacketInfo 打印数据包信息
func printPacketInfo(packet gopacket.Packet, verbose bool, outFile *os.File, payloadLen int) {
	// 获取时间戳
	timestamp := packet.Metadata().Timestamp.Format("15:04:05.000000")

	var output string

	// 解析链路层
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		eth, _ := ethernetLayer.(*layers.Ethernet)
		output += fmt.Sprintf("%s %s > %s, ", timestamp, eth.SrcMAC, eth.DstMAC)
	}

	// 解析IP层
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		output += fmt.Sprintf("IPv4 %s > %s, ", ip.SrcIP, ip.DstIP)
	}

	// 解析IPv6层
	ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
	if ipv6Layer != nil {
		ipv6, _ := ipv6Layer.(*layers.IPv6)
		output += fmt.Sprintf("IPv6 %s > %s, ", ipv6.SrcIP, ipv6.DstIP)
	}

	// 解析TCP层
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		flags := ""
		if tcp.SYN {
			flags += "S"
		}
		if tcp.ACK {
			flags += "A"
		}
		if tcp.FIN {
			flags += "F"
		}
		if tcp.RST {
			flags += "R"
		}
		if tcp.PSH {
			flags += "P"
		}
		if tcp.URG {
			flags += "U"
		}
		output += fmt.Sprintf("TCP %d > %d [%s] Seq=%d Ack=%d Win=%d",
			tcp.SrcPort, tcp.DstPort, flags, tcp.Seq, tcp.Ack, tcp.Window)
	}

	// 解析UDP层
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		output += fmt.Sprintf("UDP %d > %d Len=%d",
			udp.SrcPort, udp.DstPort, udp.Length)
	}

	// 解析ICMP层
	icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpLayer != nil {
		icmp, _ := icmpLayer.(*layers.ICMPv4)
		output += fmt.Sprintf("ICMP Type=%d Code=%d",
			icmp.TypeCode.Type(), icmp.TypeCode.Code())
	}

	// 解析ICMPv6层
	icmpv6Layer := packet.Layer(layers.LayerTypeICMPv6)
	if icmpv6Layer != nil {
		icmpv6, _ := icmpv6Layer.(*layers.ICMPv6)
		output += fmt.Sprintf("ICMPv6 Type=%d", icmpv6.TypeCode.Type())
	}

	// 解析DNS层
	dnsLayer := packet.Layer(layers.LayerTypeDNS)
	if dnsLayer != nil {
		dns, _ := dnsLayer.(*layers.DNS)
		if dns.QR {
			output += fmt.Sprintf(" DNS Response")
			for _, answer := range dns.Answers {
				output += fmt.Sprintf(" %s", string(answer.Name))
			}
		} else {
			output += fmt.Sprintf(" DNS Query")
			for _, question := range dns.Questions {
				output += fmt.Sprintf(" %s", string(question.Name))
			}
		}
	}

	// 计算数据包大小
	if packet.Metadata().Length > 0 {
		output += fmt.Sprintf(", length %d bytes", packet.Metadata().Length)
	}

	// 应用层数据
	if payloadLen > 0 {
		applicationLayer := packet.ApplicationLayer()
		if applicationLayer != nil {
			payload := applicationLayer.Payload()
			if len(payload) > 0 {
				if len(payload) > payloadLen {
					payload = payload[:payloadLen]
				}
				output += fmt.Sprintf("\n  Payload: %s", formatPayload(payload))
			}
		}
	}

	fmt.Println(output)

	// 如果详细模式，打印更多信息
	if verbose {
		fmt.Println(packet.Dump())
	}

	// 如果指定了输出文件，则写入文件
	if outFile != nil {
		if _, err := outFile.WriteString(output + "\n"); err != nil {
			log.Printf("写入文件失败: %v", err)
		}

		if verbose {
			if _, err := outFile.WriteString(packet.Dump() + "\n"); err != nil {
				log.Printf("写入详细信息到文件失败: %v", err)
			}
		}
	}
}

// formatPayload 格式化负载数据
func formatPayload(payload []byte) string {
	// 尝试显示为ASCII，如果不可打印字符太多则显示十六进制
	asciiString := make([]byte, 0, len(payload))
	unprintable := 0

	for _, b := range payload {
		if b >= 32 && b <= 126 {
			asciiString = append(asciiString, b)
		} else if b == '\r' || b == '\n' || b == '\t' {
			asciiString = append(asciiString, ' ')
		} else {
			unprintable++
			asciiString = append(asciiString, '.')
		}
	}

	// 如果不可打印字符超过30%，显示十六进制
	if float64(unprintable)/float64(len(payload)) > 0.3 {
		return hex.EncodeToString(payload)
	}

	return strings.TrimSpace(string(asciiString))
}

// ListInterfaces 列出可用的网络接口
func ListInterfaces() ([]string, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口列表失败: %v", err)
	}

	var interfaces []string
	for _, device := range devices {
		desc := device.Description
		if desc == "" {
			desc = "无描述"
		}
		info := fmt.Sprintf("%s: %s", device.Name, desc)

		// 添加IP地址信息
		for _, address := range device.Addresses {
			info += fmt.Sprintf(" [IP: %s/%d]", address.IP, address.Netmask)
		}

		interfaces = append(interfaces, info)
	}

	return interfaces, nil
}
