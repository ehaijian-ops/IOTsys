<template>
  <div class="page-container">
    <div class="page-header">
      <h2>系统监控</h2>
      <div class="header-actions">
        <el-tag :type="autoRefresh ? 'success' : 'info'" size="small">
          {{ autoRefresh ? '自动刷新 10s' : '手动刷新' }}
        </el-tag>
        <el-button type="primary" size="small" :loading="loading" @click="fetchStatus">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button size="small" @click="autoRefresh = !autoRefresh">
          {{ autoRefresh ? '停止自动' : '自动刷新' }}
        </el-button>
      </div>
    </div>

    <!-- 服务运行信息 -->
    <el-row :gutter="16" class="section">
      <el-col :span="6">
        <el-card class="info-card" shadow="hover">
          <div class="card-metric">
            <div class="metric-icon" style="background: linear-gradient(135deg, #409eff, #66b1ff)">
              <el-icon :size="22"><Clock /></el-icon>
            </div>
            <div class="metric-body">
              <div class="metric-value">{{ status?.server.uptime || '-' }}</div>
              <div class="metric-label">运行时长</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="info-card" shadow="hover">
          <div class="card-metric">
            <div class="metric-icon" style="background: linear-gradient(135deg, #67c23a, #85ce61)">
              <el-icon :size="22"><Connection /></el-icon>
            </div>
            <div class="metric-body">
              <div class="metric-value">{{ status?.resources.goroutines ?? '-' }}</div>
              <div class="metric-label">Goroutines</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="info-card" shadow="hover">
          <div class="card-metric">
            <div class="metric-icon" style="background: linear-gradient(135deg, #e6a23c, #ebb563)">
              <el-icon :size="22"><Cpu /></el-icon>
            </div>
            <div class="metric-body">
              <div class="metric-value">{{ status?.resources.heap_alloc_mb ?? '-' }} MB</div>
              <div class="metric-label">内存占用</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="info-card" shadow="hover">
          <div class="card-metric">
            <div class="metric-icon" style="background: linear-gradient(135deg, #f56c6c, #f89898)">
              <el-icon :size="22"><Monitor /></el-icon>
            </div>
            <div class="metric-body">
              <div class="metric-value">{{ status?.resources.tcp_connections ?? 0 }} / {{ status?.resources.sse_clients ?? 0 }}</div>
              <div class="metric-label">TCP / SSE 连接</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 服务状态 -->
    <el-row :gutter="16" class="section">
      <el-col :span="24">
        <el-card class="section-card" shadow="never">
          <template #header>
            <span class="card-title">关键服务状态</span>
          </template>
          <el-row :gutter="16">
            <el-col v-for="svc in serviceList" :key="svc.key" :span="8" style="margin-bottom: 16px">
              <div class="service-item" :class="getStatusClass(svc.data?.status)">
                <div class="service-status-dot"></div>
                <div class="service-info">
                  <div class="service-name">{{ svc.label }}</div>
                  <div class="service-status-text">
                    {{ getStatusText(svc.data?.status) }}
                  </div>
                  <div class="service-details" v-if="svc.data?.details">
                    {{ svc.data.details }}
                  </div>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <!-- 服务端信息 & 连接统计 -->
    <el-row :gutter="16" class="section">
      <el-col :span="12">
        <el-card class="section-card" shadow="never">
          <template #header>
            <span class="card-title">服务器信息</span>
          </template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="服务名称">{{ status?.server.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="运行环境">{{ status?.server.env || '-' }}</el-descriptions-item>
            <el-descriptions-item label="监听端口">HTTP: {{ status?.server.port || '-' }} / TCP: 7000</el-descriptions-item>
            <el-descriptions-item label="CPU 核数">{{ status?.resources.num_cpu ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="启动时间">{{ status?.server.start_time || '-' }}</el-descriptions-item>
            <el-descriptions-item label="检测时间">
              <span class="checked-time">{{ status?.checked_at || '-' }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card class="section-card" shadow="never">
          <template #header>
            <span class="card-title">协议适配器 & 连接统计</span>
          </template>
          <div class="adapter-list">
            <div class="adapter-item">
              <el-tag type="success" size="small">AP3000</el-tag>
              <span class="adapter-name">电单车充电桩 v8.6</span>
              <span class="adapter-type">二进制帧 (DNY)</span>
            </div>
            <div class="adapter-item">
              <el-tag type="success" size="small">WSD</el-tag>
              <span class="adapter-name">电单车充电桩(微小电) v1.0</span>
              <span class="adapter-type">二进制帧 (0xEE)</span>
            </div>
            <div class="adapter-item">
              <el-tag type="success" size="small">TF100</el-tag>
              <span class="adapter-name">汽车充电桩 v1.0</span>
              <span class="adapter-type">JSON 帧 (CCMD/SCMD)</span>
            </div>
          </div>
          <el-divider />
          <el-row :gutter="12">
            <el-col :span="12">
              <div class="conn-stat">
                <div class="conn-stat-value">{{ status?.resources.tcp_connections ?? 0 }}</div>
                <div class="conn-stat-label">TCP 设备连接</div>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="conn-stat">
                <div class="conn-stat-value">{{ status?.resources.sse_clients ?? 0 }}</div>
                <div class="conn-stat-label">SSE 订阅客户端</div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { getSystemStatus, type SystemStatus, type ServiceStatus } from '@/api/system'
import { ElMessage } from 'element-plus'

const status = ref<SystemStatus | null>(null)
const loading = ref(false)
const autoRefresh = ref(true)
const errorCount = ref(0)
let timer: number | null = null

const serviceList = computed(() => {
  if (!status.value) return []
  const svc = status.value.services
  return [
    { key: 'mysql',     label: 'MySQL',       data: svc.mysql },
    { key: 'redis',     label: 'Redis',       data: svc.redis },
    { key: 'mongodb',   label: 'MongoDB',     data: svc.mongodb },
    { key: 'kafka',     label: 'Kafka',       data: svc.kafka },
    { key: 'tcp_server', label: 'TCP 服务器',  data: svc.tcp_server },
    { key: 'sse_hub',   label: 'SSE Hub',     data: svc.sse_hub },
  ]
})

function getStatusClass(s: string | undefined): string {
  switch (s) {
    case 'running':  return 'status-running'
    case 'stopped':
    case 'error':    return 'status-error'
    case 'disabled': return 'status-disabled'
    default:         return 'status-unknown'
  }
}

function getStatusText(s: string | undefined): string {
  const map: Record<string, string> = {
    running:  '运行中',
    stopped:  '已停止',
    error:    '异常',
    disabled: '未启用',
  }
  return map[s || ''] || '未知'
}

async function fetchStatus() {
  loading.value = true
  try {
    const res = await getSystemStatus()
    if (res && (res as any).data) {
      status.value = (res as any).data
    } else {
      status.value = res as any
    }
    errorCount.value = 0
  } catch (e) {
    errorCount.value++
    if (errorCount.value <= 2) {
      ElMessage.warning('获取系统状态失败')
    }
  } finally {
    loading.value = false
  }
}

watch(autoRefresh, (val) => {
  if (val) {
    startTimer()
  } else {
    stopTimer()
  }
})

function startTimer() {
  stopTimer()
  timer = window.setInterval(fetchStatus, 10000)
}

function stopTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

onMounted(() => {
  fetchStatus()
  if (autoRefresh.value) {
    startTimer()
  }
})

onUnmounted(() => {
  stopTimer()
})
</script>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section {
  margin-bottom: 16px;
}

.info-card {
  cursor: default;
}

.card-metric {
  display: flex;
  align-items: center;
  gap: 14px;
}

.metric-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.metric-body {
  flex: 1;
  min-width: 0;
}

.metric-value {
  font-size: 22px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}

.metric-label {
  font-size: 13px;
  color: #909399;
  margin-top: 2px;
}

.section-card {
  height: 100%;
}

.card-title {
  font-weight: 600;
  font-size: 15px;
}

/* 服务状态项 */
.service-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 8px;
  background: #fafafa;
  border: 1px solid #ebeef5;
  transition: all 0.2s;
}

.service-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.service-status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  margin-top: 5px;
  flex-shrink: 0;
  background: #c0c4cc;
}

.status-running .service-status-dot {
  background: #67c23a;
  box-shadow: 0 0 6px rgba(103, 194, 58, 0.5);
}

.status-error .service-status-dot {
  background: #f56c6c;
  box-shadow: 0 0 6px rgba(245, 108, 108, 0.5);
  animation: pulse 1.5s infinite;
}

.status-disabled .service-status-dot {
  background: #e6a23c;
}

.status-unknown .service-status-dot {
  background: #c0c4cc;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.service-info {
  flex: 1;
  min-width: 0;
}

.service-name {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 2px;
}

.service-status-text {
  font-size: 13px;
  margin-bottom: 2px;
}

.status-running .service-status-text { color: #67c23a; }
.status-error .service-status-text { color: #f56c6c; }
.status-disabled .service-status-text { color: #e6a23c; }

.service-details {
  font-size: 12px;
  color: #909399;
  word-break: break-all;
  line-height: 1.4;
}

/* 适配器列表 */
.adapter-list {
  margin-bottom: 8px;
}

.adapter-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  border-bottom: 1px solid #f2f3f5;
}

.adapter-item:last-child {
  border-bottom: none;
}

.adapter-name {
  font-size: 13px;
  font-weight: 500;
}

.adapter-type {
  font-size: 12px;
  color: #909399;
  margin-left: auto;
}

/* 连接统计 */
.conn-stat {
  text-align: center;
  padding: 16px 0;
}

.conn-stat-value {
  font-size: 32px;
  font-weight: 700;
  color: #409eff;
}

.conn-stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

.checked-time {
  color: #909399;
  font-size: 12px;
}
</style>
