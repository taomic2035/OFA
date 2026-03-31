# 安全策略

## 支持的版本

| 版本 | 支持状态 |
|------|---------|
| v9.0.x | ✅ 支持 |
| v8.0.x | ✅ 支持 |
| < v8.0 | ❌ 不再支持 |

## 报告安全漏洞

如果您发现安全漏洞，请**不要**通过公开的 Issue 报告。

请通过以下方式私下报告：

- Email: 279316081@qq.com
- 主题: [SECURITY] OFA 安全漏洞报告

请在报告中包含：

1. 漏洞描述
2. 影响范围
3. 复现步骤
4. 可能的修复方案（如有）

我们承诺：

- 在 48 小时内确认收到报告
- 在 7 天内提供初步评估
- 修复后将公开致谢

## 安全最佳实践

### 部署安全

1. **启用 TLS**
   ```yaml
   server:
     tls:
       enabled: true
       cert_file: /path/to/cert.pem
       key_file: /path/to/key.pem
   ```

2. **配置认证**
   ```yaml
   auth:
     jwt:
       secret: ${JWT_SECRET}
       expiry: 24h
   ```

3. **限制网络访问**
   - 使用防火墙限制端口访问
   - 仅允许可信 IP 访问 gRPC 端口

### Agent 安全

1. **验证 Agent 身份**
2. **使用最小权限原则**
3. **定期轮换密钥**

### 数据安全

1. **敏感数据加密**
2. **审计日志**
3. **定期备份**

## 已知安全问题

查看 [Security Advisories](https://github.com/taomic2035/OFA/security/advisories)