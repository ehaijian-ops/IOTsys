<template>
  <div class="page-container">
    <div class="page-header">
      <h2>设备管理</h2>
      <el-button type="primary" @click="showCreateDialog = true">
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
          <el-input v-model="query.keyword" placeholder="SN/厂商" clearable style="width: 200px" />
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
        <el-table-column prop="vendor" label="厂商" width="120" />
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
    <el-dialog v-model="showCreateDialog" title="新增设备" width="550px">
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
        <el-form-item label="厂商">
          <el-input v-model="form.vendor" />
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
import { getDevices, type Device, type DeviceQuery } from '@/api/device'

const router = useRouter()
const loading = ref(false)
const devices = ref<Device[]>([])
const total = ref(0)

const query = reactive<DeviceQuery>({
  page: 1,
  page_size: 20,
})

// 模拟数据
const mockDevices: Device[] = [
  { id: '1', sn: 'AP3000-001', device_type: 'ebike_charger', protocol: 'AP3000_v2', vendor: '安平科技', model: 'AP3000', site_id: '', install_location: 'A区1号棚', firmware_version: 'v2.1.0', status: 'online', last_online_at: '2026-06-19 00:45:00', created_at: '', updated_at: '' },
  { id: '2', sn: 'AP3000-002', device_type: 'ebike_charger', protocol: 'AP3000_v2', vendor: '安平科技', model: 'AP3000', site_id: '', install_location: 'A区2号棚', firmware_version: 'v2.1.0', status: 'charging', last_online_at: '2026-06-19 00:44:00', created_at: '', updated_at: '' },
  { id: '3', sn: 'TF100-001', device_type: 'ev_charger', protocol: 'TF100_v1', vendor: '特来电', model: 'TF100', site_id: '', install_location: 'B1停车位', firmware_version: 'v1.5.0', status: 'online', last_online_at: '2026-06-19 00:45:00', created_at: '', updated_at: '' },
  { id: '4', sn: 'TF100-002', device_type: 'ev_charger', protocol: 'TF100_v1', vendor: '特来电', model: 'TF100', site_id: '', install_location: 'B2停车位', firmware_version: 'v1.5.0', status: 'offline', last_online_at: '2026-06-18 23:30:00', created_at: '', updated_at: '' },
  { id: '5', sn: 'AP3000-005', device_type: 'ebike_charger', protocol: 'AP3000_v2', vendor: '安平科技', model: 'AP3000', site_id: '', install_location: 'C区1号棚', firmware_version: 'v2.1.0', status: 'fault', last_online_at: '2026-06-19 00:40:00', created_at: '', updated_at: '' },
  { id: '6', sn: 'WSD-001', device_type: 'ebike_charger', protocol: 'WSD_v1', vendor: '微小电', model: 'WSD-12', site_id: '', install_location: 'D区1号棚', firmware_version: 'V1.0', status: 'online', last_online_at: '2026-06-19 00:45:00', created_at: '', updated_at: '' },
  { id: '7', sn: 'WSD-002', device_type: 'ebike_charger', protocol: 'WSD_v1', vendor: '微小电', model: 'WSD-12', site_id: '', install_location: 'D区2号棚', firmware_version: 'V1.0', status: 'charging', last_online_at: '2026-06-19 00:44:00', created_at: '', updated_at: '' },
]

const showCreateDialog = ref(false)
const showCommandDialog = ref(false)
const currentDevice = ref<Device | null>(null)

const form = reactive({
  sn: '',
  device_type: 'ebike_charger',
  protocol: 'AP3000_v2',
  vendor: '',
  model: '',
  install_location: '',
  firmware_version: '',
})

const commandForm = reactive({
  cmdType: 'start_charge',
  params: {} as Record<string, any>,
})

function fetchDevices() {
  loading.value = true
  // 使用模拟数据
  setTimeout(() => {
    devices.value = mockDevices
    total.value = mockDevices.length
    loading.value = false
  }, 500)
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

function openEditDialog(device: Device) {
  Object.assign(form, device)
  showCreateDialog.value = true
}

function handleCreate() {
  ElMessage.success('设备创建成功')
  showCreateDialog.value = false
  fetchDevices()
}

function handleDelete(device: Device) {
  ElMessageBox.confirm(`确定要删除设备 ${device.sn} 吗？`, '确认删除', {
    type: 'warning',
  }).then(() => {
    ElMessage.success('设备已删除')
    fetchDevices()
  })
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
