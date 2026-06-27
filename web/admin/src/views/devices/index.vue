<template>
  <div class="page-container">
    <div class="page-header">
      <h2>设备管理</h2>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon> 新增设备
      </el-button>
    </div>

    <!-- 搜索筛选 -->
    <el-card style="margin-bottom: 16px">
      <el-form :inline="true" :model="query" size="default">
        <el-form-item label="设备类型">
          <el-select v-model="query.device_type" placeholder="全部" clearable style="width: 150px">
            <el-option label="电单车充电桩" value="ebike_charger" />
            <el-option label="汽车充电桩" value="ev_charger" />
          </el-select>
        </el-form-item>
        <el-form-item label="协议">
          <el-select v-model="query.protocol" placeholder="全部" clearable style="width: 150px">
            <el-option label="AP3000_v2" value="AP3000_v2" />
            <el-option label="WSD_v1" value="WSD_v1" />
            <el-option label="TF100_v1" value="TF100_v1" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="故障" value="fault" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="query.keyword" placeholder="SN/设备厂家" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchDevices">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 设备列表 -->
    <el-card>
      <el-table :data="devices" v-loading="loading" stripe>
        <el-table-column prop="sn" label="设备SN" width="180" />
        <el-table-column prop="device_type" label="设备类型" width="130">
          <template #default="{ row }">
            <el-tag :type="row.device_type === 'ev_charger' ? 'success' : 'warning'" size="small">
              {{ row.device_type === 'ev_charger' ? '汽车充电桩' : '电单车充电桩' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="protocol" label="协议" width="120" />
        <el-table-column prop="manufacturer" label="设备厂家" width="120" />
        <el-table-column prop="port_count" label="端口数" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="firmware_version" label="固件版本" width="100" />
        <el-table-column prop="install_location" label="安装位置" min-width="150" />
        <el-table-column prop="last_online_at" label="最后在线" width="170" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row.id)">详情</el-button>
            <el-button link type="primary" @click="openCommandDialog(row)">指令</el-button>
            <el-button link type="warning" @click="openEditDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; text-align: right">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @change="fetchDevices"
        />
      </div>
    </el-card>

    <!-- 新增/编辑设备对话框 -->
    <el-dialog v-model="showCreateDialog" :title="currentDevice ? '编辑设备' : '新增设备'" width="550px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="设备SN" required>
          <el-input v-model="form.sn" placeholder="请输入设备序列号" />
        </el-form-item>
        <el-form-item label="设备类型" required>
          <el-select v-model="form.device_type" style="width: 100%">
            <el-option label="电单车充电桩" value="ebike_charger" />
            <el-option label="电动汽车充电桩" value="ev_charger" />
          </el-select>
        </el-form-item>
        <el-form-item label="协议" required>
          <el-select v-model="form.protocol" style="width: 100%">
            <el-option label="AP3000_v2 (电单车-安平)" value="AP3000_v2" />
            <el-option label="WSD_v1 (电单车-微小电)" value="WSD_v1" />
            <el-option label="TF100_v1 (汽车-特来电)" value="TF100_v1" />
          </el-select>
        </el-form-item>
        <el-form-item label="设备厂家">
          <el-input v-model="form.manufacturer" placeholder="设备制造商" />
        </el-form-item>
        <el-form-item label="设备型号">
          <el-input v-model="form.model" />
        </el-form-item>
        <el-form-item label="安装位置">
          <el-input v-model="form.install_location" />
        </el-form-item>
        <el-form-item label="固件版本">
          <el-input v-model="form.firmware_version" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">确定</el-button>
      </template>
    </el-dialog>

    <!-- 指令下发对话框 -->
    <el-dialog v-model="showCommandDialog" title="下发指令" width="450px">
      <el-form :model="commandForm" label-width="100px">
        <el-form-item label="目标设备">
          <span>{{ currentDevice?.sn }}</span>
        </el-form-item>
        <el-form-item label="指令类型" required>
          <el-select v-model="commandForm.cmdType" style="width: 100%">
            <el-option label="启动充电" value="start_charge" />
            <el-option label="停止充电" value="stop_charge" />
            <el-option label="参数配置" value="config" />
            <el-option label="远程重启" value="reboot" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="commandForm.cmdType === 'start_charge'" label="充电端口">
          <el-input-number v-model="commandForm.params.port" :min="1" :max="10" />
        </el-form-item>
        <el-form-item v-if="commandForm.cmdType === 'start_charge'" label="充电时长(分)">
          <el-input-number v-model="commandForm.params.duration" :min="0" placeholder="0=不限时" />
        </el-form-item>
        <el-form-item v-if="commandForm.cmdType === 'config'" label="最大功率(W)">
          <el-input-number v-model="commandForm.params.max_power" :min="0" :max="7000" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCommandDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSendCommand">下发</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDevices, createDevice, updateDevice, deleteDevice, type Device, type DeviceQuery } from '@/api/device'

const router = useRouter()
const loading = ref(false)
const devices = ref<Device[]>([])
const total = ref(0)

const query = reactive<DeviceQuery>({
  page: 1,
  page_size: 20,
})

const showCreateDialog = ref(false)
const showCommandDialog = ref(false)
const currentDevice = ref<Device | null>(null)

const form = reactive({
  sn: '',
  device_type: 'ebike_charger',
  protocol: 'AP3000_v2',
  manufacturer: '',
  model: '',
  install_location: '',
  firmware_version: '',
})

const commandForm = reactive({
  cmdType: 'start_charge',
  params: {} as Record<string, any>,
})

async function fetchDevices() {
  loading.value = true
  try {
    const res = await getDevices(query)
    devices.value = res.data || []
    total.value = res.total || 0
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.device_type = undefined
  query.protocol = undefined
  query.status = undefined
  query.keyword = undefined
  query.page = 1
  fetchDevices()
}

function statusType(status: string) {
  const map: Record<string, string> = {
    online: 'success',
    offline: 'info',
    fault: 'danger',
    charging: '',
    maintenance: 'warning',
  }
  return map[status] || 'info'
}

function statusText(status: string) {
  const map: Record<string, string> = {
    online: '在线',
    offline: '离线',
    fault: '故障',
    charging: '充电中',
    maintenance: '维护中',
  }
  return map[status] || status
}

function viewDetail(id: string) {
  router.push(`/devices/${id}`)
}

function openCommandDialog(device: Device) {
  currentDevice.value = device
  showCommandDialog.value = true
}

function openCreateDialog() {
  currentDevice.value = null
  form.sn = ''
  form.device_type = 'ebike_charger'
  form.protocol = 'AP3000_v2'
  form.manufacturer = ''
  form.model = ''
  form.install_location = ''
  form.firmware_version = ''
  showCreateDialog.value = true
}

function openEditDialog(device: Device) {
  currentDevice.value = device
  form.sn = device.sn
  form.device_type = device.device_type
  form.protocol = device.protocol
  form.manufacturer = device.manufacturer || ''
  form.model = device.model || ''
  form.install_location = device.install_location || ''
  form.firmware_version = device.firmware_version || ''
  showCreateDialog.value = true
}

async function handleCreate() {
  loading.value = true
  try {
    if (currentDevice.value) {
      // 编辑模式
      await updateDevice(currentDevice.value.id, {
        device_type: form.device_type,
        protocol: form.protocol,
        manufacturer: form.manufacturer,
        model: form.model,
        install_location: form.install_location,
        firmware_version: form.firmware_version,
      } as any)
      ElMessage.success('设备更新成功')
    } else {
      // 新增模式
      await createDevice(form as any)
      ElMessage.success('设备创建成功')
    }
    showCreateDialog.value = false
    currentDevice.value = null
    fetchDevices()
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

async function handleDelete(device: Device) {
  try {
    await ElMessageBox.confirm(`确定要删除设备 ${device.sn} 吗？`, '确认删除', {
      type: 'warning',
    })
    await deleteDevice(device.id)
    ElMessage.success('设备已删除')
    fetchDevices()
  } catch {
    // cancelled or error
  }
}

function handleSendCommand() {
  ElMessage.success(`指令 [${commandForm.cmdType}] 已下发`)
  showCommandDialog.value = false
  commandForm.params = {}
}

onMounted(() => {
  fetchDevices()
})
</script>
