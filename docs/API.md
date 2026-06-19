# API 接口文档

> IOTsys 物联网充电桩管理平台 · API 参考 v1.0

---

## 一、通用说明

### 1.1 基础 URL

| 环境     | 地址                           |
| -------- | ------------------------------ |
| 本地开发 | `http://localhost:8080/api/v1` |
| 生产环境 | `https://iot.your-domain.com/api/v1` |

### 1.2 认证方式

所有需要认证的接口在请求头中携带 JWT Token：

```
Authorization: Bearer <access_token>
```

### 1.3 通用响应格式

#### 成功响应

```json
{
  "code": "SUCCESS",
  "data": { ... },
  "total": 100,
  "page": 1
}
```

#### 错误响应

```json
{
  "code": "ERROR_CODE",
  "message": "错误描述信息"
}
```

### 1.4 常见错误码

| 错误码              | HTTP 状态码 | 说明             |
| ------------------- | ----------- | ---------------- |
| SUCCESS             | 200         | 操作成功          |
| BAD_REQUEST         | 400         | 请求参数错误      |
| UNAUTHORIZED        | 401         | 未登录或 Token 过期 |
| FORBIDDEN           | 403         | 无权限访问        |
| NOT_FOUND           | 404         | 资源不存在        |
| DEVICE_OFFLINE      | 400         | 设备不在线        |
| INTERNAL_ERROR      | 500         | 服务器内部错误     |

### 1.5 分页参数

| 参数      | 类型   | 默认值 | 说明       |
| --------- | ------ | ------ | ---------- |
| page      | int    | 1      | 页码       |
| page_size | int    | 20     | 每页条数    |

---

## 二、健康检查

### 2.1 健康检查

```
GET /health
```

**无需认证**

**响应示例**：

```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 2.2 就绪检查

```
GET /ready
```

**无需认证**

**响应示例**：

```json
{
  "status": "ok",
  "tcp_connections": 1523
}
```

---

## 三、认证接口

### 3.1 用户登录

```
POST /api/v1/auth/login
```

**无需认证**

**请求体**：

```json
{
  "username": "admin",
  "password": "your-password"
}
```

**响应**：

```json
{
  "code": "SUCCESS",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200,
    "user": {
      "id": "uuid",
      "username": "admin",
      "real_name": "系统管理员",
      "role": "admin"
    }
  }
}
```

> 当前版本登录接口为占位实现，实际认证逻辑待后续开发。

---

## 四、设备管理

### 4.1 设备列表

```
GET /api/v1/devices
```

**需要认证**：是

**查询参数**：

| 参数        | 类型   | 必填 | 说明                              |
| ----------- | ------ | ---- | --------------------------------- |
| device_type | string | 否   | `ebike_charger` / `ev_charger`   |
| protocol    | string | 否   | `AP3000_v2` / `TF100_v1`         |
| status      | string | 否   | `online` / `offline` / `fault` / `maintenance` |
| site_id     | string | 否   | 站点 UUID                          |
| keyword     | string | 否   | 模糊搜索（SN / 安装位置）          |
| page        | int    | 否   | 页码，默认 1                       |
| page_size   | int    | 否   | 每页条数，默认 20                  |

**响应示例**：

```json
{
  "code": "SUCCESS",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "sn": "AP3000-20240101-001",
      "device_type": "ebike_charger",
      "protocol": "AP3000_v2",
      "vendor": "XX科技",
      "model": "AP3000-10S",
      "site_id": "660e8400-e29b-41d4-a716-446655440001",
      "install_location": "A小区B1层001号",
      "firmware_version": "V8.6",
      "status": "online",
      "last_online_at": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 156,
  "page": 1
}
```

### 4.2 设备详情

```
GET /api/v1/devices/:id
```

**需要认证**：是

**路径参数**：

| 参数 | 类型   | 说明    |
| ---- | ------ | ------- |
| id   | string | 设备 ID |

**响应示例**：

```json
{
  "code": "SUCCESS",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "sn": "AP3000-20240101-001",
    "device_type": "ebike_charger",
    "protocol": "AP3000_v2",
    "vendor": "XX科技",
    "model": "AP3000-10S",
    "site_id": "660e8400-e29b-41d4-a716-446655440001",
    "install_location": "A小区B1层001号",
    "firmware_version": "V8.6",
    "status": "online",
    "is_online": true,
    "realtime_data": {
      "voltage": "220.5",
      "current": "15.2",
      "power": "3344.0",
      "energy_total": "12345.6",
      "energy_today": "120.3",
      "temperature": "35.2",
      "charging_status": "charging",
      "charging_progress": "75",
      "updated_at": "2024-01-01T12:00:00Z"
    },
    "last_online_at": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### 4.3 创建设备

```
POST /api/v1/devices
```

**需要认证**：是
**需要角色**：`admin`

**请求体**：

```json
{
  "sn": "AP3000-20240101-001",
  "device_type": "ebike_charger",
  "protocol": "AP3000_v2",
  "vendor": "XX科技",
  "model": "AP3000-10S",
  "site_id": "660e8400-e29b-41d4-a716-446655440001",
  "install_location": "A小区B1层001号",
  "firmware_version": "V8.6"
}
```

| 字段             | 类型   | 必填 | 说明                            |
| ---------------- | ------ | ---- | ------------------------------- |
| sn               | string | ✅   | 设备序列号（需与设备端一致）      |
| device_type      | string | ✅   | `ebike_charger` / `ev_charger`  |
| protocol         | string | ✅   | `AP3000_v2` / `TF100_v1`        |
| vendor           | string | 否   | 设备厂商                        |
| model            | string | 否   | 设备型号                        |
| site_id          | string | 否   | 所属站点 ID                      |
| install_location | string | 否   | 安装位置描述                    |
| firmware_version | string | 否   | 固件版本                        |

**响应**：返回创建的设备对象（同 4.1 设备列表中的格式）。

### 4.4 更新设备

```
PUT /api/v1/devices/:id
```

**需要认证**：是
**需要角色**：`admin`

**请求体**（所有字段可选）：

```json
{
  "site_id": "660e8400-e29b-41d4-a716-446655440002",
  "install_location": "B小区C1层002号",
  "firmware_version": "V8.7",
  "status": "maintenance"
}
```

| 字段             | 类型   | 必填 | 说明                              |
| ---------------- | ------ | ---- | --------------------------------- |
| site_id          | string | 否   | 所属站点 ID                        |
| install_location | string | 否   | 安装位置描述                      |
| firmware_version | string | 否   | 固件版本                          |
| status           | string | 否   | `online` / `offline` / `fault` / `maintenance` |

### 4.5 删除设备

```
DELETE /api/v1/devices/:id
```

**需要认证**：是
**需要角色**：`admin`

**响应**：

```json
{
  "code": "SUCCESS",
  "message": "Device deleted"
}
```

---

## 五、指令管理

### 5.1 下发指令

```
POST /api/v1/devices/:id/commands
```

**需要认证**：是
**需要角色**：`admin` / `operator`

**请求体**：

```json
{
  "cmd_type": "start_charge",
  "params": {
    "port": 1,
    "duration": 120
  }
}
```

| 字段     | 类型   | 必填 | 说明                          |
| -------- | ------ | ---- | ----------------------------- |
| cmd_type | string | ✅   | 指令类型（见下方指令类型表）    |
| params   | object | 否   | 指令参数（依指令类型而定）      |

#### 电单车桩（AP3000）指令类型

| cmd_type       | 说明         | params 参数              |
| -------------- | ------------ | ------------------------ |
| start_charge   | 远程启动充电  | `port` (int, 端口号), `duration` (int, 时长/分钟) |
| stop_charge    | 远程停止充电  | `port` (int, 端口号)     |
| query_status   | 查询状态     | 无                       |
| config         | 参数配置     | `key` (string), `value`  |
| reboot         | 远程重启     | 无                       |
| clear_storage  | 清除存储     | 无                       |
| ota            | OTA 升级     | `url` (string, 固件地址) |

#### 汽车桩（TF100）指令类型

| cmd_type       | 说明         | params 参数              |
| -------------- | ------------ | ------------------------ |
| start_charge   | 远程启动充电  | `port` (int, 枪号)       |
| stop_charge    | 远程停止充电  | `port` (int, 枪号)       |
| set_rate       | 费率下发     | `rates` (array, 费率配置) |
| query_rate     | 查询费率     | 无                       |
| config         | 参数配置     | `key` (string), `value`  |
| reboot         | 远程重启     | 无                       |
| lock_control   | 地锁控制     | `lock_addr` (int), `action` (string: up/down) |
| voice          | 语音下发     | `content` (string)       |
| clear_storage  | 清除存储     | 无                       |

**响应示例**：

```json
{
  "code": "SUCCESS",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "device_id": "550e8400-e29b-41d4-a716-446655440000",
    "cmd_type": "start_charge",
    "payload": {
      "port": 1,
      "duration": 120
    },
    "status": "sent",
    "created_by": "admin-uuid",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

**错误响应**（设备离线）：

```json
{
  "code": "DEVICE_OFFLINE",
  "message": "Device is offline"
}
```

### 5.2 指令列表

```
GET /api/v1/devices/:id/commands
```

**需要认证**：是

**查询参数**：

| 参数      | 类型 | 必填 | 说明             |
| --------- | ---- | ---- | ---------------- |
| page      | int  | 否   | 页码，默认 1      |
| page_size | int  | 否   | 每页条数，默认 20 |

**响应示例**：

```json
{
  "code": "SUCCESS",
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "device_id": "550e8400-e29b-41d4-a716-446655440000",
      "cmd_type": "start_charge",
      "payload": { "port": 1, "duration": 120 },
      "status": "success",
      "result": { "msg": "充电已启动" },
      "created_by": "admin-uuid",
      "created_at": "2024-01-01T12:00:00Z",
      "responded_at": "2024-01-01T12:00:01Z"
    }
  ],
  "total": 42,
  "page": 1
}
```

### 5.3 指令详情

```
GET /api/v1/commands/:id
```

**需要认证**：是

**响应**：返回单条指令对象（同 5.2 列表中的格式）。

---

## 六、站点管理

> 当前版本站点管理接口已在后端实现数据模型，HTTP 处理器待开发。

### 6.1 站点列表

```
GET /api/v1/sites
```

**规划中**

### 6.2 创建站点

```
POST /api/v1/sites
```

**规划中**

---

## 七、告警管理

> 当前版本告警模块待开发。

### 7.1 告警列表

```
GET /api/v1/alerts
```

**规划中**

### 7.2 处理告警

```
PUT /api/v1/alerts/:id
```

**规划中**

---

## 八、数据报表

> 当前版本报表模块待开发。

### 8.1 充电统计

```
GET /api/v1/reports/charging
```

**规划中**

---

## 九、WebSocket 实时推送

### 9.1 设备实时数据

```
WS /ws/device/:id
```

**需要认证**：通过 URL 参数传递 Token：
```
ws://localhost:8080/ws/device/550e8400?token=<access_token>
```

**推送数据格式**：

```json
{
  "type": "device_data",
  "device_id": "550e8400-e29b-41d4-a716-446655440000",
  "data": {
    "voltage": 220.5,
    "current": 15.2,
    "power": 3344.0,
    "charging_status": "charging",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### 9.2 平台概览

```
WS /ws/dashboard
```

**推送数据格式**：

```json
{
  "type": "dashboard_update",
  "data": {
    "online_count": 1523,
    "total_count": 5000,
    "today_energy": 12345.6,
    "active_alerts": 3
  }
}
```

---

## 十、错误响应格式

所有错误响应遵循统一格式：

```json
{
  "code": "ERROR_CODE",
  "message": "人类可读的错误描述",
  "details": { ... }
}
```

### 10.1 错误码列表

| 错误码              | HTTP 状态码 | 说明                          |
| ------------------- | ----------- | ----------------------------- |
| SUCCESS             | 200         | 操作成功                       |
| BAD_REQUEST         | 400         | 请求参数校验失败               |
| UNAUTHORIZED        | 401         | 未提供有效 Token 或 Token 过期  |
| FORBIDDEN           | 403         | 当前角色无权限执行此操作        |
| NOT_FOUND           | 404         | 请求的资源不存在               |
| METHOD_NOT_ALLOWED  | 405         | 不支持的 HTTP 方法             |
| CONFLICT            | 409         | 资源冲突（如 SN 重复）          |
| DEVICE_OFFLINE      | 400         | 目标设备当前离线                |
| COMMAND_TIMEOUT     | 408         | 指令下发超时                   |
| RATE_LIMIT          | 429         | 请求频率超过限制               |
| INTERNAL_ERROR      | 500         | 服务器内部错误                 |

### 10.2 请求校验错误示例

```json
{
  "code": "BAD_REQUEST",
  "message": "参数校验失败",
  "details": {
    "sn": "设备序列号不能为空",
    "device_type": "设备类型必须是 ebike_charger 或 ev_charger"
  }
}
```

---

## 十一、请求限制

| 限制项       | 值            |
| ------------ | ------------- |
| 请求体最大   | 1 MB          |
| 请求超时     | 30 秒         |
| 单 IP 频率   | 100 req/s     |
| Token 有效期 | 2 小时 (Access) |
