<template>
  <div class="page-container">
    <div class="page-header">
      <h2>设备高级管理</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 设备类型 -->
      <el-tab-pane label="设备类型" name="types">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openTypeDialog()">
            <el-icon><Plus /></el-icon> 新增设备类型
          </el-button>
        </div>
        <el-card>
          <el-table :data="deviceTypes" stripe>
            <el-table-column prop="name" label="类型名称" width="180" />
            <el-table-column prop="code" label="类型编码" width="150" />
            <el-table-column prop="protocol" label="默认协议" width="140" />
            <el-table-column prop="max_ports" label="最大端口数" width="110" align="center" />
            <el-table-column prop="power_type" label="功率类型" width="100">
              <template #default="{ row }">{{ { dc: '直流', ac: '交流', dc_ac: '交直流' }[row.power_type] || row.power_type }}</template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openTypeDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteType(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- OTA固件 -->
      <el-tab-pane label="OTA固件" name="ota">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openOTADialog()">
            <el-icon><Plus /></el-icon> 上传固件
          </el-button>
        </div>
        <el-card>
          <el-table :data="firmwares" stripe>
            <el-table-column prop="version" label="版本号" width="120" />
            <el-table-column prop="device_type" label="设备类型" width="140" />
            <el-table-column prop="file_name" label="文件名" width="240" show-overflow-tooltip />
            <el-table-column prop="file_size" label="文件大小" width="100" align="right">
              <template #default="{ row }">{{ ((row.file_size || 0) / 1024).toFixed(1) }} KB</template>
            </el-table-column>
            <el-table-column prop="md5" label="MD5" width="280" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'released' ? 'success' : row.status === 'testing' ? 'warning' : 'info'" size="small">
                  {{ { draft: '草稿', testing: '测试中', released: '已发布', deprecated: '已废弃' }[row.status] || row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button v-if="row.status === 'draft' || row.status === 'testing'" type="success" link size="small" @click="releaseOTA(row)">发布</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteOTA(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 二维码 -->
      <el-tab-pane label="设备二维码" name="qrcode">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openQRDialog()">
            <el-icon><Plus /></el-icon> 批量生成
          </el-button>
        </div>
        <el-card>
          <el-table :data="qrcodes" stripe>
            <el-table-column prop="device_sn" label="设备SN" width="180" />
            <el-table-column prop="qr_code" label="二维码内容" min-width="280" show-overflow-tooltip />
            <el-table-column prop="created_at" label="生成时间" width="170" />
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="viewQR(row)">查看</el-button>
                <el-button type="success" link size="small" @click="downloadQR(row)">下载</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 虚拟设备 -->
      <el-tab-pane label="虚拟设备" name="virtual">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openVDeviceDialog()">
            <el-icon><Plus /></el-icon> 新增虚拟设备
          </el-button>
        </div>
        <el-card>
          <el-table :data="virtualDevices" stripe>
            <el-table-column prop="sn" label="设备SN" width="180" />
            <el-table-column prop="device_type" label="设备类型" width="140" />
            <el-table-column prop="protocol" label="协议" width="120" />
            <el-table-column prop="port_count" label="端口数" width="100" align="center" />
            <el-table-column prop="data_interval" label="数据间隔(s)" width="120" align="center" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'running' ? 'success' : 'info'" size="small">
                  {{ row.status === 'running' ? '运行中' : '已停止' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="240">
              <template #default="{ row }">
                <el-button v-if="row.status !== 'running'" type="success" link size="small" @click="toggleVDevice(row, 'running')">启动</el-button>
                <el-button v-else type="warning" link size="small" @click="toggleVDevice(row, 'stopped')">停止</el-button>
                <el-button type="primary" link size="small" @click="openVDeviceDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteVDevice(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 设备类型弹窗 -->
    <el-dialog v-model="typeDialog.visible" :title="typeDialog.isEdit ? '编辑设备类型' : '新增设备类型'" width="500px">
      <el-form :model="typeDialog.form" label-width="100px">
        <el-form-item label="类型名称">
          <el-input v-model="typeDialog.form.name" placeholder="如：电单车充电桩" />
        </el-form-item>
        <el-form-item label="类型编码">
          <el-input v-model="typeDialog.form.code" placeholder="如：ebike_charger" :disabled="typeDialog.isEdit" />
        </el-form-item>
        <el-form-item label="默认协议">
          <el-input v-model="typeDialog.form.protocol" placeholder="如：AP3000_v2" />
        </el-form-item>
        <el-form-item label="最大端口数">
          <el-input-number v-model="typeDialog.form.max_ports" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="功率类型">
          <el-select v-model="typeDialog.form.power_type" style="width: 100%">
            <el-option label="直流(DC)" value="dc" />
            <el-option label="交流(AC)" value="ac" />
            <el-option label="交直流" value="dc_ac" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="typeDialog.form.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="typeDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveType">保存</el-button>
      </template>
    </el-dialog>

    <!-- OTA弹窗 -->
    <el-dialog v-model="otaDialog.visible" title="上传固件" width="500px">
      <el-form :model="otaDialog.form" label-width="100px">
        <el-form-item label="版本号">
          <el-input v-model="otaDialog.form.version" placeholder="如：v2.1.0" />
        </el-form-item>
        <el-form-item label="设备类型">
          <el-input v-model="otaDialog.form.device_type" placeholder="如：ebike_charger" />
        </el-form-item>
        <el-form-item label="固件文件">
          <el-upload :show-file-list="true" :auto-upload="false">
            <el-button>选择文件</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item label="版本说明">
          <el-input v-model="otaDialog.form.changelog" type="textarea" :rows="3" placeholder="固件更新内容说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="otaDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveOTA">上传</el-button>
      </template>
    </el-dialog>

    <!-- 二维码弹窗 -->
    <el-dialog v-model="qrDialog.visible" title="批量生成二维码" width="600px">
      <el-form label-width="100px">
        <el-form-item label="选择设备">
          <el-select v-model="qrDialog.deviceIds" multiple filterable placeholder="选择要生成二维码的设备" style="width: 100%">
            <el-option v-for="d in qrDialog.devices" :key="d.id" :label="`${d.sn} (${d.device_type})`" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="二维码类型">
          <el-radio-group v-model="qrDialog.qrType">
            <el-radio value="charge">充电二维码</el-radio>
            <el-radio value="info">设备信息二维码</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="qrDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doGenerateQR">生成</el-button>
      </template>
    </el-dialog>

    <!-- 虚拟设备弹窗 -->
    <el-dialog v-model="vdDialog.visible" :title="vdDialog.isEdit ? '编辑虚拟设备' : '新增虚拟设备'" width="500px">
      <el-form :model="vdDialog.form" label-width="100px">
        <el-form-item label="设备SN">
          <el-input v-model="vdDialog.form.sn" placeholder="虚拟设备标识" :disabled="vdDialog.isEdit" />
        </el-form-item>
        <el-form-item label="设备类型">
          <el-input v-model="vdDialog.form.device_type" placeholder="如：ebike_charger" />
        </el-form-item>
        <el-form-item label="协议">
          <el-input v-model="vdDialog.form.protocol" placeholder="如：AP3000_v2" />
        </el-form-item>
        <el-form-item label="端口数">
          <el-input-number v-model="vdDialog.form.port_count" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="数据间隔(s)">
          <el-input-number v-model="vdDialog.form.data_interval" :min="1" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="vdDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveVDevice">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const activeTab = ref('types')

interface DeviceOption { id: number; sn: string; device_type: string }

// 设备类型 (mock for now — backend API not implemented yet)
const deviceTypes = ref<any[]>([])
const typeDialog = reactive({ visible: false, isEdit: false, form: {} as any })

// OTA
const firmwares = ref<any[]>([])
const otaDialog = reactive({ visible: false, form: {} as any })

// 二维码
const qrcodes = ref<any[]>([])
const qrDialog = reactive({ visible: false, deviceIds: [] as number[], devices: [] as DeviceOption[], qrType: 'charge' })

// 虚拟设备
const virtualDevices = ref<any[]>([])
const vdDialog = reactive({ visible: false, isEdit: false, form: {} as any })

// 设备类型
function openTypeDialog(row?: any) {
  typeDialog.isEdit = !!row
  typeDialog.form = row ? { ...row } : { name: '', code: '', protocol: '', max_ports: 1, power_type: 'ac', description: '' }
  typeDialog.visible = true
}

function doSaveType() {
  if (typeDialog.isEdit) {
    const idx = deviceTypes.value.findIndex((t: any) => t.id === typeDialog.form.id)
    if (idx >= 0) deviceTypes.value[idx] = { ...typeDialog.form }
  } else {
    deviceTypes.value.push({ ...typeDialog.form, id: Date.now(), created_at: new Date().toISOString() })
  }
  ElMessage.success(typeDialog.isEdit ? '已更新' : '已创建')
  typeDialog.visible = false
}

function handleDeleteType(row: any) {
  ElMessageBox.confirm('确认删除该设备类型？', '删除确认', { type: 'warning' }).then(() => {
    deviceTypes.value = deviceTypes.value.filter((t: any) => t.id !== row.id)
    ElMessage.success('已删除')
  }).catch(() => {})
}

// OTA
function openOTADialog() {
  otaDialog.form = { version: '', device_type: '', file_name: '', changelog: '' }
  otaDialog.visible = true
}

async function doSaveOTA() {
  ElMessage.success('固件上传功能已触发')
  otaDialog.visible = false
}

async function releaseOTA(row: any) {
  try {
    ElMessage.success('固件已发布')
  } catch {}
}

function handleDeleteOTA(row: any) {
  ElMessageBox.confirm('确认删除该固件？', '删除确认', { type: 'warning' }).then(async () => {
    ElMessage.success('已删除')
  }).catch(() => {})
}

// 二维码
function openQRDialog() {
  qrDialog.deviceIds = []
  qrDialog.visible = true
}

async function doGenerateQR() {
  if (qrDialog.deviceIds.length === 0) {
    ElMessage.warning('请选择至少一个设备')
    return
  }
  ElMessage.success(`已为 ${qrDialog.deviceIds.length} 个设备生成二维码`)
  qrDialog.visible = false
}

function viewQR(row: any) {
  ElMessage.info('二维码预览：' + (row.qr_code || row.device_sn))
}

function downloadQR(row: any) {
  ElMessage.info('二维码下载：' + (row.qr_code || row.device_sn))
}

// 虚拟设备
function openVDeviceDialog(row?: any) {
  vdDialog.isEdit = !!row
  vdDialog.form = row ? { ...row } : { sn: '', device_type: '', protocol: '', port_count: 1, data_interval: 60 }
  vdDialog.visible = true
}

async function doSaveVDevice() {
  if (vdDialog.isEdit) {
    const idx = virtualDevices.value.findIndex((d: any) => d.id === vdDialog.form.id)
    if (idx >= 0) virtualDevices.value[idx] = { ...vdDialog.form }
  } else {
    virtualDevices.value.push({ ...vdDialog.form, id: Date.now(), status: 'stopped', created_at: new Date().toISOString() })
  }
  ElMessage.success(vdDialog.isEdit ? '已更新' : '已创建')
  vdDialog.visible = false
}

function toggleVDevice(row: any, status: string) {
  row.status = status
  ElMessage.success(status === 'running' ? '虚拟设备已启动' : '虚拟设备已停止')
}

function handleDeleteVDevice(row: any) {
  ElMessageBox.confirm('确认删除该虚拟设备？', '删除确认', { type: 'warning' }).then(() => {
    virtualDevices.value = virtualDevices.value.filter((d: any) => d.id !== row.id)
    ElMessage.success('已删除')
  }).catch(() => {})
}

onMounted(() => {
  // 初始化mock数据
  deviceTypes.value = [
    { id: 1, name: '电单车充电桩', code: 'ebike_charger', protocol: 'AP3000_v2', max_ports: 10, power_type: 'ac', description: '适用于电单车充电' },
    { id: 2, name: '汽车充电桩', code: 'ev_charger', protocol: 'WSD_v1', max_ports: 1, power_type: 'dc', description: '适用于电动汽车直流快充' },
  ]
})
</script>
