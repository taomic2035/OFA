# 贡献指南

感谢您考虑为 OFA 项目做出贡献！

## 如何贡献

### 报告问题

如果您发现了 bug 或有功能建议，请：

1. 在 [Issues](https://github.com/taomic2035/OFA/issues) 中搜索是否已有相关问题
2. 如果没有，创建新的 Issue，包含：
   - 问题描述
   - 复现步骤
   - 预期行为
   - 实际行为
   - 环境信息（操作系统、Go版本等）

### 提交代码

1. **Fork 仓库**

2. **克隆您的 Fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/OFA.git
   cd OFA
   ```

3. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **编写代码**
   - 遵循 Go 代码规范
   - 添加单元测试
   - 更新相关文档

5. **运行测试**
   ```bash
   make test
   make lint
   ```

6. **提交更改**
   ```bash
   git add .
   git commit -m "feat: add some feature"
   ```

7. **推送分支**
   ```bash
   git push origin feature/your-feature-name
   ```

8. **创建 Pull Request**

## 代码规范

### Go 代码

- 使用 `gofmt` 格式化代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go) 指南
- 添加必要的注释

### 提交信息格式

```
<type>: <subject>

<body>
```

类型：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

示例：
```
feat: add LLM integration support

- Add OpenAI adapter
- Add Claude adapter
- Add local model adapter
```

## 开发环境设置

```bash
# 安装依赖
go mod download

# 生成 Proto 文件
make proto

# 运行开发服务器
make dev
```

## 许可证

通过贡献代码，您同意您的代码将以 MIT 许可证授权。