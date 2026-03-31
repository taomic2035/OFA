# OFA Web Dashboard

OFA 管理控制台，基于 Vue 3 + TypeScript + Vite 构建。

## 功能特性

- **Dashboard 首页**: 系统概览、实时统计、Agent/Task 摘要
- **Agent 管理**: 查看、搜索、删除 Agent
- **Task 管理**: 提交、查看、取消 Task
- **监控面板**: 实时指标、性能数据
- **消息中心**: Agent 间消息发送、广播

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 预览构建结果
npm run preview
```

## 项目结构

```
src/
├── api/
│   └── client.ts       # API 客户端
├── components/
│   ├── Layout.vue      # 布局组件
│   └── Sidebar.vue     # 侧边栏
├── types/
│   └── index.ts        # TypeScript 类型定义
├── views/
│   ├── Dashboard.vue   # 首页
│   ├── Agents.vue      # Agent 管理
│   ├── Tasks.vue       # Task 管理
│   ├── Monitor.vue     # 监控面板
│   └── Messages.vue    # 消息中心
├── styles/
│   └── main.css        # 全局样式
├── App.vue             # 根组件
└── main.ts             # 入口文件
```

## API 端点

| 功能 | 方法 | 端点 |
|------|------|------|
| Agent 列表 | GET | /api/v1/agents |
| Agent 详情 | GET | /api/v1/agents/{id} |
| 删除 Agent | DELETE | /api/v1/agents/{id} |
| Task 列表 | GET | /api/v1/tasks |
| 提交 Task | POST | /api/v1/tasks |
| Task 详情 | GET | /api/v1/tasks/{id} |
| 取消 Task | POST | /api/v1/tasks/{id}/cancel |
| 系统信息 | GET | /api/v1/system/info |
| 系统指标 | GET | /api/v1/system/metrics |
| 技能列表 | GET | /api/v1/skills |
| 发送消息 | POST | /api/v1/messages |
| 广播消息 | POST | /api/v1/messages/broadcast |

## 部署

Dashboard 作为 Center 服务的一部分部署：

1. 构建前端：
   ```bash
   npm run build
   ```

2. 将 `dist` 目录复制到 Center 运行目录的 `dashboard/` 子目录

3. 启动 Center 服务后访问 `http://localhost:8080/dashboard`

## 技术栈

- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全
- **Vite** - 下一代前端构建工具
- **Vue Router** - 官方路由
- **Chart.js** - 图表库（可选）