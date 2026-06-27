<template>
  <div class="page-container">
    <div class="page-header">
      <el-button link @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon> 返回
      </el-button>
      <h2>设备详情 - {{ device?.sn }}</h2>
    </div>

    <el-row :gutter="20">
      <!-- 基本信息 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>基本信息</span>
            <el-tag :type="device?.is_online ? 'success' : 'info'" style="margin-left: 12px">
              {{ device?.is_online ? '在线' : '离线' }}
            </el-tag>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="设备SN">{{ device?.sn }}</el-descriptions-item>
            <el-descriptions-item label="设备类型">{{ device?.device_type === 'ev_charger' ? '汽车充电桩' : '电单车充电桩' }}</el-descriptions-item>
            <el-descriptions-item label="协议">{{ device?.protocol }}</el-descriptions-item>
            <el-descriptions-item label="设备厂家">{{ device?.manufacturer }}</el-descriptions-item>
            <el-descriptions-item label="设备型号">{{ device?.model }}</el-descriptions-item>
            <el-descriptions-item label="安装位置">{{ device?.install_location }}</el-descriptions-item>
            <el-descriptions-item label="固件版本">{{ device?.firmware_version }}</el-descriptions-item>
            <el-descriptions-item label="端口数">{{ device?.port_count }}</el-descriptions-item>
            <el-descriptions-item label="最后在线">{{ device?.last_online_at }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <!-- 实时数据 -->
      <el-col :span="12">
        <el-card>
          <template #header><span>实时数据</span></template>
          <el-row :gutter="16">
            <el-col :span="12" v-for="item in realtimeItems" :key="item.label" style="margin-bottom: 16px">
              <el-card shadow="never" class="realtime-card">
                <div class="realtime-label">{{ item.label }}</div>
                <div class="realtime-value" :style="{ color: item.color }">
                  {{ item.value }} <span class="realtime-unit">{{ item.unit }}</span>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 快速操作 -->
          <div style="margin-top: 20px">
            <el-button type="success" @click="handleSendCommand('start_charge')">启动充电</el-button>
            <el-button type="danger" @click="handleSendCommand('stop_charge')">停止充电</el-button>
            <el-button type="warning" @click="handleSendCommand('reboot')">远程重启</el-button>
            <el-button @click="handleSendCommand('config')">参数配置</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 指令历史 -->
    <el-card style="margin-top: 20px">
      <template #header><span>指令历史</span></template>
      <el-table :data="commands" size="small">
        <el-table-column prop="id" label="指令ID" width="280" />
        <el-table-column prop="cmd_type" label="指令类型" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="cmdStatusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="下发时间" width="170" />
        <el-table-column prop="responded_at" label="响应时间" width="170" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getDevice, sendCommand, type DeviceDetail } from '@/api/device'

const route = useRoute()
const deviceId = route.params.id as string

const device = ref<DeviceDetail | null>(null)
const loading = ref(true)

const realtimeItems = computed(() => {
  if (!device.value?.realtime_data) return []
  const d = device.value.realtime_data
  const fmt = (v: string, decimals: number) => {
    const n = parseFloat(v)
    return isNaN(n) ? v : n.toFixed(decimals)
  }
  const items: { label: string; value: string; unit: string; color: string }[] = []
  if (d.voltage) items.push({ label: '电压', value: fmt(d.voltage, 1), unit: 'V', color: '#409eff' })
  if (d.current) items.push({ label: '电流', value: fmt(d.current, 2), unit: 'A', color: '#67c23a' })
  if (d.power) items.push({ label: '功率', value: fmt(d.power, 2), unit: 'W', color: '#e6a23c' })
  if (d.temperature) items.push({ label: '温度', value: fmt(d.temperature, 1), unit: '℃', color: '#f56c6c' })
  if (d.energy_total) items.push({ label: '累计电量', value: fmt(d.energy_total, 2), unit: 'kWh', color: '#909399' })
  if (d.charging_status) items.push({ label: '充电状态', value: d.charging_status, unit: '', color: '#409eff' })
  return items
})

const commands = ref<any[]>([])

function statusType(status?: string) {
  const map: Record<string, string> = { online: 'success', offline: 'info', fault: 'danger' }
  return map[status || ''] || 'info'
}

function statusText(status?: string) {
  const map: Record<string, string> = { online: '在线', offline: '离线', fault: '故障' }
  return map[status || ''] || status || ''
}

function cmdStatusType(status: string) {
  const map: Record<string, string> = { success: 'success', failed: 'danger', pending: 'warning', sent: 'info' }
  return map[status] || 'info'
}

async function handleSendCommand(cmdType: string) {
  try {
    await sendCommand(deviceId, cmdType, {})
    ElMessage.success(`指令 [${cmdType}] 已下发至 ${device.value?.sn}`)
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || '指令下发失败'
    ElMessage.error(`指令下发失败: ${msg}`)
  }
}

async function fetchDetail() {
  loading.value = true
  try {
    const res = await getDevice(deviceId)
    device.value = res.data
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style scoped>
.realtime-card {
  text-align: center;
}
.realtime-label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
}
.realtime-value {
  font-size: 24px;
  font-weight: 700;
}
.realtime-unit {
  font-size: 14px;
  font-weight: 400;
  color: #909399;
}
</style>
