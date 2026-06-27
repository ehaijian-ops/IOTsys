<template>
  <div class="logs-container">
    <!-- 顶部过滤栏 -->
    <el-card class="filter-card" shadow="never">
      <el-row :gutter="12" align="middle">
        <el-col :span="5">
          <el-input
            v-model="filter.deviceId"
            placeholder="设备ID"
            clearable
            @keyup.enter="applyFilter"
          />
        </el-col>
        <el-col :span="4">
          <el-select v-model="filter.protocol" placeholder="设备型号/协议" clearable>
            <el-option label="TF100" value="TF100_v1" />
            <el-option label="AP3000" value="AP3000_v2" />
          </el-select>
        </el-col>
        <el-col :span="5">
          <el-input
            v-model="filter.keyword"
            placeholder="关键字搜索"
            clearable
            @keyup.enter="applyFilter"
          />
        </el-col>
        <el-col :span="6">
          <el-button type="primary" @click="applyFilter">
            <el-icon><Search /></el-icon> 应用过滤
          </el-button>
          <el-button @click="clearFilter">
            <el-icon><RefreshLeft /></el-icon> 重置
          </el-button>
          <el-button @click="clearLogs">
            <el-icon><Delete /></el-icon> 清空
          </el-button>
        </el-col>
        <el-col :span="4" style="text-align: right">
          <el-switch
            v-model="autoScroll"
            active-text="自动滚动"
            inactive-text="暂停"
          />
        </el-col>
      </el-row>
    </el-card>

    <!-- 状态栏 -->
    <div class="status-bar">
      <el-tag :type="connected ? 'success' : 'danger'" size="small" effect="dark">
        <el-icon style="margin-right: 4px"><component :is="connected ? 'CircleCheck' : 'CircleClose'" /></el-icon>
        {{ connected ? '已连接' : '已断开' }}
      </el-tag>
      <span class="stat-item">设备组: <b>{{ groupedLogs.length }}</b></span>
      <span class="stat-item">消息: <b>{{ totalFiltered }}</b></span>
      <span class="stat-item">↓接收: <b style="color:#89b4fa">{{ rxTotal }}</b></span>
      <span class="stat-item">↑发送: <b style="color:#a6e3a1">{{ txTotal }}</b></span>
      <span class="stat-item" v-if="connected && logCount === 0" style="color: #909399">
        等待设备数据...
      </span>
      <span class="stat-item" style="margin-left: auto">
        <el-button size="small" text @click="expandAll">展开全部</el-button>
        <el-button size="small" text @click="collapseAll">收起全部</el-button>
      </span>
    </div>

    <!-- 日志列表（树形分组） -->
    <el-card class="log-card" shadow="never">
      <div class="log-list" ref="logContainer">
        <div v-if="groupedLogs.length === 0" class="log-empty">
          <el-icon :size="48" color="#c0c4cc"><Document /></el-icon>
          <p>{{ connected ? '等待设备上报数据...' : '正在连接服务器...' }}</p>
        </div>

        <template v-for="group in groupedLogs" :key="group.key">
          <!-- 设备组头部 -->
          <div class="group-header" @click="toggleGroup(group.key)">
            <el-icon class="group-arrow" :class="{ expanded: isExpanded(group.key) }">
              <ArrowRight />
            </el-icon>
            <span class="group-icon-device">🖥</span>
            <span class="group-device-id">{{ group.deviceId || '-' }}</span>
            <span class="group-protocol" v-if="group.protocol">{{ group.protocol }}</span>
            <span class="group-addr" v-if="group.remoteAddr">{{ group.remoteAddr }}</span>
            <span class="group-stats">
              <span class="grp-rx">↓{{ group.rxCount }}</span>
              <span class="grp-tx">↑{{ group.txCount }}</span>
              <span class="grp-total">共{{ group.entries.length }}条</span>
            </span>
          </div>

          <!-- 展开的日志条目 -->
          <div v-show="isExpanded(group.key)" class="group-body">
            <div
              v-for="log in group.entries"
              :key="log._id"
              class="log-item"
              :class="[logStatusClass(log), logDirectionClass(log)]"
            >
              <!-- 方向标识 -->
              <span class="log-direction" :class="directionClass(log)">
                {{ directionLabel(log) }}
              </span>
              <span class="log-time">{{ log.timestamp }}</span>
              <span class="log-type" :class="logTypeClass(log)">[{{ displayTypeLabel(log) }}]</span>
              <span class="log-values">
                <!-- 原文报文：始终显示，点击复制完整HEX -->
                <span
                  v-if="log.raw_hex"
                  class="val-raw-hex hex-copy-btn"
                  :title="'点击复制完整报文: ' + log.raw_hex"
                  @click="copyHex(log.raw_hex)"
                >{{ truncateHex(log.raw_hex) }}</span>
                <template v-if="log.type === 'data'">
                  <span class="val-voltage">V:{{ log.voltage?.toFixed(1) }}V</span>
                  <span class="val-current">A:{{ log.current?.toFixed(1) }}A</span>
                  <span class="val-power">P:{{ log.power?.toFixed(0) }}W</span>
                  <span class="val-energy">{{ log.energy_total?.toFixed(2) }}kWh</span>
                  <span :class="['val-status', statusBadgeClass(log.charging_status)]">{{ statusLabel(log.charging_status) }}</span>
                </template>
                <template v-else-if="log.type === 'raw' || log.type === 'reply'">
                  <span :class="['val-status', log.type === 'reply' ? 'st-reply' : 'st-raw']">
                    {{ log.type === 'reply' ? cmdLabel(log.status) : (log.charging_status || log.status) }}
                  </span>
                </template>
                <template v-else-if="log.type === 'sim_card'">
                  <span class="val-sim">{{ log.status }}</span>
                </template>
                <template v-if="log.fault_code">
                  <span class="val-fault">故障:{{ log.fault_code }}</span>
                </template>
              </span>
              <span class="log-temp" v-if="log.temperature">T:{{ log.temperature }}°C</span>
            </div>
          </div>
        </template>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowRight } from '@element-plus/icons-vue'

// 日志条目类型
interface LogEntry {
  _id: number
  device_id: string
  protocol: string
  timestamp: string
  msg_id: number
  msg_type: string
  voltage: number
  current: number
  power: number
  energy_total: number
  charging_status: string
  status: string
  temperature: number
  fault_code: string
  remote_addr: string
  raw_hex: string
  type: string
  direction: string  // "rx" | "tx"
}

// 树形节点
interface GroupNode {
  key: string
  deviceId: string
  protocol: string
  remoteAddr: string
  entries: LogEntry[]
  rxCount: number
  txCount: number
}

let logIdCounter = 0
const connected = ref(false)
const autoScroll = ref(true)
const allLogs = ref<LogEntry[]>([])
const logContainer = ref<HTMLElement | null>(null)
const expandedGroups = reactive<Set<string>>(new Set())
let eventSource: EventSource | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

const filter = reactive({
  deviceId: '',
  protocol: '',
  keyword: '',
})

// ===== 计算属性 =====
const logCount = computed(() => allLogs.value.length)

// 将过滤后的日志按 remote_addr 分组
const groupedLogs = computed<GroupNode[]>(() => {
  const list = filteredLogs.value
  const map = new Map<string, GroupNode>()

  for (const log of list) {
    const key = log.remote_addr || '__orphan__'
    let group = map.get(key)
    if (!group) {
      group = {
        key,
        deviceId: log.device_id || '',
        protocol: (log.protocol && log.protocol !== 'SIM_CARD') ? log.protocol : '',
        remoteAddr: log.remote_addr || '',
        entries: [],
        rxCount: 0,
        txCount: 0,
      }
      map.set(key, group)
    }
    // 合并设备信息（取最新的非空值）
    if (log.device_id && !group.deviceId) group.deviceId = log.device_id
    // 优先显示真实协议名（SIM_CARD/WSD/Unknown 等占位协议可被真实协议覆盖）
    if (log.protocol && log.protocol !== 'SIM_CARD' && (!group.protocol || group.protocol === 'SIM_CARD')) group.protocol = log.protocol
    else if (log.protocol && !group.protocol) group.protocol = log.protocol

    group.entries.push(log)
    if (log.direction === 'tx') group.txCount++
    else group.rxCount++
  }

  // 排序：orphan 放最后，其他按最新消息倒序
  const result = Array.from(map.values())
  result.sort((a, b) => {
    if (a.key === '__orphan__') return 1
    if (b.key === '__orphan__') return -1
    const aLast = a.entries[a.entries.length - 1]?.timestamp || ''
    const bLast = b.entries[b.entries.length - 1]?.timestamp || ''
    return bLast.localeCompare(aLast)
  })
  return result
})

const filteredLogs = computed(() => {
  let list = allLogs.value
  if (filter.deviceId) {
    list = list.filter(l => l.device_id?.includes(filter.deviceId))
  }
  if (filter.protocol) {
    list = list.filter(l => l.protocol === filter.protocol)
  }
  if (filter.keyword) {
    const kw = filter.keyword.toLowerCase()
    list = list.filter(l => JSON.stringify(l).toLowerCase().includes(kw))
  }
  return list
})

const totalFiltered = computed(() => filteredLogs.value.length)
const rxTotal = computed(() => filteredLogs.value.filter(l => l.direction !== 'tx').length)
const txTotal = computed(() => filteredLogs.value.filter(l => l.direction === 'tx').length)

// ===== 展开/收起 =====
function isExpanded(key: string) {
  return expandedGroups.has(key)
}

function toggleGroup(key: string) {
  if (expandedGroups.has(key)) {
    expandedGroups.delete(key)
  } else {
    expandedGroups.add(key)
  }
}

function expandAll() {
  for (const g of groupedLogs.value) {
    expandedGroups.add(g.key)
  }
}

function collapseAll() {
  expandedGroups.clear()
}

// 自动展开新设备组
watch(groupedLogs, (groups) => {
  for (const g of groups) {
    // 新出现的设备组自动展开
    if (!expandedGroups.has(g.key) && g.entries.length <= 3) {
      expandedGroups.add(g.key)
    }
  }
})

// ===== 自动滚动 =====
watch(groupedLogs, () => {
  if (autoScroll.value) {
    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    })
  }
}, { flush: 'post', deep: true })

// ===== SSE 连接 =====
function connectSSE() {
  if (eventSource) eventSource.close()

  const params = new URLSearchParams()
  if (filter.deviceId) params.set('device_id', filter.deviceId)
  if (filter.protocol) params.set('protocol', filter.protocol)
  if (filter.keyword) params.set('keyword', filter.keyword)

  const url = `/api/v1/devices/logs/stream`
  const fullUrl = params.toString() ? `${url}?${params}` : url

  eventSource = new EventSource(fullUrl)

  eventSource.addEventListener('connected', (e) => {
    const data = JSON.parse(e.data)
    console.log('SSE connected:', data.client_id)
    connected.value = true
  })

  eventSource.addEventListener('log', (e) => {
    try {
      const entry: LogEntry = JSON.parse(e.data)
      entry._id = ++logIdCounter
      allLogs.value.push(entry)
      if (allLogs.value.length > 3000) {
        allLogs.value.splice(0, 800)
      }
    } catch (err) {
      console.error('Parse log error:', err)
    }
  })

  eventSource.onerror = () => {
    connected.value = false
    eventSource?.close()
    scheduleReconnect()
  }

  eventSource.onopen = () => {
    connected.value = true
  }
}

function scheduleReconnect() {
  if (reconnectTimer) return
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    if (!connected.value) connectSSE()
  }, 3000)
}

function applyFilter() {
  if (eventSource) eventSource.close()
  connectSSE()
  ElMessage.success('过滤条件已更新')
}

function clearFilter() {
  filter.deviceId = ''
  filter.protocol = ''
  filter.keyword = ''
  applyFilter()
}

function clearLogs() {
  allLogs.value = []
  expandedGroups.clear()
  ElMessage.success('日志已清空')
}

// ===== 样式函数 =====
function logStatusClass(log: LogEntry) {
  if (log.fault_code) return 'log-fault'
  if (log.charging_status === 'charging') return 'log-charging'
  return ''
}

function logDirectionClass(log: LogEntry) {
  return log.direction === 'tx' ? 'log-tx' : 'log-rx'
}

function directionClass(log: LogEntry) {
  return log.direction === 'tx' ? 'dir-tx' : 'dir-rx'
}

function directionLabel(log: LogEntry) {
  return log.direction === 'tx' ? '↑ 发送' : '↓ 接收'
}

function logTypeClass(log: LogEntry) {
  if (log.type === 'data') return 'type-data'
  if (log.type === 'reply') return 'type-reply'
  if (log.type === 'raw') return 'type-raw'
  if (log.type === 'sim_card') return 'type-sim'
  return 'type-event'
}

function typeLabel(type: string) {
  const map: Record<string, string> = {
    data: '数据',
    raw: '原始',
    reply: '应答',
    connect: '连接',
    disconnect: '断开',
  }
  return map[type] || type || '数据'
}

// 显示类型标签：优先使用协议层报文类型
function displayTypeLabel(log: LogEntry) {
  if (log.msg_type) return msgTypeLabel(log.msg_type)
  return typeLabel(log.type)
}

// 协议层报文类型 → 中文标签
function msgTypeLabel(msgType: string) {
  const map: Record<string, string> = {
    login: '登录',
    register: '注册',
    heartbeat: '心跳上报',
    time_request: '校时请求',
    get_time: '校时请求',
    all_ports_reply: '全端口状态',
    one_port_reply: '单端口状态',
    local_charge: '本地充电',
    swipe_card: '刷卡',
    remote_start_ack: '远程启动ACK',
    remote_stop_ack: '远程停止ACK',
    settlement: '结算上报',
    fault_report: '故障上报',
    gear_report: '分档上报',
    param_set_ack: '参数应答',
    config_upload: '配置上传',
    platform_param: '平台参数',
    query_card: '查询卡',
    remote_ctrl_ack: '远程控制ACK',
    hb_interval_ack: '心跳间隔ACK',
    charge_record: '充电记录',
    port_confirm: '端口确认',
    charge_progress: '充电中',
    charge_record_v2: '充电记录V2',
    charge_progress_v2: '充电中V2',
    lock_notify: '充满通知',
    port_alarm: '端口报警',
    device_alarm: '设备报警',
    comm_info: '通信信息',
    meter_energy: '电表读数',
    push_info: '推送信息',
    lock_heartbeat: '锁心跳',
    lock_occupied: '锁占用',
    lock_status: '锁状态',
    lock_addr_query: '锁地址查询',
    gun_timeline: '枪时间线',
    status_report: '状态上报',
    sim_card: 'SIM卡号',
  }
  return map[msgType] || msgType
}

// 截断十六进制显示（前8字节 + ...）
function truncateHex(hex: string) {
  if (!hex) return ''
  if (hex.length <= 16) return hex
  return hex.slice(0, 16) + '...'
}

// 点击复制完整报文
async function copyHex(hex: string) {
  if (!hex) return
  try {
    await navigator.clipboard.writeText(hex)
    ElMessage({ message: '报文已复制到剪贴板', type: 'success', duration: 1500, showClose: false })
  } catch {
    // 降级方案
    const ta = document.createElement('textarea')
    ta.value = hex
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    ElMessage({ message: '报文已复制', type: 'success', duration: 1500, showClose: false })
  }
}

function cmdLabel(cmd: string) {
  const map: Record<string, string> = {
    login: '登录ACK',
    heartbeat: '心跳ACK',
    get_time: '校时',
  }
  return map[cmd] || cmd || '回复'
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    idle: '空闲',
    charging: '充电中',
    finished: '已完成',
    fault: '故障',
  }
  return map[status] || status || '-'
}

function statusBadgeClass(status: string) {
  const map: Record<string, string> = {
    charging: 'st-charging',
    idle: 'st-idle',
    finished: 'st-finished',
    fault: 'st-fault',
  }
  return map[status] || 'st-idle'
}

onMounted(() => connectSSE())
onUnmounted(() => {
  if (eventSource) { eventSource.close(); eventSource = null }
  if (reconnectTimer) clearTimeout(reconnectTimer)
})
</script>

<style scoped>
.logs-container { padding: 4px; }
.filter-card { margin-bottom: 10px; }
.filter-card :deep(.el-card__body) { padding: 14px 18px; }

/* 状态栏 */
.status-bar {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 10px;
  padding: 8px 12px;
  background: #fff;
  border-radius: 6px;
  border: 1px solid #ebeef5;
  font-size: 13px;
  color: #606266;
}
.stat-item { font-size: 13px; }

/* 日志区域 */
.log-card { height: calc(100vh - 210px); }
.log-card :deep(.el-card__body) { padding: 0; height: 100%; }
.log-list {
  height: 100%;
  overflow-y: auto;
  background: #1e1e2e;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  padding: 4px 0;
}
.log-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #6c6c8a;
  font-size: 14px;
  gap: 12px;
}

/* ===== 设备组头部 ===== */
.group-header {
  display: flex;
  align-items: center;
  padding: 6px 12px 6px 8px;
  cursor: pointer;
  user-select: none;
  background: #262640;
  border-bottom: 1px solid #31315a;
  gap: 8px;
  transition: background 0.15s;
}
.group-header:hover { background: #2d2d4a; }
.group-arrow {
  color: #6c7086;
  font-size: 12px;
  transition: transform 0.2s;
  flex-shrink: 0;
}
.group-arrow.expanded { transform: rotate(90deg); }
.group-icon-device { font-size: 14px; flex-shrink: 0; }
.group-device-id { color: #f5c2e7; font-weight: 600; flex-shrink: 0; }
.group-protocol { color: #94e2d5; font-size: 11px; flex-shrink: 0; }
.group-addr { color: #585b70; font-size: 11px; }
.group-stats {
  margin-left: auto;
  display: flex;
  gap: 12px;
  font-size: 12px;
  flex-shrink: 0;
}
.grp-rx { color: #89b4fa; }
.grp-tx { color: #a6e3a1; }
.grp-total { color: #6c7086; }

/* ===== 日志条目 ===== */
.group-body {
  border-left: 2px solid #363652;
  margin-left: 12px;
}

.log-item {
  display: flex;
  align-items: baseline;
  padding: 3px 16px 3px 12px;
  color: #cdd6f4;
  white-space: nowrap;
  transition: background 0.15s;
  gap: 8px;
  border-left: 3px solid transparent;
}
.log-item:hover { background: rgba(255, 255, 255, 0.04); }
.log-item.log-tx { border-left-color: #40a02b; }
.log-item.log-rx { border-left-color: transparent; }

.log-item.log-fault {
  background: rgba(243, 139, 168, 0.08);
  border-left-color: #f38ba8;
}
.log-item.log-charging {
  background: rgba(166, 227, 161, 0.04);
}

/* 方向标签 */
.log-direction {
  min-width: 48px;
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
  text-align: center;
  padding: 0 4px;
  border-radius: 3px;
}
.dir-rx { color: #89b4fa; background: #89b4fa15; }
.dir-tx { color: #a6e3a1; background: #a6e3a115; }

.log-time { color: #6c7086; min-width: 150px; flex-shrink: 0; }
.log-type { min-width: 52px; flex-shrink: 0; font-weight: 600; }
.log-type.type-data { color: #89b4fa; }
.log-type.type-reply { color: #a6e3a1; }
.log-type.type-raw { color: #f9e2af; }
.log-type.type-sim { color: #cba6f7; }
.log-type.type-event { color: #a6e3a1; }

.log-values { display: flex; gap: 10px; flex-shrink: 0; }
.val-voltage { color: #f9e2af; }
.val-current { color: #89dceb; }
.val-power { color: #f5c2e7; }
.val-energy { color: #a6e3a1; }

.val-status {
  padding: 0 6px;
  border-radius: 3px;
  font-size: 12px;
}
.st-charging { background: #a6e3a120; color: #a6e3a1; }
.st-idle { background: #89b4fa20; color: #89b4fa; }
.st-finished { background: #f9e2af20; color: #f9e2af; }
.st-fault { background: #f38ba820; color: #f38ba8; }
.st-raw { background: #f9e2af20; color: #f9e2af; font-size: 11px; }
.st-reply { background: #a6e3a120; color: #a6e3a1; font-size: 11px; }

.val-raw-hex {
  color: #bac2de;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  word-break: break-all;
}
.hex-copy-btn {
  color: #7f849c;
  font-size: 11px;
  cursor: pointer;
  padding: 1px 4px;
  border-radius: 3px;
  transition: all 0.15s;
  border: 1px dashed transparent;
}
.hex-copy-btn:hover { color: #cba6f7; background: rgba(203, 166, 247, 0.1); border-color: rgba(203, 166, 247, 0.3); }
.val-fault { color: #f38ba8; font-weight: 600; }
.val-sim {
  color: #cba6f7;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 14px;
  letter-spacing: 2px;
}

.log-temp { color: #fab387; flex-shrink: 0; margin-left: auto; }

/* 滚动条 */
.log-list::-webkit-scrollbar { width: 6px; }
.log-list::-webkit-scrollbar-track { background: transparent; }
.log-list::-webkit-scrollbar-thumb { background: #45475a; border-radius: 3px; }
</style>
