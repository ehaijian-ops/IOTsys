# 开发者指南

> IOTsys 物联网充电桩管理平台 · 开发者贡献指南 v1.0

---

## 一、项目概述

IOTsys 是一套基于 Go + Vue 3 的物联网充电桩管理平台，核心特点：

- **多协议适配**：插件式协议适配器架构，支持 AP3000 和 TF100 协议
- **高并发接入**：原生 TCP Server 支持 5 万+ 设备并发连接
- **实时数据处理**：Redis 缓存 + Kafka 消息队列 + MongoDB 时序存储
- **前后端分离**：Go (Gin) + Vue 3 (Element Plus)

---

## 二、技术栈

### 后端

| 技术         | 版本   | 用途               |
| ------------ | ------ | ------------------ |
| Go           | 1.22   | 开发语言            |
| Gin          | 1.10   | HTTP Web 框架       |
| GORM         | 1.25   | ORM                |
| go-redis     | 8.11   | Redis 客户端        |
| kafka-go     | 0.4    | Kafka 客户端        |
| mongo-driver | 1.16   | MongoDB 客户端      |
| Viper        | 1.19   | 配置管理            |
| Zap          | 1.27   | 结构化日志          |
| JWT          | 5.2    | 认证鉴权            |
| UUID         | 1.6    | 唯一 ID 生成        |

### 前端

| 技术           | 版本 | 用途            |
| -------------- | ---- | --------------- |
| Vue            | 3.4  | 前端框架         |
| TypeScript     | 5.4  | 类型安全         |
| Vite           | 5.2  | 构建工具         |
| Element Plus   | 2.7  | UI 组件库        |
| Pinia          | 2.1  | 状态管理         |
| Vue Router     | 4.3  | 路由管理         |
| ECharts        | 5.5  | 图表可视化        |
| Axios          | 1.7  | HTTP 请求        |

---

## 三、本地开发环境搭建

### 3.1 前置依赖

- Go 1.22+
- Node.js 20+ & pnpm
- Docker & Docker Compose（用于启动基础设施）

### 3.2 启动基础设施

```bash
# 在项目根目录执行
docker compose -f deploy/docker-compose/docker-compose.yml up -d

# 这会启动 MySQL、MongoDB、Redis、Kafka、Zookeeper
# iot-server 服务使用 build 配置，开发时可先不启动
```

### 3.3 配置后端

```bash
cd server

# 复制配置模板
cp config/config.example.yaml config/config.yaml

# 编辑配置（本地开发通常只需确认端口不冲突）
# 关键配置项：
#   mysql.port: 3306
#   redis.addr: localhost:6379
#   kafka.brokers: localhost:9092
#   tcp.port: 7000
#   server.port: 8080
```

### 3.4 启动后端

```bash
cd server

# 安装依赖
go mod tidy

# 启动服务
go run cmd/main.go

# 预期输出：
# INFO  Starting IoT Platform...  {"env": "development", "port": 8080}
# INFO  TCP server started  {"addr": "0.0.0.0:7000", "max_connections": 50000}
# INFO  HTTP server started  {"port": 8080}
```

### 3.5 启动前端

```bash
cd web/admin

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 访问 http://localhost:3000
```

### 3.6 验证

```bash
# 后端健康检查
curl http://localhost:8080/health

# 前端可访问登录页面
# 打开浏览器访问 http://localhost:3000
```

---

## 四、项目目录结构

```
IOTsys/
├── ap3000_protocol.txt        # AP3000 原始协议文档
├── tf100_protocol.txt         # TF100 原始协议文档
├── README.md                  # 项目说明
├── docs/                      # 文档
│   ├── PRD-物联网充电桩管理平台.md  # 产品需求文档
│   ├── ARCHITECTURE.md             # 架构设计文档
│   ├── DATABASE.md                 # 数据库设计文档
│   ├── DEPLOYMENT.md               # 部署运维手册
│   ├── USER_MANUAL.md              # 用户操作手册
│   └── CONTRIBUTING.md             # 本文件
├── deploy/
│   └── docker-compose/
│       └── docker-compose.yml      # 基础设施编排
├── server/                    # 后端 Go 项目
│   ├── cmd/main.go            # 应用入口
│   ├── config/
│   │   └── config.example.yaml    # 配置模板
│   ├── internal/
│   │   ├── connector/
│   │   │   └── tcpserver/         # TCP 设备接入服务器
│   │   ├── protocol/
│   │   │   ├── engine/            # 协议适配器引擎
│   │   │   ├── model/             # 标准数据模型
│   │   │   └── adapters/
│   │   │       ├── ap3000/        # AP3000 协议适配器
│   │   │       └── tf100/         # TF100 协议适配器
│   │   ├── device/                # 设备管理模块
│   │   │   ├── model/             # 设备数据模型
│   │   │   ├── repository/        # 数据访问层
│   │   │   ├── service/           # 业务逻辑层
│   │   │   └── handler/           # HTTP 处理器
│   │   └── command/               # 指令管理模块
│   │       ├── service/           # 指令业务逻辑
│   │       └── handler/           # 指令 HTTP 处理器
│   ├── pkg/                       # 公共包
│   │   ├── config/                # 配置管理
│   │   ├── database/
│   │   │   ├── mysql/             # MySQL 连接
│   │   │   ├── redis/             # Redis 连接
│   │   │   └── mongodb/           # MongoDB 连接
│   │   ├── mq/kafka/              # Kafka 生产/消费
│   │   ├── auth/                  # JWT 认证
│   │   ├── middleware/            # HTTP 中间件
│   │   ├── errors/                # 错误定义
│   │   └── logger/                # 日志封装
│   ├── migrations/                # 数据库迁移
│   ├── go.mod
│   └── Dockerfile
└── web/                       # 前端项目
    └── admin/
        ├── index.html
        ├── package.json
        ├── vite.config.ts
        └── src/
            ├── App.vue
            ├── main.ts
            ├── api/               # API 层
            ├── layout/            # 布局组件
            ├── router/            # 路由配置
            ├── styles/            # 样式
            └── views/             # 页面组件
```

---

## 五、核心架构概念

### 5.1 协议适配器模式

平台通过 `ProtocolAdapter` 接口实现多协议支持：

```go
type ProtocolAdapter interface {
    Name() string
    Version() string
    DeviceType() string
    Detect(raw []byte) bool
    Decode(raw []byte) (*model.StandardData, error)
    Encode(cmd *model.StandardCommand) ([]byte, error)
    DecodeResponse(raw []byte, cmdID string) (*model.StandardCommandResponse, error)
    Validate(raw []byte) error
}
```

每个协议适配器通过 `init()` 自动注册到引擎：

```go
func init() {
    engine.Register(&AP3000Adapter{})
}
```

### 5.2 数据流转

```
设备 ──TCP──> tcpserver ──Detect──> ProtocolAdapter.Decode()
                                        │
                                   StandardData
                                        │
                              ┌─────────┼─────────┐
                              │         │         │
                           Redis      Kafka   WebSocket
                         (实时缓存)  (消息队列) (实时推送)
                              │         │
                              │    MongoDB/MySQL
                              │    (持久存储)
```

### 5.3 指令下发流程

```
前端/API ──> CommandService.CreateCommand()
                │
                ├── 1. 检查设备在线状态
                ├── 2. 获取设备协议类型
                ├── 3. 创建指令记录 (MySQL)
                ├── 4. ProtocolAdapter.Encode() → 协议帧
                ├── 5. tcpserver.SendCommand() → 设备
                └── 6. 更新指令状态
```

---

## 六、开发指南

### 6.1 新增协议适配器

1. 在 `server/internal/protocol/adapters/` 下创建新目录
2. 实现 `ProtocolAdapter` 接口的 7 个方法
3. 在 `init()` 中调用 `engine.Register()`
4. 在 `cmd/main.go` 中匿名导入新适配器包

参考：`ap3000/ap3000.go` 和 `tf100/tf100.go`

### 6.2 新增业务模块

以"告警模块"为例：

```
1. 创建 internal/alert/
   ├── model/alert.go          # GORM 模型
   ├── repository/alert_repo.go # 数据访问层
   ├── service/alert_service.go # 业务逻辑
   └── handler/alert_handler.go # HTTP 处理器

2. 在 migrations/migrate.go 添加 AutoMigrate

3. 在 cmd/main.go 中初始化并注册路由
```

### 6.3 新增 API 端点

```go
// 在 handler 中添加方法
func (h *DeviceHandler) GetStatistics(c *gin.Context) {
    // ...业务逻辑...
    c.JSON(200, gin.H{"data": result})
}

// 在 cmd/main.go 中注册路由
devices.GET("/statistics", devH.GetStatistics)
```

### 6.4 前端新增页面

```
1. 在 src/views/ 下创建组件文件
2. 在 src/router/index.ts 添加路由
3. 在 src/api/ 下添加 API 接口定义
4. 在侧边栏 layout/index.vue 添加菜单项
```

---

## 七、代码规范

### 7.1 Go 代码规范

- 遵循 [Effective Go](https://go.dev/doc/effective_go) 和 [Uber Go Style Guide](https://github.com/uber-go/guide)
- 使用 `gofmt` 和 `goimports` 格式化代码
- 包名使用小写单词，不使用下划线
- 导出函数/类型添加注释（格式：`// FunctionName does something`）
- 错误处理：优先使用 `fmt.Errorf("context: %w", err)` 包装错误

```bash
# 格式化
go fmt ./...
goimports -w ./...

# 静态检查
go vet ./...
golangci-lint run ./...
```

### 7.2 TypeScript 代码规范

- 使用 ESLint + Prettier
- 组件命名：PascalCase（如 `DeviceList.vue`）
- 文件命名：kebab-case（如 `device-service.ts`）
- API 接口使用 `interface` 定义类型

```bash
cd web/admin
pnpm lint
pnpm format
```

### 7.3 Git 提交规范

```
<type>(<scope>): <subject>

类型：
  feat     新功能
  fix      修复 Bug
  docs     文档变更
  refactor 重构
  test     测试
  chore    构建/工具

示例：
  feat(protocol): add OCPP 1.6 adapter
  fix(tcp): handle connection timeout edge case
  docs(api): add device endpoint documentation
```

---

## 八、测试策略

### 8.1 单元测试

```bash
# 运行所有测试
go test ./...

# 运行指定包的测试
go test ./internal/protocol/adapters/ap3000/

# 带覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 8.2 测试重点

| 模块           | 测试重点                           |
| -------------- | ---------------------------------- |
| 协议适配器     | 帧解析正确性、边界条件、异常数据    |
| TCP Server     | 连接管理、心跳检测、断线重连        |
| 设备管理       | CRUD 操作、分页查询、权限校验       |
| 指令管理       | 指令下发流程、状态流转、超时处理    |
| 认证鉴权       | Token 生成/验证、角色权限、过期处理 |

### 8.3 协议测试数据

在 `server/internal/protocol/adapters/ap3000/testdata/` 目录下存放：
- 有效帧样本
- 异常帧样本（长度错误、校验失败、非法命令码）
- 边界帧样本（空数据、最大长度）

---

## 九、调试技巧

### 9.1 后端调试

```bash
# 启用详细日志
# 修改 config.yaml:
#   log.level: debug

# 使用 pprof 分析性能
go tool pprof http://localhost:8080/debug/pprof/profile

# 查看 goroutine
curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

### 9.2 模拟设备连接

```bash
# 模拟 AP3000 设备（发送登录帧）
echo -ne '\x44\x4E\x59...' | nc localhost 7000

# 模拟 TF100 设备（发送注册消息）
echo 'CCMD:42{"MsgType":20,"DevID":"TEST001","MsgID":1}' | nc localhost 7000
```

### 9.3 前端调试

```bash
# 启用 Vue DevTools
# 浏览器安装 Vue DevTools 扩展

# API 代理
# vite.config.ts 中已配置 /api/v1 → localhost:8080
```

---

## 十、常见开发任务

### 10.1 修改配置项

1. 在 `pkg/config/config.go` 添加配置结构体
2. 在 `config/config.example.yaml` 添加对应配置
3. 在 `cmd/main.go` 中使用配置初始化服务

### 10.2 添加数据库迁移

```go
// migrations/migrate.go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &model.Device{},
        &model.Site{},
        &cmdService.DeviceCommand{},
        &alert.Alert{},       // 新增
        &user.User{},         // 新增
    )
}
```

### 10.3 添加中间件

```go
// pkg/middleware/ratelimit.go
func RateLimit(limit int) gin.HandlerFunc {
    // ...实现...
}

// cmd/main.go
router.Use(middleware.RateLimit(100))
```

---

## 十一、相关资源

- **协议文档**：`ap3000_protocol.txt` / `tf100_protocol.txt`
- **产品需求**：`docs/PRD-物联网充电桩管理平台.md`
- **架构设计**：`docs/ARCHITECTURE.md`
- **数据库设计**：`docs/DATABASE.md`
- **部署运维**：`docs/DEPLOYMENT.md`
