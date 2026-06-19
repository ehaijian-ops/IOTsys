# 数据库设计文档

> IOTsys 物联网充电桩管理平台 · 数据库设计 v1.0

---

## 一、数据库选型

| 存储引擎  | 用途                     | 版本   |
| --------- | ------------------------ | ------ |
| MySQL     | 业务数据（设备、站点、指令、用户、订单） | 8.0+   |
| Redis     | 缓存（实时数据、在线状态、会话信息）     | 7.x    |
| MongoDB   | 时序数据（充电记录、历史数据、日志）     | 7.0+   |
| Kafka     | 消息队列（数据上报、事件、指令流转）     | 3.x    |

---

## 二、MySQL 表结构

### 2.1 设备表 `devices`

| 字段             | 类型          | 约束        | 说明                     |
| ---------------- | ------------- | ----------- | ------------------------ |
| id               | VARCHAR(36)   | PK          | UUID                     |
| sn               | VARCHAR(64)   | UNIQUE, NOT NULL | 设备序列号                |
| device_type      | VARCHAR(20)   | NOT NULL    | `ebike_charger` / `ev_charger` |
| protocol         | VARCHAR(50)   | NOT NULL    | `AP3000_v2` / `TF100_v1` |
| vendor           | VARCHAR(100)  |             | 设备厂商                 |
| model            | VARCHAR(100)  |             | 设备型号                 |
| site_id          | VARCHAR(36)   | INDEX       | 所属站点 ID              |
| install_location | VARCHAR(255)  |             | 安装位置                 |
| firmware_version | VARCHAR(50)   |             | 固件版本                 |
| status           | VARCHAR(20)   | DEFAULT 'offline' | `online` / `offline` / `fault` / `maintenance` |
| last_online_at   | DATETIME      |             | 最后上线时间             |
| created_at       | DATETIME      | NOT NULL    | 创建时间                 |
| updated_at       | DATETIME      | NOT NULL    | 更新时间                 |

**索引**：
- `idx_site_id` ON `site_id`
- `idx_status` ON `status`
- `idx_device_type_protocol` ON `(device_type, protocol)`

```sql
CREATE TABLE `devices` (
  `id` varchar(36) NOT NULL,
  `sn` varchar(64) NOT NULL,
  `device_type` varchar(20) NOT NULL,
  `protocol` varchar(50) NOT NULL,
  `vendor` varchar(100) DEFAULT '',
  `model` varchar(100) DEFAULT '',
  `site_id` varchar(36) DEFAULT '',
  `install_location` varchar(255) DEFAULT '',
  `firmware_version` varchar(50) DEFAULT '',
  `status` varchar(20) DEFAULT 'offline',
  `last_online_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sn` (`sn`),
  KEY `idx_site_id` (`site_id`),
  KEY `idx_status` (`status`),
  KEY `idx_device_type_protocol` (`device_type`, `protocol`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2.2 站点表 `sites`

| 字段      | 类型          | 约束        | 说明                     |
| --------- | ------------- | ----------- | ------------------------ |
| id        | VARCHAR(36)   | PK          | UUID                     |
| name      | VARCHAR(100)  | NOT NULL    | 站点名称                 |
| address   | VARCHAR(255)  |             | 详细地址                 |
| latitude  | DOUBLE        |             | 纬度                     |
| longitude | DOUBLE        |             | 经度                     |
| contact   | VARCHAR(50)   |             | 联系人                   |
| phone     | VARCHAR(20)   |             | 联系电话                 |
| status    | VARCHAR(20)   | DEFAULT 'active' | `active` / `inactive` |
| created_at| DATETIME      | NOT NULL    | 创建时间                 |
| updated_at| DATETIME      | NOT NULL    | 更新时间                 |

```sql
CREATE TABLE `sites` (
  `id` varchar(36) NOT NULL,
  `name` varchar(100) NOT NULL,
  `address` varchar(255) DEFAULT '',
  `latitude` double DEFAULT 0,
  `longitude` double DEFAULT 0,
  `contact` varchar(50) DEFAULT '',
  `phone` varchar(20) DEFAULT '',
  `status` varchar(20) DEFAULT 'active',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2.3 设备指令表 `device_commands`

| 字段         | 类型          | 约束        | 说明                            |
| ------------ | ------------- | ----------- | ------------------------------- |
| id           | VARCHAR(36)   | PK          | UUID                            |
| device_id    | VARCHAR(36)   | INDEX, NOT NULL | 关联设备 ID                 |
| cmd_type     | VARCHAR(50)   | NOT NULL    | 指令类型                        |
| payload      | JSON          |             | 指令参数                        |
| status       | VARCHAR(20)   | DEFAULT 'pending' | `pending` / `sent` / `responded` / `success` / `failed` / `timeout` |
| result       | JSON          |             | 执行结果                        |
| created_by   | VARCHAR(36)   |             | 操作人 ID                       |
| created_at   | DATETIME      | NOT NULL    | 创建时间                        |
| responded_at | DATETIME      |             | 设备响应时间                    |

**索引**：
- `idx_device_id` ON `device_id`
- `idx_status` ON `status`
- `idx_created_at` ON `created_at`

```sql
CREATE TABLE `device_commands` (
  `id` varchar(36) NOT NULL,
  `device_id` varchar(36) NOT NULL,
  `cmd_type` varchar(50) NOT NULL,
  `payload` json DEFAULT NULL,
  `status` varchar(20) DEFAULT 'pending',
  `result` json DEFAULT NULL,
  `created_by` varchar(36) DEFAULT '',
  `created_at` datetime NOT NULL,
  `responded_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_device_id` (`device_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2.4 用户表 `users`（规划中）

| 字段       | 类型          | 约束        | 说明                     |
| ---------- | ------------- | ----------- | ------------------------ |
| id         | VARCHAR(36)   | PK          | UUID                     |
| username   | VARCHAR(64)   | UNIQUE, NOT NULL | 用户名              |
| password   | VARCHAR(255)  | NOT NULL    | bcrypt 加密              |
| real_name  | VARCHAR(50)   |             | 真实姓名                 |
| email      | VARCHAR(128)  |             | 邮箱                     |
| phone      | VARCHAR(20)   |             | 手机号                   |
| role       | VARCHAR(20)   | NOT NULL    | `admin` / `operator` / `viewer` |
| status     | VARCHAR(20)   | DEFAULT 'active' | `active` / `disabled` |
| last_login | DATETIME      |             | 最后登录时间             |
| created_at | DATETIME      | NOT NULL    | 创建时间                 |
| updated_at | DATETIME      | NOT NULL    | 更新时间                 |

### 2.5 告警记录表 `alerts`（规划中）

| 字段       | 类型          | 约束        | 说明                            |
| ---------- | ------------- | ----------- | ------------------------------- |
| id         | VARCHAR(36)   | PK          | UUID                            |
| device_id  | VARCHAR(36)   | INDEX, NOT NULL | 关联设备 ID                 |
| alert_type | VARCHAR(50)   | NOT NULL    | `offline` / `fault` / `overload` / `high_temp` / `leakage` |
| severity   | VARCHAR(20)   | NOT NULL    | `critical` / `warning` / `info` |
| message    | VARCHAR(500)  |             | 告警描述                        |
| status     | VARCHAR(20)   | DEFAULT 'active' | `active` / `resolved` / `acknowledged` |
| resolved_by| VARCHAR(36)   |             | 处理人 ID                       |
| resolved_at| DATETIME      |             | 处理时间                        |
| created_at | DATETIME      | NOT NULL    | 告警时间                        |

### 2.6 充电记录表 `charge_records`（规划中）

| 字段         | 类型          | 约束        | 说明                     |
| ------------ | ------------- | ----------- | ------------------------ |
| id           | VARCHAR(36)   | PK          | UUID                     |
| device_id    | VARCHAR(36)   | INDEX, NOT NULL | 关联设备 ID          |
| port_num     | INT           |             | 端口号                   |
| order_no     | VARCHAR(64)   | UNIQUE      | 订单号                   |
| card_id      | VARCHAR(32)   |             | 刷卡卡号                 |
| start_time   | DATETIME      | NOT NULL    | 开始充电时间             |
| end_time     | DATETIME      |             | 结束充电时间             |
| duration     | INT           |             | 充电时长（秒）           |
| energy       | DECIMAL(10,2) |             | 充电电量（kWh）          |
| amount       | DECIMAL(10,2) |             | 充电金额（元）           |
| stop_reason  | INT           |             | 停止原因码               |
| start_type   | INT           |             | 启动方式                 |
| status       | VARCHAR(20)   | NOT NULL    | `charging` / `finished` / `abnormal` |
| created_at   | DATETIME      | NOT NULL    | 创建时间                 |

---

## 三、ER 关系图

```
┌──────────┐       ┌──────────────┐       ┌───────────────────┐
│   sites  │       │   devices    │       │  device_commands  │
├──────────┤       ├──────────────┤       ├───────────────────┤
│ id (PK)  │──<    │ id (PK)      │──<    │ id (PK)           │
│ name     │       │ sn (UQ)      │       │ device_id (FK)    │
│ address  │       │ site_id (FK) │       │ cmd_type          │
│ ...      │       │ device_type  │       │ payload (JSON)    │
└──────────┘       │ protocol     │       │ status            │
                   │ vendor       │       │ result (JSON)     │
                   │ status       │       │ ...               │
                   └──────────────┘       └───────────────────┘
                         │
                         │
          ┌──────────────┼──────────────┐
          │              │              │
    ┌─────┴─────┐  ┌────┴─────┐  ┌─────┴──────┐
    │  alerts   │  │  users   │  │charge_records│
    ├───────────┤  ├──────────┤  ├─────────────┤
    │ id (PK)   │  │ id (PK)  │  │ id (PK)     │
    │ device_id │  │ username │  │ device_id   │
    │ alert_type│  │ role     │  │ order_no    │
    │ severity  │  │ ...      │  │ energy      │
    │ status    │  └──────────┘  │ amount      │
    └───────────┘                └─────────────┘
```

**关系说明**：
- `sites` 1 : N `devices`（一个站点有多台设备）
- `devices` 1 : N `device_commands`（一台设备有多条指令）
- `devices` 1 : N `alerts`（一台设备有多条告警）
- `devices` 1 : N `charge_records`（一台设备有多条充电记录）

---

## 四、Redis 数据结构

### 4.1 设备实时数据

| Key                      | 类型 | TTL   | 说明                  |
| ------------------------ | ---- | ----- | --------------------- |
| `device:realtime:{id}`   | Hash | 5 min | 设备最新采集数据       |

**Hash 字段**：

| 字段               | 类型   | 说明           |
| ------------------ | ------ | -------------- |
| voltage            | float  | 电压 (V)       |
| current            | float  | 电流 (A)       |
| power              | float  | 功率 (W)       |
| energy_total       | float  | 累计电量 (kWh) |
| energy_today       | float  | 今日电量 (kWh) |
| temperature        | float  | 温度 (℃)       |
| charging_status    | string | 充电状态        |
| charging_progress  | int    | 充电进度 (%)   |
| updated_at         | string | 更新时间 (RFC3339) |

### 4.2 在线状态

| Key                      | 类型   | TTL   | 说明                  |
| ------------------------ | ------ | ----- | --------------------- |
| `device:online:{id}`     | String | 2 min | 值为 `"1"` 表示在线   |

### 4.3 会话信息

| Key                      | 类型 | 说明                     |
| ------------------------ | ---- | ------------------------ |
| `device:session:{id}`    | Hash | 设备当前连接会话信息       |

**Hash 字段**：
- `connector_id` — 连接器 ID（UUID）
- `connected_at` — 连接建立时间
- `ip` — 设备 IP 地址
- `protocol` — 通信协议

### 4.4 在线设备集合

| Key                    | 类型 | 说明                      |
| ---------------------- | ---- | ------------------------- |
| `devices:online`       | Set  | 所有在线设备 ID 集合       |
| `protocol:online:{p}`  | Set  | 按协议分组的在线设备 ID    |

### 4.5 JWT Token 黑名单（规划中）

| Key                    | 类型   | TTL        | 说明              |
| ---------------------- | ------ | ---------- | ----------------- |
| `jwt:blacklist:{jti}`  | String | Token 剩余有效期 | 已注销的 Token   |

---

## 五、MongoDB 集合设计

### 5.1 设备原始数据 `device_raw_data`

存储设备上报的原始数据（含协议原始帧），用于历史回溯和审计。

```json
{
  "_id": "ObjectId",
  "device_id": "uuid-string",
  "protocol": "AP3000_v2",
  "msg_id": 12345,
  "raw_frame": "hex-string",
  "standard_data": { /* StandardData JSON */ },
  "received_at": "ISODate"
}
```

**索引**：
- `{ device_id: 1, received_at: -1 }` — 按设备查询历史
- `{ received_at: 1 }` — 按时间范围查询
- TTL 索引：`{ received_at: 1 }` expireAfterSeconds: 7776000（90 天自动清理）

### 5.2 充电记录 `charge_records`（规划中）

存储完整的充电会话记录。

```json
{
  "_id": "ObjectId",
  "device_id": "uuid-string",
  "order_no": "ORD20240101001",
  "port_num": 1,
  "card_id": "CARD001",
  "card_type": 1,
  "start_time": "ISODate",
  "end_time": "ISODate",
  "duration": 3600,
  "energy": 1.5,
  "amount": 1.2,
  "stop_reason": 0,
  "start_type": 1,
  "charge_mode": 0,
  "peak_power": 800,
  "created_at": "ISODate"
}
```

**索引**：
- `{ device_id: 1, start_time: -1 }`
- `{ order_no: 1 }` unique
- `{ start_time: 1 }`

### 5.3 设备告警历史 `device_alerts`（规划中）

```json
{
  "_id": "ObjectId",
  "device_id": "uuid-string",
  "alert_type": "fault",
  "severity": "warning",
  "message": "端口3 继电器粘连",
  "fault_code": "0x09",
  "status": "active",
  "created_at": "ISODate",
  "resolved_at": "ISODate",
  "resolved_by": "user-id"
}
```

### 5.4 操作审计日志 `audit_logs`（规划中）

```json
{
  "_id": "ObjectId",
  "user_id": "uuid-string",
  "username": "admin",
  "action": "send_command",
  "resource_type": "device",
  "resource_id": "device-uuid",
  "detail": { "cmd_type": "start_charge", "params": {} },
  "ip": "192.168.1.100",
  "user_agent": "...",
  "created_at": "ISODate"
}
```

---

## 六、Kafka Topic 设计

| Topic                    | 分区数 | 消费者组            | 说明                  |
| ------------------------ | ------ | ------------------- | --------------------- |
| `device.data.report`     | 8      | `data-processor`    | 设备数据上报           |
| `device.event.online`    | 4      | `event-processor`   | 设备上线事件           |
| `device.event.offline`   | 4      | `event-processor`   | 设备离线事件           |
| `device.event.fault`     | 4      | `alert-service`     | 设备故障事件           |
| `device.command.send`    | 4      | `command-executor`  | 指令下发记录           |
| `device.command.result`  | 4      | `command-processor` | 指令执行结果           |
| `alert.triggered`        | 4      | `alert-service`     | 告警触发              |
| `alert.resolved`         | 4      | `alert-service`     | 告警解除              |
| `device.ota.upgrade`     | 2      | `ota-service`       | OTA 升级指令           |
| `device.ota.progress`    | 2      | `ota-service`       | OTA 升级进度           |
| `device.ota.result`      | 2      | `ota-service`       | OTA 升级结果           |

---

## 七、数据流转说明

```
设备 TCP 连接
    │
    ▼
tcpserver (协议检测 → Decode → StandardData)
    │
    ├── Redis: device:realtime:{id} (实时缓存, TTL 5min)
    ├── Redis: device:online:{id}   (在线标记, TTL 2min)
    │
    └── Kafka: device.data.report
              │
              ▼
         data-processor (消费者)
              │
              ├── MongoDB: device_raw_data (原始数据存储)
              └── MySQL: 更新设备状态
```

---

## 八、容量估算

以 **5 万设备** 接入为基准：

| 存储     | 估算数据量                      | 日增量        |
| -------- | ------------------------------- | ------------- |
| MySQL    | 5 万设备 + 1 万站点 + 历史指令   | ~50 MB/月     |
| Redis    | 5 万实时 Hash + Set             | ~2 GB 内存    |
| MongoDB  | 每条上报 ~1KB，30s/条            | ~144 GB/天    |
| Kafka    | 同上吞吐量                       | ~5 MB/s 峰值  |

**建议**：
- MongoDB 启用 TTL 索引，原始数据保留 30-90 天
- MySQL 指令表按季度归档
- Redis 使用 Cluster 模式或哨兵高可用
