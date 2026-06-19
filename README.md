# IoT 充电桩管理平台

## 项目简介

一套面向物联网充电桩（电单车/电动汽车）的统一数据采集与指令下发管理平台。基于 Go 语言构建，支持高并发设备接入（5万+设备），采用插件化协议适配架构，可灵活扩展支持不同厂商设备。

## 核心特性

- **多协议支持**: 插件化协议适配引擎，已内置 AP3000（电单车）和 TF100（汽车）协议
- **高并发接入**: 基于 evio 的 TCP 服务器，支持 5 万+ 设备长连接
- **实时数据**: Redis 缓存设备状态 + WebSocket 实时推送
- **指令下发**: 异步指令下发，状态追踪，支持远程启停/参数配置/OTA
- **告警管理**: 设备异常实时告警，支持规则配置
- **扩展机制**: 新增协议适配器只需实现接口 + init 注册，3 天完成

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端语言 | Go 1.22 |
| Web 框架 | Gin |
| TCP 服务器 | evio |
| 数据库 | MySQL 8.0 + MongoDB 7.0 |
| 缓存 | Redis 7.x |
| 消息队列 | Kafka |
| 前端 | Vue 3 + TypeScript + Element Plus |
| 容器化 | Docker + Docker Compose |

## 文档导航

| 文档 | 说明 | 适用对象 |
|------|------|---------|
| [产品需求文档](docs/PRD-物联网充电桩管理平台.md) | 产品背景、功能规划、里程碑排期 | 产品/项目管理者 |
| [架构设计文档](docs/ARCHITECTURE.md) | 技术选型、架构图、协议规范、数据流转 | 架构师/后端开发 |
| [数据库设计文档](docs/DATABASE.md) | ER图、表结构、Redis/MongoDB/Kafka设计 | 后端开发/DBA |
| [API 接口文档](docs/API.md) | 完整接口列表、请求响应格式、错误码 | 前端/后端/测试 |
| [部署运维手册](docs/DEPLOYMENT.md) | Docker/K8s部署、监控、备份、故障处理 | 运维工程师 |
| [用户操作手册](docs/USER_MANUAL.md) | 平台使用指南、常见操作场景 | 管理员/运维人员 |
| [开发者指南](docs/CONTRIBUTING.md) | 本地环境搭建、代码规范、测试策略 | 新加入的开发者 |

## 项目结构

```
e:/IOTsys/
├── docs/                    # 文档
│   ├── PRD-物联网充电桩管理平台.md  # 产品需求文档
│   ├── ARCHITECTURE.md             # 架构设计文档
│   ├── DATABASE.md                 # 数据库设计文档
│   ├── API.md                      # API 接口文档
│   ├── DEPLOYMENT.md               # 部署运维手册
│   ├── USER_MANUAL.md              # 用户操作手册
│   └── CONTRIBUTING.md             # 开发者指南
├── server/                  # Go 后端
│   ├── cmd/main.go          # 入口
│   ├── internal/
│   │   ├── connector/       # 设备接入层 (TCP/MQTT)
│   │   ├── protocol/        # 协议适配引擎
│   │   │   ├── engine/      # 适配器注册中心
│   │   │   ├── model/       # 标准数据模型
│   │   │   └── adapters/    # 各协议适配器
│   │   │       ├── ap3000/  # 电单车充电桩
│   │   │       └── tf100/   # 汽车充电桩
│   │   ├── device/          # 设备管理模块
│   │   └── command/         # 指令管理模块
│   ├── pkg/                 # 公共库
│   │   ├── config/          # 配置管理
│   │   ├── database/        # 数据库连接
│   │   ├── mq/              # 消息队列
│   │   ├── auth/            # 认证鉴权
│   │   └── middleware/      # 中间件
│   └── migrations/          # 数据库迁移
├── web/admin/               # 前端管理后台
│   └── src/views/           # 页面组件
├── deploy/                  # 部署配置
│   └── docker-compose/
└── README.md
```

## 快速开始

### 前置条件

- Go 1.22+
- Docker & Docker Compose
- Node.js 18+ (前端开发)

### 1. 启动基础设施

```bash
cd deploy/docker-compose
docker-compose up -d mysql mongodb redis zookeeper kafka
```

### 2. 配置并启动后端

```bash
cd server
cp config/config.example.yaml config/config.yaml
# 修改 config.yaml 中的数据库密码等配置
go mod tidy
go run cmd/main.go
```

### 3. 启动前端

```bash
cd web/admin
npm install
npm run dev
```

### 4. 访问

- 管理后台: http://localhost:3000
- API 服务: http://localhost:8080
- 设备 TCP 端口: 7000

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/devices | 设备列表 |
| POST | /api/v1/devices | 创建设备 |
| GET | /api/v1/devices/:id | 设备详情 |
| PUT | /api/v1/devices/:id | 更新设备 |
| DELETE | /api/v1/devices/:id | 删除设备 |
| POST | /api/v1/devices/:id/commands | 下发指令 |
| GET | /api/v1/devices/:id/commands | 指令列表 |
| GET | /health | 健康检查 |
| GET | /ready | 就绪检查 |

## 新增协议适配器

1. 在 `server/internal/protocol/adapters/` 下创建新目录
2. 实现 `ProtocolAdapter` 接口
3. 在 `init()` 中调用 `engine.Register()`

```go
type MyAdapter struct{}

func init() {
    engine.Register(&MyAdapter{})
}

func (a *MyAdapter) Name() string { return "MY_PROTOCOL_v1" }
func (a *MyAdapter) Version() string { return "1.0" }
func (a *MyAdapter) DeviceType() string { return "ebike_charger" }
func (a *MyAdapter) Validate(raw []byte) bool { ... }
func (a *MyAdapter) Decode(raw []byte) (*model.StandardData, error) { ... }
func (a *MyAdapter) Encode(cmd *model.StandardCommand) ([]byte, error) { ... }
func (a *MyAdapter) DecodeResponse(raw []byte) (*model.StandardCommandResponse, error) { ... }
```

## 数据流转

```
设备 ──TCP──▶ TCP Server ──原始报文──▶ 协议适配器 ──标准数据──▶ Redis(实时) + Kafka(事件)
                                                                      │
                                                                      ▼
                                                              告警判断 + MongoDB(时序)
                                                                      │
                                                                      ▼
                                                              WebSocket推送前端
```

## 性能指标

| 指标 | 目标 |
|------|------|
| 设备并发连接 | 50,000+ |
| 消息处理吞吐 | 10,000 msg/s |
| 数据上报延迟 | P99 < 200ms |
| 指令下发延迟 | P99 < 3s |
| API 响应时间 | P99 < 500ms |

## 协议文档

- [AP3000 电单车充电桩协议](https://docs.qq.com/doc/DRVdoeFFRaWFQUnRp)
- [TF100 汽车充电桩协议](https://docs.qq.com/doc/DRWx5RFhhcmtmSFhj)
