// Package audit - 安全审计工具
// Sprint 25: 正式发布准备 - 安全审计
package audit

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// AuditConfig 审计配置
type AuditConfig struct {
	TargetURL      string   `json:"target_url"`
	Timeout        time.Duration `json:"timeout"`
	CheckSSL       bool     `json:"check_ssl"`
	CheckHeaders   bool     `json:"check_headers"`
	CheckEndpoints bool     `json:"check_endpoints"`
	CheckFiles     bool     `json:"check_files"`
	ScanPaths      []string `json:"scan_paths"`
	ExcludedPaths  []string `json:"excluded_paths"`
}

// AuditReport 审计报告
type AuditReport struct {
	StartTime      time.Time        `json:"start_time"`
	EndTime        time.Time        `json:"end_time"`
	Duration       time.Duration    `json:"duration"`
	Summary        *AuditSummary    `json:"summary"`
	Findings       []AuditFinding   `json:"findings"`
	SSLCheck       *SSLCheckResult  `json:"ssl_check,omitempty"`
	HeaderCheck    *HeaderCheckResult `json:"header_check,omitempty"`
	EndpointCheck  *EndpointCheckResult `json:"endpoint_check,omitempty"`
	FileCheck      *FileCheckResult `json:"file_check,omitempty"`
}

// AuditSummary 审计摘要
type AuditSummary struct {
	TotalChecks    int `json:"total_checks"`
	PassedChecks   int `json:"passed_checks"`
	WarningCount   int `json:"warning_count"`
	CriticalCount  int `json:"critical_count"`
	InfoCount      int `json:"info_count"`
	SecurityScore  int `json:"security_score"` // 0-100
}

// AuditFinding 审计发现
type AuditFinding struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Severity    string `json:"severity"` // critical, high, medium, low, info
	Title       string `json:"title"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
	Reference   string `json:"reference,omitempty"`
}

// SSLCheckResult SSL检查结果
type SSLCheckResult struct {
	TLSVersion       string `json:"tls_version"`
	CipherSuite      string `json:"cipher_suite"`
	CertificateValid bool   `json:"certificate_valid"`
	ExpiresIn        string `json:"expires_in,omitempty"`
	Issuer           string `json:"issuer,omitempty"`
	SelfSigned       bool   `json:"self_signed"`
	WeakProtocol     bool   `json:"weak_protocol"`
	WeakCipher       bool   `json:"weak_cipher"`
}

// HeaderCheckResult HTTP头检查结果
type HeaderCheckResult struct {
	SecurityHeaders map[string]HeaderStatus `json:"security_headers"`
	MissingHeaders  []string                `json:"missing_headers"`
	WeakHeaders     map[string]string       `json:"weak_headers"`
}

// HeaderStatus 头状态
type HeaderStatus struct {
	Present bool   `json:"present"`
	Value   string `json:"value,omitempty"`
	Secure  bool   `json:"secure"`
}

// EndpointCheckResult 端点检查结果
type EndpointCheckResult struct {
	Endpoints []EndpointStatus `json:"endpoints"`
	Exposed   []string         `json:"exposed"`
	Protected []string         `json:"protected"`
}

// EndpointStatus 端点状态
type EndpointStatus struct {
	Path       string `json:"path"`
	Method     string `json:"method"`
	Status     int    `json:"status"`
	Accessible bool   `json:"accessible"`
	Protected  bool   `json:"protected"`
}

// FileCheckResult 文件检查结果
type FileCheckResult struct {
	SensitiveFiles []SensitiveFile `json:"sensitive_files"`
	PermissionIssues []PermissionIssue `json:"permission_issues"`
}

// SensitiveFile 敏感文件
type SensitiveFile struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	Risk        string `json:"risk"`
	Accessible  bool   `json:"accessible"`
}

// PermissionIssue 权限问题
type PermissionIssue struct {
	Path        string `json:"path"`
	CurrentPerm string `json:"current_perm"`
	ExpectedPerm string `json:"expected_perm"`
	Risk        string `json:"risk"`
}

// SecurityAuditor 安全审计器
type SecurityAuditor struct {
	config AuditConfig
	client *http.Client
	mu     sync.RWMutex
}

// NewSecurityAuditor 创建安全审计器
func NewSecurityAuditor(config AuditConfig) *SecurityAuditor {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &SecurityAuditor{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 用于测试
				},
			},
		},
	}
}

// RunAudit 运行安全审计
func (sa *SecurityAuditor) RunAudit(ctx context.Context) *AuditReport {
	report := &AuditReport{
		StartTime: time.Now(),
		Findings:  make([]AuditFinding, 0),
	}

	// SSL/TLS检查
	if sa.config.CheckSSL {
		report.SSLCheck = sa.checkSSL(ctx)
	}

	// HTTP头检查
	if sa.config.CheckHeaders {
		report.HeaderCheck = sa.checkHeaders(ctx)
	}

	// 端点检查
	if sa.config.CheckEndpoints {
		report.EndpointCheck = sa.checkEndpoints(ctx)
	}

	// 文件检查
	if sa.config.CheckFiles {
		report.FileCheck = sa.checkFiles(ctx)
	}

	// 计算摘要
	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.Summary = sa.calculateSummary(report)

	return report
}

// checkSSL 检查SSL配置
func (sa *SecurityAuditor) checkSSL(ctx context.Context) *SSLCheckResult {
	result := &SSLCheckResult{}

	if sa.config.TargetURL == "" {
		return result
	}

	// 尝试建立TLS连接
	dialer := &tls.Dialer{
		Config: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	host := strings.TrimPrefix(sa.config.TargetURL, "https://")
	host = strings.TrimPrefix(host, "http://")

	conn, err := dialer.DialContext(ctx, "tcp", host+":443")
	if err != nil {
		return result
	}
	defer conn.Close()

	tlsConn := conn.(*tls.Conn)
	state := tlsConn.ConnectionState()

	result.TLSVersion = tlsVersionName(state.Version)
	result.CipherSuite = tlsCipherName(state.CipherSuite)

	// 检查证书
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		result.CertificateValid = true
		result.Issuer = cert.Issuer.CommonName
		result.SelfSigned = cert.Issuer.CommonName == cert.Subject.CommonName

		expiresIn := cert.NotAfter.Sub(time.Now())
		result.ExpiresIn = expiresIn.String()

		if expiresIn < 30*24*time.Hour {
			result.CertificateValid = false
		}
	}

	// 检查弱协议
	if state.Version < tls.VersionTLS12 {
		result.WeakProtocol = true
	}

	// 检查弱加密套件
	weakCiphers := []uint16{
		tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	}
	for _, wc := range weakCiphers {
		if state.CipherSuite == wc {
			result.WeakCipher = true
			break
		}
	}

	return result
}

// checkHeaders 检查HTTP安全头
func (sa *SecurityAuditor) checkHeaders(ctx context.Context) *HeaderCheckResult {
	result := &HeaderCheckResult{
		SecurityHeaders: make(map[string]HeaderStatus),
		MissingHeaders:  make([]string, 0),
		WeakHeaders:     make(map[string]string),
	}

	if sa.config.TargetURL == "" {
		return result
	}

	req, err := http.NewRequestWithContext(ctx, "GET", sa.config.TargetURL, nil)
	if err != nil {
		return result
	}

	resp, err := sa.client.Do(req)
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	// 安全头列表
	securityHeaders := map[string]struct {
		required bool
		secureValue string
	}{
		"X-Content-Type-Options":    {true, "nosniff"},
		"X-Frame-Options":           {true, "DENY"},
		"X-XSS-Protection":          {true, "1; mode=block"},
		"Strict-Transport-Security": {true, "max-age=31536000"},
		"Content-Security-Policy":   {false, ""},
		"Referrer-Policy":           {false, "strict-origin-when-cross-origin"},
		"Permissions-Policy":        {false, ""},
	}

	for header, config := range securityHeaders {
		value := resp.Header.Get(header)
		status := HeaderStatus{
			Present: value != "",
			Value:   value,
		}

		if value != "" {
			if config.secureValue != "" && value != config.secureValue {
				status.Secure = false
				result.WeakHeaders[header] = value
			} else {
				status.Secure = true
			}
		} else if config.required {
			result.MissingHeaders = append(result.MissingHeaders, header)
		}

		result.SecurityHeaders[header] = status
	}

	return result
}

// checkEndpoints 检查端点安全性
func (sa *SecurityAuditor) checkEndpoints(ctx context.Context) *EndpointCheckResult {
	result := &EndpointCheckResult{
		Endpoints: make([]EndpointStatus, 0),
		Exposed:   make([]string, 0),
		Protected: make([]string, 0),
	}

	// 常见敏感端点
	sensitiveEndpoints := []struct {
		path   string
		method string
	}{
		{"/admin", "GET"},
		{"/debug", "GET"},
		{"/metrics", "GET"},
		{"/api/v1/users", "GET"},
		{"/api/v1/config", "GET"},
		{"/.env", "GET"},
		{"/config.yaml", "GET"},
		{"/api/v1/tasks", "POST"},
	}

	for _, ep := range sensitiveEndpoints {
		status := EndpointStatus{
			Path:   ep.path,
			Method: ep.method,
		}

		var req *http.Request
		var err error

		url := sa.config.TargetURL + ep.path
		if sa.config.TargetURL == "" {
			status.Accessible = false
			status.Protected = true
			result.Endpoints = append(result.Endpoints, status)
			continue
		}

		req, err = http.NewRequestWithContext(ctx, ep.method, url, nil)
		if err != nil {
			status.Accessible = false
			result.Endpoints = append(result.Endpoints, status)
			continue
		}

		resp, err := sa.client.Do(req)
		if err != nil {
			status.Accessible = false
			result.Endpoints = append(result.Endpoints, status)
			continue
		}
		resp.Body.Close()

		status.Status = resp.StatusCode
		status.Accessible = resp.StatusCode != 404

		// 检查是否需要认证
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			status.Protected = true
			result.Protected = append(result.Protected, ep.path)
		} else if resp.StatusCode == 200 {
			status.Protected = false
			result.Exposed = append(result.Exposed, ep.path)
		}

		result.Endpoints = append(result.Endpoints, status)
	}

	return result
}

// checkFiles 检查敏感文件
func (sa *SecurityAuditor) checkFiles(ctx context.Context) *FileCheckResult {
	result := &FileCheckResult{
		SensitiveFiles:   make([]SensitiveFile, 0),
		PermissionIssues: make([]PermissionIssue, 0),
	}

	// 敏感文件模式
	sensitivePatterns := []struct {
		pattern string
		fileType string
		risk     string
	}{
		{".env", "environment", "high"},
		{"*.pem", "certificate", "high"},
		{"*.key", "private_key", "critical"},
		{"*.p12", "keystore", "high"},
		{"id_rsa", "ssh_key", "critical"},
		{"credentials.json", "credentials", "critical"},
		{"secrets.yaml", "secrets", "critical"},
		{".git/config", "git_config", "medium"},
	}

	for _, sp := range sensitivePatterns {
		matches, _ := filepath.Glob(sp.pattern)
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}

			sf := SensitiveFile{
				Path:       match,
				Type:       sp.fileType,
				Risk:       sp.risk,
				Accessible: info.Mode().Perm()&0044 != 0,
			}
			result.SensitiveFiles = append(result.SensitiveFiles, sf)
		}
	}

	return result
}

// calculateSummary 计算摘要
func (sa *SecurityAuditor) calculateSummary(report *AuditReport) *AuditSummary {
	summary := &AuditSummary{}

	// SSL检查
	if report.SSLCheck != nil {
		summary.TotalChecks += 5
		if report.SSLCheck.CertificateValid {
			summary.PassedChecks++
		} else {
			summary.WarningCount++
		}
		if report.SSLCheck.TLSVersion != "" && !report.SSLCheck.WeakProtocol {
			summary.PassedChecks++
		} else {
			summary.CriticalCount++
		}
		if !report.SSLCheck.WeakCipher {
			summary.PassedChecks++
		} else {
			summary.CriticalCount++
		}
		if !report.SSLCheck.SelfSigned {
			summary.PassedChecks++
		} else {
			summary.WarningCount++
		}
		summary.PassedChecks++ // 完成SSL检查
	}

	// 头检查
	if report.HeaderCheck != nil {
		summary.TotalChecks += len(report.HeaderCheck.SecurityHeaders)
		summary.PassedChecks += len(report.HeaderCheck.SecurityHeaders) - len(report.HeaderCheck.MissingHeaders) - len(report.HeaderCheck.WeakHeaders)
		summary.WarningCount += len(report.HeaderCheck.WeakHeaders)
		summary.CriticalCount += len(report.HeaderCheck.MissingHeaders)
	}

	// 端点检查
	if report.EndpointCheck != nil {
		summary.TotalChecks += len(report.EndpointCheck.Endpoints)
		summary.PassedChecks += len(report.EndpointCheck.Protected)
		summary.CriticalCount += len(report.EndpointCheck.Exposed)
	}

	// 文件检查
	if report.FileCheck != nil {
		for _, sf := range report.FileCheck.SensitiveFiles {
			summary.TotalChecks++
			if !sf.Accessible {
				summary.PassedChecks++
			} else if sf.Risk == "critical" {
				summary.CriticalCount++
			} else {
				summary.WarningCount++
			}
		}
	}

	// 计算安全评分
	if summary.TotalChecks > 0 {
		summary.SecurityScore = int(float64(summary.PassedChecks) / float64(summary.TotalChecks) * 100)
		summary.SecurityScore -= summary.CriticalCount * 10
		summary.SecurityScore -= summary.WarningCount * 5
		if summary.SecurityScore < 0 {
			summary.SecurityScore = 0
		}
		if summary.SecurityScore > 100 {
			summary.SecurityScore = 100
		}
	}

	return summary
}

// ExportReport 导出报告
func (r *AuditReport) ExportReport() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// PrintReport 打印报告
func (r *AuditReport) PrintReport() string {
	var sb string
	sb = fmt.Sprintf("\n=== OFA 安全审计报告 ===\n")
	sb += fmt.Sprintf("审计时间: %s\n", r.StartTime.Format(time.RFC3339))
	sb += fmt.Sprintf("耗时: %v\n\n", r.Duration)

	sb += fmt.Sprintf("安全评分: %d/100\n\n", r.Summary.SecurityScore)

	sb += fmt.Sprintf("检查统计:\n")
	sb += fmt.Sprintf("  总检查项: %d\n", r.Summary.TotalChecks)
	sb += fmt.Sprintf("  通过: %d\n", r.Summary.PassedChecks)
	sb += fmt.Sprintf("  警告: %d\n", r.Summary.WarningCount)
	sb += fmt.Sprintf("  严重: %d\n", r.Summary.CriticalCount)

	if r.SSLCheck != nil {
		sb += fmt.Sprintf("\nSSL/TLS检查:\n")
		sb += fmt.Sprintf("  TLS版本: %s\n", r.SSLCheck.TLSVersion)
		sb += fmt.Sprintf("  加密套件: %s\n", r.SSLCheck.CipherSuite)
		sb += fmt.Sprintf("  证书有效: %v\n", r.SSLCheck.CertificateValid)
	}

	if r.HeaderCheck != nil {
		sb += fmt.Sprintf("\nHTTP头检查:\n")
		sb += fmt.Sprintf("  缺失头: %v\n", r.HeaderCheck.MissingHeaders)
		sb += fmt.Sprintf("  弱配置: %v\n", r.HeaderCheck.WeakHeaders)
	}

	if r.EndpointCheck != nil {
		sb += fmt.Sprintf("\n端点检查:\n")
		sb += fmt.Sprintf("  暴露端点: %v\n", r.EndpointCheck.Exposed)
		sb += fmt.Sprintf("  受保护端点: %v\n", r.EndpointCheck.Protected)
	}

	return sb
}

// tlsVersionName TLS版本名称
func tlsVersionName(version uint16) string {
	names := map[uint16]string{
		tls.VersionTLS10: "TLS 1.0",
		tls.VersionTLS11: "TLS 1.1",
		tls.VersionTLS12: "TLS 1.2",
		tls.VersionTLS13: "TLS 1.3",
	}
	if name, ok := names[version]; ok {
		return name
	}
	return "Unknown"
}

// tlsCipherName 加密套件名称
func tlsCipherName(cipher uint16) string {
	// 简化返回
	return fmt.Sprintf("0x%04x", cipher)
}

// 敏感内容检测
var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)password\s*=\s*[^\s]+`),
	regexp.MustCompile(`(?i)api_key\s*=\s*[^\s]+`),
	regexp.MustCompile(`(?i)secret\s*=\s*[^\s]+`),
	regexp.MustCompile(`(?i)token\s*=\s*[^\s]+`),
	regexp.MustCompile(`-----BEGIN.*PRIVATE KEY-----`),
}

// DetectSensitiveContent 检测敏感内容
func DetectSensitiveContent(content string) []string {
	findings := make([]string, 0)
	for _, pattern := range sensitivePatterns {
		if pattern.MatchString(content) {
			findings = append(findings, pattern.String())
		}
	}
	return findings
}