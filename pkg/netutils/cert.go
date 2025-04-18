package netutils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

// CertInfo 存储证书的详细信息
type CertInfo struct {
	Subject          string    // 证书主体
	Issuer           string    // 颁发者
	NotBefore        time.Time // 生效时间
	NotAfter         time.Time // 过期时间
	DNSNames         []string  // DNS名称列表
	SerialNumber     string    // 序列号
	SignatureAlg     string    // 签名算法
	PublicKeyAlg     string    // 公钥算法
	Version          int       // 证书版本
	IsCA             bool      // 是否为CA证书
	RemainingDays    int       // 剩余有效天数
	HasTrustedIssuer bool      // 是否由受信任的CA颁发
}

// CertChecker 证书检查器
type CertChecker struct {
	FilePath string // 证书文件路径
}

// NewCertChecker 创建新的证书检查器
func NewCertChecker(filePath string) *CertChecker {
	return &CertChecker{
		FilePath: filePath,
	}
}

// CheckCertificate 检查证书文件
func (c *CertChecker) CheckCertificate() ([]*CertInfo, error) {
	// 读取证书文件
	certData, err := ioutil.ReadFile(c.FilePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取证书文件: %v", err)
	}

	var certs []*CertInfo
	var block *pem.Block
	var rest []byte = certData

	// 解析证书链中的所有证书
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析证书失败: %v", err)
		}

		// 验证证书链
		opts := x509.VerifyOptions{
			Roots: nil, // 使用系统根证书
		}
		_, err = cert.Verify(opts)
		hasTrustedIssuer := err == nil

		// 计算剩余有效天数
		remainingDays := int(time.Until(cert.NotAfter).Hours() / 24)

		// 添加证书信息
		certInfo := &CertInfo{
			Subject:          formatName(cert.Subject.String()),
			Issuer:           formatName(cert.Issuer.String()),
			NotBefore:        cert.NotBefore,
			NotAfter:         cert.NotAfter,
			DNSNames:         cert.DNSNames,
			SerialNumber:     fmt.Sprintf("%X", cert.SerialNumber),
			SignatureAlg:     cert.SignatureAlgorithm.String(),
			PublicKeyAlg:     cert.PublicKeyAlgorithm.String(),
			Version:          cert.Version,
			IsCA:             cert.IsCA,
			RemainingDays:    remainingDays,
			HasTrustedIssuer: hasTrustedIssuer,
		}

		certs = append(certs, certInfo)

		if len(rest) == 0 {
			break
		}
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("未在文件中找到有效的证书")
	}

	return certs, nil
}

// ValidateCertificate 验证证书的有效性
func (c *CertChecker) ValidateCertificate() ([]string, error) {
	certs, err := c.CheckCertificate()
	if err != nil {
		return nil, err
	}

	var issues []string

	for i, cert := range certs {
		certNum := ""
		if len(certs) > 1 {
			certNum = fmt.Sprintf("证书 #%d: ", i+1)
		}

		// 检查证书是否已经过期
		if time.Now().After(cert.NotAfter) {
			issues = append(issues, fmt.Sprintf("%s已过期，过期时间：%v", certNum, cert.NotAfter.Format("2006-01-02")))
		}

		// 检查证书是否还未生效
		if time.Now().Before(cert.NotBefore) {
			issues = append(issues, fmt.Sprintf("%s尚未生效，生效时间：%v", certNum, cert.NotBefore.Format("2006-01-02")))
		}

		// 检查证书剩余有效期
		if cert.RemainingDays < 30 {
			issues = append(issues, fmt.Sprintf("%s即将过期，剩余 %d 天", certNum, cert.RemainingDays))
		}

		// 检查证书是否由受信任的CA颁发
		if !cert.HasTrustedIssuer && !cert.IsCA {
			issues = append(issues, fmt.Sprintf("%s不是由受信任的CA颁发", certNum))
		}
	}

	return issues, nil
}

// formatName 格式化证书名称
func formatName(name string) string {
	// 移除多余的空格和换行符
	name = strings.ReplaceAll(name, "\n", " ")
	name = strings.Join(strings.Fields(name), " ")
	return name
}

// CertConfig 证书生成配置
type CertConfig struct {
	CommonName   string   // 通用名称（域名）
	Organization []string // 组织名称
	Country      []string // 国家代码
	Province     []string // 省份
	Locality     []string // 城市
	DNSNames     []string // 额外的DNS名称
	IPAddresses  []string // IP地址
	ValidDays    int      // 有效期（天）
	IsCA         bool     // 是否为CA证书
	KeySize      int      // RSA密钥长度
	SignerCert   string   // 签名者证书文件（可选）
	SignerKey    string   // 签名者私钥文件（可选）
}

// GenerateCertificate 生成证书和私钥
func GenerateCertificate(config CertConfig, certFile, keyFile string) error {
	// 生成私钥
	priv, err := generatePrivateKey(config.KeySize)
	if err != nil {
		return fmt.Errorf("生成私钥失败: %v", err)
	}

	// 解析IP地址
	var ips []net.IP
	for _, ip := range config.IPAddresses {
		if parsedIP := net.ParseIP(ip); parsedIP != nil {
			ips = append(ips, parsedIP)
		}
	}

	// 准备证书模板
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(config.ValidDays) * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("生成序列号失败: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   config.CommonName,
			Organization: config.Organization,
			Country:      config.Country,
			Province:     config.Province,
			Locality:     config.Locality,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              config.DNSNames,
		IPAddresses:           ips,
	}

	if config.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		template.ExtKeyUsage = nil // CA证书不需要ExtKeyUsage
	}

	// 确定签名者证书和私钥
	var signerCert *x509.Certificate
	var signerKey crypto.PrivateKey

	if config.SignerCert != "" && config.SignerKey != "" {
		// 使用提供的CA证书签名
		signerCertData, err := ioutil.ReadFile(config.SignerCert)
		if err != nil {
			return fmt.Errorf("读取签名者证书失败: %v", err)
		}

		block, _ := pem.Decode(signerCertData)
		if block == nil {
			return fmt.Errorf("解析签名者证书失败")
		}

		signerCert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("解析签名者证书失败: %v", err)
		}

		signerKeyData, err := ioutil.ReadFile(config.SignerKey)
		if err != nil {
			return fmt.Errorf("读取签名者私钥失败: %v", err)
		}

		block, _ = pem.Decode(signerKeyData)
		if block == nil {
			return fmt.Errorf("解析签名者私钥失败")
		}

		signerKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("解析签名者私钥失败: %v", err)
		}
	} else {
		// 自签名
		signerCert = template
		signerKey = priv
	}

	// 创建证书
	derBytes, err := x509.CreateCertificate(rand.Reader, template, signerCert, &priv.PublicKey, signerKey)
	if err != nil {
		return fmt.Errorf("生成证书失败: %v", err)
	}

	// 保存证书
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("创建证书文件失败: %v", err)
	}
	defer certOut.Close()

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return fmt.Errorf("写入证书文件失败: %v", err)
	}

	// 保存私钥
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("创建私钥文件失败: %v", err)
	}
	defer keyOut.Close()

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return fmt.Errorf("写入私钥文件失败: %v", err)
	}

	return nil
}

// generatePrivateKey 生成RSA私钥
func generatePrivateKey(bits int) (*rsa.PrivateKey, error) {
	if bits == 0 {
		bits = 2048 // 默认密钥长度
	}
	return rsa.GenerateKey(rand.Reader, bits)
}
