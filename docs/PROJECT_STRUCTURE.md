# OFA 项目目录结构

```
OFA/
├── docs/                           # 文档目录
│   ├── 01-FEATURE_SPECIFICATION.md # 功能规格说明书
│   ├── 02-TECH_SELECTION.md        # 技术选型分析
│   ├── 03-ARCHITECTURE_DESIGN.md   # 架构设计文档
│   ├── 04-SOLUTION_DESIGN.md       # 方案设计文档
│   ├── 05-TEST_CASES.md            # 测试用例
│   ├── 06-VERSION_PLAN.md          # 版本迭代计划
│   ├── api/                        # API文档
│   └── diagrams/                   # 架构图、流程图
│
├── src/                            # 源代码目录
│   ├── center/                     # Center服务端
│   │   ├── core/                   # 核心模块
│   │   ├── api/                    # API接口
│   │   ├── scheduler/              # 任务调度
│   │   ├── storage/                # 数据存储
│   │   └── config/                 # 配置管理
│   │
│   ├── agent/                      # Agent客户端
│   │   ├── core/                   # 核心模块
│   │   ├── communication/          # 通信模块
│   │   ├── executor/               # 任务执行
│   │   ├── skills/                 # 技能模块
│   │   └── tools/                  # 工具模块
│   │
│   ├── sdk/                        # SDK
│   │   ├── java/                   # Java SDK
│   │   ├── kotlin/                 # Kotlin SDK
│   │   ├── python/                 # Python SDK
│   │   ├── javascript/             # JavaScript SDK
│   │   └── c/                      # C/C++ SDK
│   │
│   ├── protocol/                   # 协议定义
│   │   ├── proto/                  # Protobuf定义
│   │   ├── json/                   # JSON Schema
│   │   └── openapi/                # OpenAPI规范
│   │
│   └── common/                     # 公共代码
│       ├── utils/                  # 工具类
│       ├── crypto/                 # 加密模块
│       └── models/                 # 数据模型
│
├── platforms/                      # 平台适配
│   ├── android/                    # Android平台
│   ├── ios/                        # iOS平台
│   ├── desktop/                    # 桌面平台(Windows/macOS/Linux)
│   ├── web/                        # Web平台
│   └── embedded/                   # 嵌入式平台
│
├── skills/                         # 技能库
│   ├── official/                   # 官方技能
│   └── community/                  # 社区技能
│
├── tools/                          # 工具库
│   ├── native/                     # 原生工具
│   ├── scripts/                    # 脚本工具
│   └── models/                     # 模型工具
│
├── tests/                          # 测试目录
│   ├── unit/                       # 单元测试
│   ├── integration/                # 集成测试
│   ├── e2e/                        # 端到端测试
│   ├── performance/                # 性能测试
│   └── mock/                       # Mock数据
│
├── build/                          # 构建输出
│   ├── center/                     # Center构建产物
│   ├── agent/                      # Agent构建产物
│   └── sdk/                        # SDK构建产物
│
├── deployments/                    # 部署配置
│   ├── docker/                     # Docker配置
│   ├── k8s/                        # Kubernetes配置
│   └── scripts/                    # 部署脚本
│
├── configs/                        # 配置文件
│   ├── dev/                        # 开发环境配置
│   ├── test/                       # 测试环境配置
│   └── prod/                       # 生产环境配置
│
├── scripts/                        # 脚本工具
│   ├── build.sh                    # 构建脚本
│   ├── test.sh                     # 测试脚本
│   └── deploy.sh                   # 部署脚本
│
├── docs-output/                    # 文档生成输出
│
├── README.md                       # 项目说明
├── LICENSE                         # 许可证
├── CHANGELOG.md                    # 变更日志
└── Makefile                        # 构建入口
```

## 目录说明

### docs/ - 文档
所有项目文档集中存放，按序号组织便于查阅。

### src/ - 源代码
- **center/**: 服务端代码，负责Agent管理、任务调度
- **agent/**: 客户端代码，运行在各设备上
- **sdk/**: 各语言SDK，简化接入
- **protocol/**: 通信协议定义
- **common/**: 跨平台公共代码

### platforms/ - 平台适配
各平台的特定实现和打包配置。

### skills/ & tools/ - 能力扩展
- **skills/**: 声明式能力描述
- **tools/**: 具体实现代码

### tests/ - 测试
按测试类型组织，支持自动化测试流程。

### build/ - 构建产物
构建输出目录，不纳入版本控制。

### deployments/ - 部署
容器化和部署相关配置。

### configs/ - 配置
多环境配置管理。