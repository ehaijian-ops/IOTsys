<template>
  <div class="page-container">
    <div class="page-header">
      <h2>未注册设备</h2>
      <el-button type="primary" @click="fetchData" :loading="loading">
        <el-icon><Refresh /></el-icon> 刷新
      </el-button>
    </div>

    <el-alert
      title="提示"
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 16px"
    >
      以下设备已向服务器发起注册但尚未添加到任何站点，请及时处理。
    </el-alert>

    <el-card>
      <el-table :data="devices" v-loading="loading" stripe empty-text="暂无未注册设备">
        <el-table-column prop="device_id" label="设备ID" width="180" />
        <el-table-column prop="protocol" label="协议" width="130">
          <template #default="{ row }">
            <el-tag :type="protocolTag(row.protocol)" size="small">
              {{ row.protocol || '未知' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sim_card_number" label="SIM卡号" width="200">
          <template #default="{ row }">
            <span v-if="row.sim_card_number">{{ row.sim_card_number }}</span>
            <el-tag v-else type="info" size="small">未获取</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remote_addr" label="远端地址" width="160" />
        <el-table-column prop="connected_at" label="连接时间" width="170">
          <template #default="{ row }">
            <el-tooltip :content="'最后活跃: ' + row.last_active">
              <span>{{ row.connected_at }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="last_active" label="最后活跃" width="170">
          <template #default="{ row }">
            <span :class="{ 'text-warning': isStale(row.last_active) }">
              {{ row.last_active }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openAddDialog(row)">
              <el-icon><Plus /></el-icon> 添加到站点
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="devices.length > 0" style="margin-top: 12px; text-align: right; color: #909399">
        共 {{ devices.length }} 台未注册设备
      </div>
    </el-card>

    <!-- 添加设备到站点对话框 -->
    <el-dialog v-model="dialogVisible" title="添加设备到站点" width="520px" :close-on-click-modal="false">
      <el-form :model="form" label-width="100px">
        <el-form-item label="设备ID">
          <el-input :model-value="currentDevice?.device_id" disabled />
        </el-form-item>
        <el-form-item label="协议类型">
          <el-tag :type="protocolTag(currentDevice?.protocol || '')" size="small">
            {{ currentDevice?.protocol || '未知' }}
          </el-tag>
        </el-form-item>
        <el-form-item v-if="currentDevice?.sim_card_number" label="SIM卡号">
          <span>{{ currentDevice?.sim_card_number }}</span>
        </el-form-item>
        <el-form-item label="设备类型" required>
          <el-select v-model="form.device_type" style="width: 100%">
            <el-option label="电单车充电桩" value="ebike_charger" />
            <el-option label="电动汽车充电桩" value="ev_charger" />
          </el-select>
        </el-form-item>
        <el-form-item label="归属站点" required>
          <el-select v-model="form.site_id" placeholder="请选择站点" style="width: 100%" :loading="sitesLoading">
            <el-option
              v-for="site in sites"
              :key="site.id"
              :label="site.name"
              :value="site.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="安装位置">
          <el-input v-model="form.install_location" placeholder="如：A区1号棚" />
        </el-form-item>
        <el-form-item label="端口数量" required>
          <el-input-number
            v-model="form.port_count"
            :min="1"
            :max="MAX_PORT_COUNT"
            :step="1"
            style="width: 100%"
            placeholder="请输入设备端口数量"
          />
          <template #extra>
            <span style="font-size: 12px; color: #909399">正整数，最大不超过 {{ MAX_PORT_COUNT }} 个端口</span>
          </template>
        </el-form-item>
        <el-form-item label="设备厂家">
          <el-input v-model="form.manufacturer" placeholder="请输入设备制造商" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAdd" :loading="submitting">确认添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getUnregisteredDevices,
  addUnregisteredToSite,
  type UnregisteredDevice,
} from '@/api/device'
import {
  getSitesBrief,
  type SiteBrief,
} from '@/api/site'

const loading = ref(false)
const submitting = ref(false)
const devices = ref<UnregisteredDevice[]>([])
const dialogVisible = ref(false)
const currentDevice = ref<UnregisteredDevice | null>(null)
const sites = ref<SiteBrief[]>([])
const sitesLoading = ref(false)

const form = reactive({
  device_type: 'ebike_charger',
  site_id: '',
  install_location: '',
  port_count: 1,
  manufacturer: '',
})

const MAX_PORT_COUNT = 100

onMounted(() => {
  fetchData()
})

async function fetchData() {
  loading.value = true
  try {
    const res = await getUnregisteredDevices()
    devices.value = res.data || []
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

async function openAddDialog(device: UnregisteredDevice) {
  currentDevice.value = device
  form.device_type = 'ebike_charger'
  form.site_id = ''
  form.install_location = ''
  form.manufacturer = ''
  // 优先使用设备上报的端口数量，未上报时回退默认值 1
  form.port_count = (device.port_count && device.port_count > 0) ? device.port_count : 1

  // 加载站点列表
  sitesLoading.value = true
  try {
    const res = await getSitesBrief()
    sites.value = res.data || []
  } catch {
    sites.value = []
  } finally {
    sitesLoading.value = false
  }

  dialogVisible.value = true
}

async function handleAdd() {
  if (!form.site_id) {
    ElMessage.warning('请选择归属站点')
    return
  }
  if (!currentDevice.value) return

  // 端口数量校验
  const portCount = form.port_count
  if (!Number.isInteger(portCount) || portCount <= 0) {
    ElMessage.warning('请输入有效的端口数量（正整数）')
    return
  }
  if (portCount > MAX_PORT_COUNT) {
    ElMessage.warning(`端口数量不能超过 ${MAX_PORT_COUNT}`)
    return
  }

  submitting.value = true
  try {
    await addUnregisteredToSite({
      device_id: currentDevice.value.device_id,
      protocol: currentDevice.value.protocol,
      device_type: form.device_type,
      site_id: form.site_id,
      install_location: form.install_location || undefined,
      port_count: portCount,
      manufacturer: form.manufacturer || undefined,
    })
    ElMessage.success('设备已成功添加到站点')
    dialogVisible.value = false
    fetchData()
  } catch {
    // error handled by interceptor
  } finally {
    submitting.value = false
  }
}

function protocolTag(protocol: string) {
  const map: Record<string, string> = {
    'AP3000_v2': '',
    'WSD_v1': 'success',
    'TF100_v1': 'warning',
  }
  return map[protocol] || 'info'
}

function isStale(lastActive: string) {
  if (!lastActive) return false
  try {
    const t = new Date(lastActive.replace(' ', 'T')).getTime()
    return Date.now() - t > 5 * 60 * 1000 // 5分钟无活跃视为不活跃
  } catch {
    return false
  }
}
</script>

<style scoped>
.text-warning {
  color: #e6a23c;
  font-weight: 500;
}
</style>
