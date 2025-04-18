package network

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"toolbox/pkg/netutils"

	"github.com/spf13/cobra"
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "证书工具",
	Long: `证书工具，用于检查和生成证书。

支持的功能：
1. 检查证书信息（有效期、颁发机构、证书链等）
2. 生成自签名证书（用于开发测试）`,
}

var certCheckCmd = &cobra.Command{
	Use:   "check [证书文件]",
	Short: "检查证书文件",
	Long: `检查证书文件的详细信息，包括有效期、颁发机构、证书链等。
支持检查单个证书文件或包含完整证书链的文件。

示例:
  # 检查单个证书文件
  %[1]s network cert check server.crt

  # 检查包含证书链的文件
  %[1]s network cert check fullchain.pem

  # 仅显示证书问题
  %[1]s network cert check server.crt --issues-only`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		certFile := args[0]
		issuesOnly, _ := cmd.Flags().GetBool("issues-only")

		checker := netutils.NewCertChecker(certFile)

		// 获取证书信息
		certs, err := checker.CheckCertificate()
		if err != nil {
			return fmt.Errorf("检查证书失败: %v", err)
		}

		// 验证证书
		issues, err := checker.ValidateCertificate()
		if err != nil {
			return fmt.Errorf("验证证书失败: %v", err)
		}

		// 如果只显示问题，且没有问题，则直接返回
		if issuesOnly && len(issues) == 0 {
			fmt.Println("证书有效，未发现问题")
			return nil
		}

		// 如果有问题，显示问题列表
		if len(issues) > 0 {
			fmt.Println("发现以下问题：")
			for _, issue := range issues {
				fmt.Printf("- %s\n", issue)
			}
			fmt.Println()
		}

		// 如果不是只显示问题，则显示完整信息
		if !issuesOnly {
			for i, cert := range certs {
				if len(certs) > 1 {
					fmt.Printf("\n证书 #%d:\n", i+1)
				} else {
					fmt.Println("证书信息：")
				}

				fmt.Printf("主体: %s\n", cert.Subject)
				fmt.Printf("颁发者: %s\n", cert.Issuer)
				fmt.Printf("生效时间: %s\n", cert.NotBefore.Format("2006-01-02 15:04:05"))
				fmt.Printf("过期时间: %s\n", cert.NotAfter.Format("2006-01-02 15:04:05"))
				fmt.Printf("剩余天数: %d\n", cert.RemainingDays)
				fmt.Printf("序列号: %s\n", cert.SerialNumber)
				fmt.Printf("签名算法: %s\n", cert.SignatureAlg)
				fmt.Printf("公钥算法: %s\n", cert.PublicKeyAlg)
				fmt.Printf("证书版本: %d\n", cert.Version)
				fmt.Printf("是否为CA: %v\n", cert.IsCA)
				fmt.Printf("是否由受信任的CA颁发: %v\n", cert.HasTrustedIssuer)

				if len(cert.DNSNames) > 0 {
					fmt.Printf("DNS名称: %s\n", strings.Join(cert.DNSNames, ", "))
				}
			}
		}

		return nil
	},
}

// askQuestion 从用户获取输入
func askQuestion(reader *bufio.Reader, question string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", question, defaultValue)
	} else {
		fmt.Printf("%s: ", question)
	}
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return defaultValue
	}
	return answer
}

// askYesNo 获取用户是否确认
func askYesNo(reader *bufio.Reader, question string, defaultYes bool) bool {
	defaultStr := "Y/n"
	if !defaultYes {
		defaultStr = "y/N"
	}
	fmt.Printf("%s [%s]: ", question, defaultStr)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer == "" {
		return defaultYes
	}
	return answer == "y" || answer == "yes"
}

var certGenerateCmd = &cobra.Command{
	Use:   "generate [名称]",
	Short: "生成自签名证书",
	Long: `生成自签名证书，用于开发测试环境。

将通过交互式问答获取证书信息。如果不确定，可以直接按回车使用默认值。

示例:
  # 启动交互式证书生成向导
  %[1]s network cert generate

  # 指定域名并启动向导
  %[1]s network cert generate example.com

  # 使用所有默认值（不推荐）
  %[1]s network cert generate example.com --no-interactive`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noInteractive, _ := cmd.Flags().GetBool("no-interactive")
		reader := bufio.NewReader(os.Stdin)
		var name string

		if len(args) > 0 {
			name = args[0]
		}

		if !noInteractive {
			fmt.Println("欢迎使用证书生成向导！")
			fmt.Println("请回答以下问题，或直接按回车使用默认值。")
			fmt.Println()

			// 1. 获取通用名称（域名）
			if name == "" {
				name = askQuestion(reader, "请输入域名（例如：example.com）", "localhost")
			}

			// 2. 确认是否需要通配符证书
			if askYesNo(reader, "是否需要通配符证书（可用于所有子域名）", false) {
				if !strings.HasPrefix(name, "*.") {
					name = "*." + strings.TrimPrefix(name, "www.")
				}
			}

			// 3. 获取其他DNS名称
			var dnsNames []string
			baseDomain := strings.TrimPrefix(name, "*.")
			if askYesNo(reader, "是否添加其他DNS名称", true) {
				fmt.Println("请输入其他DNS名称，每行一个，留空结束：")
				for {
					dns := askQuestion(reader, "> ", "")
					if dns == "" {
						break
					}
					dnsNames = append(dnsNames, dns)
				}
			}
			if !strings.HasPrefix(name, "*.") {
				dnsNames = append([]string{name}, dnsNames...)
			}
			if baseDomain != name {
				dnsNames = append(dnsNames, baseDomain)
			}

			// 4. 获取IP地址
			var ips []string
			if askYesNo(reader, "是否添加IP地址", false) {
				fmt.Println("请输入IP地址，每行一个，留空结束：")
				for {
					ip := askQuestion(reader, "> ", "")
					if ip == "" {
						break
					}
					ips = append(ips, ip)
				}
			}

			// 5. 获取有效期
			days := 3650
			if !askYesNo(reader, "是否使用默认有效期（10年）", true) {
				for {
					daysStr := askQuestion(reader, "请输入有效期（天数）", "3650")
					fmt.Sscanf(daysStr, "%d", &days)
					if days > 0 {
						break
					}
					fmt.Println("请输入大于0的数字！")
				}
			}

			// 6. 获取密钥长度
			bits := 2048
			if !askYesNo(reader, "是否使用默认密钥长度（2048位）", true) {
				for {
					bitsStr := askQuestion(reader, "请输入密钥长度（2048或4096）", "2048")
					fmt.Sscanf(bitsStr, "%d", &bits)
					if bits == 2048 || bits == 4096 {
						break
					}
					fmt.Println("请输入2048或4096！")
				}
			}

			// 7. 获取输出文件名
			certFile := askQuestion(reader, "请输入证书文件名", name+".crt")
			keyFile := askQuestion(reader, "请输入私钥文件名", name+".key")

			// 创建输出目录
			certDir := filepath.Dir(certFile)
			keyDir := filepath.Dir(keyFile)
			if certDir != "." {
				if err := os.MkdirAll(certDir, 0755); err != nil {
					return fmt.Errorf("创建证书目录失败: %v", err)
				}
			}
			if keyDir != "." {
				if err := os.MkdirAll(keyDir, 0755); err != nil {
					return fmt.Errorf("创建私钥目录失败: %v", err)
				}
			}

			// 生成证书
			config := netutils.CertConfig{
				CommonName:  name,
				DNSNames:    dnsNames,
				IPAddresses: ips,
				ValidDays:   days,
				IsCA:        false,
				KeySize:     bits,
			}

			if err := netutils.GenerateCertificate(config, certFile, keyFile); err != nil {
				return fmt.Errorf("生成证书失败: %v", err)
			}

			fmt.Printf("\n证书已生成：\n证书文件：%s\n私钥文件：%s\n", certFile, keyFile)
			return nil
		}

		// 非交互式模式
		if name == "" {
			return fmt.Errorf("非交互式模式下必须指定域名")
		}

		certFile := name
		if !strings.HasSuffix(certFile, ".crt") {
			certFile += ".crt"
		}
		keyFile := name
		if !strings.HasSuffix(keyFile, ".key") {
			keyFile += ".key"
		}

		config := netutils.CertConfig{
			CommonName: name,
			DNSNames:   []string{name},
			ValidDays:  3650,
			IsCA:       false,
			KeySize:    2048,
		}

		if err := netutils.GenerateCertificate(config, certFile, keyFile); err != nil {
			return fmt.Errorf("生成证书失败: %v", err)
		}

		fmt.Printf("证书已生成：\n证书文件：%s\n私钥文件：%s\n", certFile, keyFile)
		return nil
	},
}

func init() {
	// 检查命令的选项
	certCheckCmd.Flags().Bool("issues-only", false, "仅显示证书问题")

	// 生成命令的选项
	certGenerateCmd.Flags().Bool("no-interactive", false, "使用默认值（不进行交互）")

	certCmd.AddCommand(certCheckCmd)
	certCmd.AddCommand(certGenerateCmd)
	NetworkCmd.AddCommand(certCmd)
}
