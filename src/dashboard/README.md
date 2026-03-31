# OFA Web Dashboard

OFA 管理控制台，基于 Vue 3 + TypeScript + Vite 构建。

## 功能特性

- **控制台首页**: 系统概览、渐变统计卡片、实时活动流、性能指标
- **智能体管理**: 卡片网格布局、搜索过滤、详情弹窗、删除操作
- **任务管理**: 状态图标、新建表单、状态筛选、统计条
- **系统监控**: 指标卡片、WebSocket 实时更新、任务进度条
- **消息中心**: 消息发送面板、广播消息、发送历史、快捷操作
- **系统设置**: 连接配置、显示设置、数据设置

## 快速开始

```bash
# 进入 Dashboard 目录
cd src/dashboard

# 安装依赖
npm install

# 启动开发服务器 (端口 3000)
npm run dev

# 访问 http://localhost:3000
```

**注意**: 需要 Center 服务运行在端口 8080 以提供 API 支持。

## 构建

```bash
# 构建生产版本
npm run build

# 预览构建结果
npm run preview
```

## 项目结构

```
src/
├── api/
│   └── client.ts       # REST + WebSocket API 客户端
├── components/
│   ├── Layout.vue      # 布局组件
│   └── Sidebar.vue     # 侧边导航栏
├── types/
│   └── index.ts        # TypeScript 类型定义
├── views/
│   ├── Dashboard.vue   # 控制台首页
│   ├── Agents.vue      # 智能体管理
│   ├── Tasks.vue       # 任务管理
│   ├── Monitor.vue     # 系统监控
│   ├── Messages.vue    # 消息中心
│   └── Settings.vue    # 系统设置
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

| 技术 | 版本 | 用途 |
|------|------|------|
| Vue | 3.4+ | 前端框架 |
| TypeScript | 5.3+ | 类型安全 |
| Vite | 5.0+ | 构建工具 |
| Vue Router | 4.2+ | 路由管理 |
| Chart.js | 4.4+ | 图表库 |

## 特性

- 🎨 **渐变色卡片** - 美观的统计卡片设计
- 🔄 **WebSocket** - 实时数据更新
- 📱 **响应式布局** - 适配不同屏幕尺寸
- 🌐 **中文界面** - 全中文 UI
- ⚡ **Vite HMR** - 快速热更新开发体验