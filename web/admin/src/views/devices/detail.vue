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
            <el-tag :type="statusType(device?.status)" style="margin-left: 12px">
              {{ statusText(device?.status) }}
            </el-tag>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="设备SN">{{ device?.sn }}</el-descriptions-item>
            <el-descriptions-item label="设备类型">{{ device?.device_type === 'ev_charger' ? '汽车充电桩' : '电单车充电桩' }}</el-descriptions-item>
            <el-descriptions-item label="协议">{{ device?.protocol }}</el-descriptions-item>
            <el-descriptions-item label="厂商">{{ device?.vendor }}</el-descriptions-item>
            <el-descriptions-item label="设备型号">{{ device?.model }}</el-descriptions-item>
            <el-descriptions-item label="安装位置">{{ device?.install_location }}</el-descriptions-item>
            <el-descriptions-item label="固件版本">{{ device?.firmware_version }}</el-descriptions-item>
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
            <el-button type="success" @click="sendCommand('start_charge')">启动充电</el-button>
            <el-button type="danger" @click="sendCommand('stop_charge')">停止充电</el-button>
            <el-button type="warning" @click="sendCommand('reboot')">远程重启</el-button>
            <el-button @click="sendCommand('config')">参数配置</el-button>
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
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'

const route = useRoute()
const deviceId = route.params.id as string

const device = ref<any>({
  id: deviceId,
  sn: 'AP3000-001',
  device_type: 'ebike_charger',
  protocol: 'AP3000_v2',
  vendor: '安平科技',
  model: 'AP3000',
  install_location: 'A区1号棚',
  firmware_version: 'v2.1.0',
  status: 'online',
  last_online_at: '2026-06-19 00:45:00',
})

const realtimeItems = reactive([
  { label: '电压', value: '220.5', unit: 'V', color: '#409eff' },
  { label: '电流', value: '10.2', unit: 'A', color: '#67c23a' },
  { label: '功率', value: '2246', unit: 'W', color: '#e6a23c' },
  { label: '温度', value: '35.0', unit: '℃', color: '#f56c6c' },
  { label: '累计电量', value: '12345.6', unit: 'kWh', color: '#909399' },
  { label: '充电进度', value: '65', unit: '%', color: '#409eff' },
])

const commands = ref([
  { id: 'cmd-001', cmd_type: 'start_charge', status: 'success', created_at: '2026-06-19 00:30', responded_at: '2026-06-19 00:30' },
  { id: 'cmd-002', cmd_type: 'config', status: 'success', created_at: '2026-06-18 22:00', responded_at: '2026-06-18 22:00' },
])

function statusType(status: string) {
  const map: Record<string, string> = { online: 'success', offline: 'info', fault: 'danger' }
  return map[status] || 'info'
}

function statusText(status: string) {
  const map: Record<string, string> = { online: '在线', offline: '离线', fault: '故障' }
  return map[status] || status
}

function cmdStatusType(status: string) {
  const map: Record<string, string> = { success: 'success', failed: 'danger', pending: 'warning', sent: 'info' }
  return map[status] || 'info'
}

function sendCommand(cmdType: string) {
  ElMessage.success(`指令 [${cmdType}] 已下发至 ${device.value?.sn}`)
}

onMounted(() => {
  // fetch device detail
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
